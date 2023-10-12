package post

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"redditclone/pkg/comment"
	"redditclone/pkg/myerrors"
	"redditclone/pkg/other"
	"redditclone/pkg/user"
	"redditclone/pkg/votes"
	"sync"
	"time"

	mgo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"gopkg.in/mgo.v2/bson"
)

type Post struct {
	Score            int                `json:"score" bson:"score"`
	Views            uint32             `json:"views" bson:"views"`
	Type             string             `json:"type" bson:"type"`
	Title            string             `json:"title" bson:"title"`
	Url              string             `json:"url,omitempty" bson:"url,omitempty"`
	Text             string             `json:"text,omitempty" bson:"text,omitempty"`
	Author           user.User          `json:"author" bson:"author"`
	Category         string             `json:"category" bson:"category"`
	Votes            []*votes.Vote      `json:"votes" bson:"votes"`
	Comments         []*comment.Comment `json:"comments" bson:"comments"`
	Created          string             `json:"created" bson:"created"`
	UpvotePercentage int                `json:"upvotePercentage" bson:"upvotePercentage"`
	Id               bson.ObjectId      `json:"id" bson:"_id"`
}

type PostsInterface interface {
	Init() error
	Create(r *http.Request, usr user.User) ([]byte, error)
	Delete(id string) error
	GetOne(id string, viewsAdd uint32) ([]byte, error)
	GetAll() ([]byte, error)
	FindByCategory(category string) []*Post
	FindByUsername(usr user.UserInterface) []*Post
	CreateComment(r *http.Request, usr user.User, postId string) error
	DeleteComment(postId, commentId string) error
	MakeVote(postId, userId string, vote int) error
	Destroy()
}

type Posts struct {
	All CollectionHelper
	cl  *mgo.Client
	mu  *sync.Mutex
	ctx context.Context
}

func (p *Posts) Init() error {
	client, err := mgo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return fmt.Errorf("posts init : %w", err)
	}
	err = client.Connect(p.ctx)
	if err != nil {
		return fmt.Errorf("posts init : %w", err)
	}
	err = client.Ping(p.ctx, readpref.Primary())
	if err != nil {
		return fmt.Errorf("posts init : %w", err)
	}
	db := client.Database("coursera")
	p.All = MgoLayer{
		c: db.Collection("posts"),
	}
	p.cl = client
	p.mu = &sync.Mutex{}
	return nil
}

func (p Posts) Destroy() {
	p.cl.Disconnect(p.ctx)
}

func (p Posts) FindByCategory(category string) []*Post {
	res := make([]*Post, 0)
	cursor, err := p.All.Find(p.ctx, bson.M{"category": category})
	if err != nil {
		return nil
	}
	err = cursor.All(p.ctx, &res)
	if err != nil {
		return nil
	}
	return res
}

func (p Posts) FindByUsername(usr user.UserInterface) []*Post {
	res := make([]*Post, 0)
	cursor, err := p.All.Find(
		p.ctx,
		bson.M{
			"author": bson.M{
				"id":       usr.ID(),
				"username": usr.Login(),
			},
		})
	if err != nil {
		return nil
	}

	err = cursor.All(p.ctx, &res)
	if err != nil {
		return nil
	}
	return res
}

func (p Posts) Create(r *http.Request, usr user.User) ([]byte, error) {
	m, err := other.GetMapFromPost(r)
	r.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("create post : %w", err)
	}
	typee := m["type"]
	category := m["category"]
	title := m["title"]
	url := m["url"]
	text := m["text"]
	if typee == "" || (typee != "text" && typee != "link") || (url == "" && text == "") || category == "" || title == "" {
		return nil, fmt.Errorf("%w : create post : wrong parametrs", myerrors.ErrBadRequest)
	}

	post := &Post{
		Category:         category,
		Title:            title,
		Url:              url,
		Text:             text,
		Type:             typee,
		Created:          time.Now().Local().Format(time.RFC3339),
		Views:            0,
		Score:            0,
		Author:           usr,
		UpvotePercentage: 0,
		Id:               bson.NewObjectId(),
		Comments:         make([]*comment.Comment, 0),
		Votes:            make([]*votes.Vote, 0),
	}
	str, err := json.Marshal(post)
	if err != nil {
		return nil, fmt.Errorf("%w : create post : cant murshal to json", myerrors.ErrInternalError)
	}
	_, err = p.All.InsertOne(p.ctx, post)
	if err != nil {
		return nil, fmt.Errorf("%w : create post : %s", myerrors.ErrInternalError, err)
	}
	return str, nil
}

func (p Posts) Delete(id string) error {
	if !bson.IsObjectIdHex(id) {
		return fmt.Errorf("%w : Delete post : wrong post id", myerrors.ErrBadRequest)
	}
	bid := bson.ObjectIdHex(id)

	res, err := p.All.DeleteOne(p.ctx, bson.M{"_id": bid})
	if err != nil {
		return fmt.Errorf("%w : delete post : %s", myerrors.ErrInternalError, err)
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("%w : no post with id : %s", myerrors.ErrComplex{Msg: "post not found"}, id)
	}
	return nil
}

func (p Posts) GetAll() ([]byte, error) {
	posts := []Post{}
	cursor, err := p.All.Find(p.ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("%w : GetAll : %s", myerrors.ErrInternalError, err)
	}

	err = cursor.All(p.ctx, &posts)
	if err != nil {
		return nil, fmt.Errorf("%w : GetAll : %s", myerrors.ErrInternalError, err)
	}
	str, err := json.Marshal(posts)
	if err != nil {
		return nil, fmt.Errorf("GetAll : %s", err)
	}
	return str, nil
}

func (p Posts) GetOne(id string, viewsAdd uint32) ([]byte, error) {
	post := &Post{}
	p.mu.Lock()
	err := p.findById(id, post)
	if err != nil {
		p.mu.Unlock()
		return nil, fmt.Errorf("%w : GetOne", err)
	}
	post.Views += viewsAdd
	_, err = p.All.ReplaceOne(p.ctx, bson.M{"_id": bson.ObjectIdHex(id)}, &post)
	p.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("%w : Get one : %s", myerrors.ErrInternalError, err)
	}
	re, err := json.Marshal(post)
	if err != nil {
		return nil, fmt.Errorf("%w : GetOne : cant Marshal", myerrors.ErrInternalError)
	}
	return re, nil
}

func (p Posts) CreateComment(r *http.Request, usr user.User, postId string) error {
	post := &Post{}
	p.mu.Lock()
	defer p.mu.Unlock()
	err := p.findById(postId, post)
	if err != nil {
		return fmt.Errorf("%w : CreateComment", err)
	}
	comm, err := comment.CreateComment(r, usr, uint32(len(post.Comments)))
	if err != nil {
		return fmt.Errorf("Post.CreateComment : %w", err)
	}
	post.Comments = append(post.Comments, comm)
	_, err = p.All.ReplaceOne(p.ctx, bson.M{"_id": bson.ObjectIdHex(postId)}, &post)
	if err != nil {
		return fmt.Errorf("%w : CreateComment : %s", myerrors.ErrInternalError, err)
	}
	return nil
}

func (p Posts) DeleteComment(postId, commentId string) error {
	post := &Post{}
	p.mu.Lock()
	defer p.mu.Unlock()
	err := p.findById(postId, post)
	if err != nil {
		return fmt.Errorf("%w : DeleteComment", err)
	}

	for i, c := range post.Comments {
		if c.Id == commentId {
			post.Comments = append(post.Comments[:i], post.Comments[i+1:]...)
			_, err = p.All.ReplaceOne(p.ctx, bson.M{"_id": bson.ObjectIdHex(postId)}, &post)
			if err != nil {
				return fmt.Errorf("%w : DeleteComment : %s", myerrors.ErrInternalError, err)
			}
			return nil
		}
	}
	return fmt.Errorf("%w : DeleteComment", myerrors.ErrComplex{Msg: "post not found"})
}

func (p Posts) MakeVote(postId, userId string, vote int) error {
	post := &Post{}
	p.mu.Lock()
	defer p.mu.Unlock()
	err := p.findById(postId, post)
	if err != nil {
		return fmt.Errorf("%w : MakeVote", err)
	}
	flag := true
	for i, v := range post.Votes {
		if v.UserId == userId {
			post.Score -= v.Vote
			if vote == 0 {
				post.Votes = append(post.Votes[:i], post.Votes[i+1:]...)
				flag = false
				break
			}
			v.Vote = vote
			post.Score += vote
			flag = false
			break
		}
	}
	if flag {
		post.Votes = append(post.Votes, votes.MakeVote(userId, vote))
		post.Score += vote
	}
	post.SolveScore()
	_, err = p.All.ReplaceOne(p.ctx, bson.M{"_id": bson.ObjectIdHex(postId)}, &post)
	if err != nil {
		return fmt.Errorf("%w : MakeVote : %s", myerrors.ErrInternalError, err)
	}
	return nil
}

func (p *Post) SolveScore() {
	s := 0
	for _, v := range p.Votes {
		if v.Vote == 1 {
			s++
		}
	}
	if len(p.Votes) == 0 {
		p.UpvotePercentage = 0
		return
	}
	p.UpvotePercentage = int(s * 100 / len(p.Votes))
}

func MakeJsonFromPostSlice(posts []*Post) ([]byte, error) {
	str, err := json.Marshal(posts)
	if err != nil {
		return nil, fmt.Errorf("%w : MakeJsonFromSlice : cant Marshal", myerrors.ErrInternalError)
	}
	return str, nil
}

func (p *Posts) findById(id string, post *Post) error {
	if !bson.IsObjectIdHex(id) {
		return fmt.Errorf("%w : findById : wrong post id", myerrors.ErrBadRequest)
	}
	bid := bson.ObjectIdHex(id)

	cur := p.All.FindOne(p.ctx, bson.M{"_id": bid})
	err := cur.Decode(post)
	if err == mgo.ErrNoDocuments {
		return myerrors.ErrComplex{Msg: "post not found"}
	}
	if err != nil {
		return fmt.Errorf("%w : findById : %s", myerrors.ErrInternalError, err)
	}
	return nil
}
