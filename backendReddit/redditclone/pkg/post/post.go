package post

import (
	"encoding/json"
	"fmt"
	"net/http"
	"redditclone/pkg/comment"
	"redditclone/pkg/myerrors"
	"redditclone/pkg/other"
	"redditclone/pkg/user"
	"redditclone/pkg/votes"
	"strconv"
	"sync"
	"time"
)

type Post struct {
	Score            int                `json:"score"`
	Views            uint32             `json:"views"`
	Type             string             `json:"type"`
	Title            string             `json:"title"`
	Url              string             `json:"url,omitempty"`
	Text             string             `json:"text,omitempty"`
	Author           user.User          `json:"author"`
	Category         string             `json:"category"`
	Votes            []*votes.Vote      `json:"votes"`
	Comments         []*comment.Comment `json:"comments"`
	Created          string             `json:"created"`
	UpvotePercentage int                `json:"upvotePercentage"`
	Id               string             `json:"id"`
}

type Posts struct {
	All   map[string]*Post
	mu    *sync.RWMutex
	IndID uint32
}

func (p *Posts) Init() {
	p.All = make(map[string]*Post)
	p.mu = &sync.RWMutex{}
	p.IndID = 0
}

func (p Posts) FindByCategory(category string) []Post {
	res := make([]Post, 0)
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, v := range p.All {
		if v.Category == category {
			res = append(res, *v)
		}
	}
	return res
}

func (p Posts) FindByUsername(username string) []Post {
	res := make([]Post, 0)
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, v := range p.All {
		if v.Author.Log == username {
			res = append(res, *v)
		}
	}
	return res
}

func (p *Posts) Create(r *http.Request, usr user.User) ([]byte, error) {
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
		Id:               strconv.Itoa(int(p.IndID)),
		Comments:         make([]*comment.Comment, 0),
		Votes:            make([]*votes.Vote, 0),
	}
	str, err := json.Marshal(post)
	if err != nil {
		return nil, fmt.Errorf("%w : create post : cant murshal to json", myerrors.ErrInternalError)
	}
	p.mu.Lock()
	p.All[post.Id] = post
	p.IndID++
	p.mu.Unlock()
	return str, nil
}

func (p *Posts) Delete(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.All[id]
	if !ok {
		return fmt.Errorf("%w : no user with id %s", myerrors.ErrComplex{Msg: "post not found"}, id)
	}
	delete(p.All, id)
	return nil
}

func (p *Posts) GetAll() (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	str := "["
	i := 0
	for _, v := range p.All {
		if i != 0 {
			str += ","
		}
		i++
		re, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("%w : GetAll : cant Marshal", myerrors.ErrInternalError)
		}
		str += string(re)
	}
	str += "]"
	return str, nil
}

func (p *Posts) GetOne(id string, viewsAdd uint32) ([]byte, error) {
	p.mu.RLock()
	post, ok := p.All[id]
	if !ok {
		p.mu.RUnlock()
		return nil, myerrors.ErrComplex{Msg: "post not found"}
	}
	post.Views += viewsAdd
	p.mu.RUnlock()
	re, err := json.Marshal(post)
	if err != nil {
		return nil, fmt.Errorf("%w : GetOne : cant Marshal", myerrors.ErrInternalError)
	}
	return re, nil
}

func (p *Posts) CreateComment(r *http.Request, usr user.User, postId string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	post, ok := p.All[postId]
	if !ok {
		return myerrors.ErrComplex{Msg: "post not found"}
	}
	comm, err := comment.CreateComment(r, usr, uint32(len(post.Comments)))
	if err != nil {
		return fmt.Errorf("Post.CreateComment : %w", err)
	}
	post.Comments = append(post.Comments, comm)
	return nil
}

func (p *Posts) DeleteComment(postId, commentId string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	post, ok := p.All[postId]
	if !ok {
		return myerrors.ErrComplex{Msg: "post not found"}
	}

	for i, c := range post.Comments {
		if c.Id == commentId {
			post.Comments = append(post.Comments[:i], post.Comments[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("%w : DeleteComment", myerrors.ErrComplex{Msg: "post not found"})
}

func (p *Posts) MakeVote(postId, userId string, vote int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	post, ok := p.All[postId]
	if !ok {
		return myerrors.ErrComplex{Msg: "post not found"}
	}
	defer post.SolveScore()
	for i, v := range post.Votes {
		if v.UserId == userId {
			post.Score -= v.Vote
			if vote == 0 {
				post.Votes = append(post.Votes[:i], post.Votes[i+1:]...)
				return nil
			}
			v.Vote = vote
			post.Score += vote
			return nil
		}
	}
	post.Votes = append(post.Votes, votes.MakeVote(userId, vote))
	post.Score += vote
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

func MakeJsonFromPostSlice(posts []Post) ([]byte, error) {
	str, err := json.Marshal(posts)
	if err != nil {
		return nil, fmt.Errorf("%w : MakeJsonFromSlice : cant Marshal", myerrors.ErrInternalError)
	}
	return str, nil
}
