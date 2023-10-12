package post

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"redditclone/pkg/comment"
	"redditclone/pkg/myerrors"
	"redditclone/pkg/post/mocks"
	"redditclone/pkg/user"
	"redditclone/pkg/votes"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"
	mongo "go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

type OnePostTestCase struct {
	postID bson.ObjectId
	post   Post
	err    error
	resErr error
}

func TestFindById(t *testing.T) {
	id := bson.NewObjectId()
	Cases := []OnePostTestCase{
		{
			postID: "NOTBSONID",
			resErr: fmt.Errorf("%w : findById : wrong post id", myerrors.ErrBadRequest),
		},
		{
			postID: id,
			post: Post{
				Id:       id,
				Category: "testing",
			},
		},
		{
			postID: id,
			err:    mongo.ErrNoDocuments,
			resErr: myerrors.ErrComplex{Msg: "post not found"},
		},
		{
			postID: id,
			err:    mongo.ErrNilDocument,
			resErr: fmt.Errorf("%w : findById : %s", myerrors.ErrInternalError, mongo.ErrNilDocument),
		},
	}
	P := makePosts()
	np := Post{}

	for i, f := range Cases {
		m := mocks.NewCollectionHelper(t)
		P.All = m
		if f.postID != "NOTBSONID" {
			m.On("FindOne", P.ctx, bson.M{"_id": f.postID}).
				Return(mongo.NewSingleResultFromDocument(f.post, f.err, nil))
		}
		err := P.findById(f.postID.Hex(), &np)

		ErrComparing(err, f.resErr, t, i)

		if f.resErr == nil && !reflect.DeepEqual(f.post, np) {
			t.Errorf("case [%d]: expected : %#v\ngot: %#v", i, f.post, np)
			return
		}
	}
}

type DeleteCases struct {
	postID       bson.ObjectId
	deletecCount int64
	err          error
	resErr       error
}

func TestDeletePost(t *testing.T) {
	id := bson.NewObjectId()
	Cases := []DeleteCases{
		{
			postID: "NOTBSONID",
			resErr: fmt.Errorf("%w : Delete post : wrong post id", myerrors.ErrBadRequest),
		},
		{
			postID:       id,
			deletecCount: 1,
		},
		{
			postID:       id,
			deletecCount: 0,
			resErr:       fmt.Errorf("%w : no post with id : %s", myerrors.ErrComplex{Msg: "post not found"}, id.Hex()),
		},
		{
			postID: id,
			err:    fmt.Errorf("big error"),
			resErr: fmt.Errorf("%w : delete post : %s", myerrors.ErrInternalError, fmt.Errorf("big error")),
		},
	}
	P := makePosts()

	for i, f := range Cases {
		m := mocks.NewCollectionHelper(t)
		P.All = m
		if f.postID != "NOTBSONID" {
			m.On("DeleteOne", P.ctx, bson.M{"_id": f.postID}).
				Return(&mongo.DeleteResult{
					DeletedCount: f.deletecCount,
				}, f.err)
		}
		err := P.Delete(f.postID.Hex())

		ErrComparing(err, f.resErr, t, i)
	}
}

func TestGetOne(t *testing.T) {
	id := bson.NewObjectId()
	Cases := []OnePostTestCase{
		{
			postID: "NOTBSONID",
			resErr: fmt.Errorf("%w : findById : wrong post id : GetOne", myerrors.ErrBadRequest),
		},
		{
			postID: id,
			post: Post{
				Id:    id,
				Views: 5,
			},
		},
		{
			postID: id,
			post: Post{
				Id:    id,
				Views: 5,
			},
			err:    fmt.Errorf("big replace error"),
			resErr: fmt.Errorf("internal error : Get one : big replace error"),
		},
	}
	P := makePosts()

	for i, f := range Cases {
		m := mocks.NewCollectionHelper(t)
		P.All = m
		if f.postID != "NOTBSONID" {
			m.On("FindOne", P.ctx, bson.M{"_id": f.postID}).
				Return(mongo.NewSingleResultFromDocument(f.post, nil, nil))
			f.post.Views++
			tmp := &f.post
			m.On("ReplaceOne", P.ctx, bson.M{"_id": f.postID}, &tmp).
				Return(nil, f.err)
		}

		resStr, err := P.GetOne(f.postID.Hex(), 1)

		ErrComparing(err, f.resErr, t, i)

		if f.resErr == nil {
			str, _ := json.Marshal(f.post)
			if string(str) != string(resStr) {
				t.Errorf("case [%d]: expected : %#v\ngot: %#v", i, str, resStr)
				return
			}
		}
	}
}

func ErrComparing(err, resErr error, t *testing.T, i int) {
	if err == nil && resErr != nil {
		t.Errorf("case [%d]: expected %s got nil", i, resErr)
		return
	}

	if err != nil && resErr == nil {
		t.Errorf("case [%d]: unexpected error : %s", i, err)
		return
	}

	if resErr != nil && resErr.Error() != err.Error() {
		t.Errorf("case [%d]: expected : %s\ngot: %s", i, resErr, err)
		return
	}
}

type GetGetByCategory struct {
	category string
	user     user.User
	posts    []*Post
	err      error
}

func TestGetByCategoty(t *testing.T) {
	id := bson.NewObjectId()
	Cases := []GetGetByCategory{
		{
			category: "cat",
			posts:    []*Post{},
		},
		{
			category: "cat",
			posts: []*Post{
				{
					Id:       id,
					Category: "cat",
				},
				{
					Id:       bson.NewObjectId(),
					Category: "cat",
				},
			},
		},
		{
			category: "cat",
			err:      fmt.Errorf("big error"),
			posts:    []*Post{nil, nil},
		},

		{
			user: user.User{
				Id:  "123",
				Log: "123",
			},
			posts: []*Post{},
		},
		{
			user: user.User{
				Id:  "123",
				Log: "123",
			},
			posts: []*Post{
				{
					Id:       id,
					Category: "cat",
				},
				{
					Id:       bson.NewObjectId(),
					Category: "cat",
				},
			},
		},
		{
			user: user.User{
				Id:  "123",
				Log: "123",
			},
			err:   fmt.Errorf("big error"),
			posts: []*Post{nil, nil},
		},
	}

	P := makePosts()

	for i, f := range Cases {
		m := mocks.NewCollectionHelper(t)
		P.All = m
		var posts []*Post
		if f.category != "" {
			m.On("Find", P.ctx, bson.M{"category": "cat"}).
				Return(mongo.NewCursorFromDocuments(func() (arr []interface{}) {
					for _, v := range f.posts {
						arr = append(arr, v)
					}
					return
				}(), f.err, nil))
			posts = P.FindByCategory("cat")
		} else if f.user.Id != "" {
			m.On("Find", P.ctx, bson.M{
				"author": bson.M{
					"id":       f.user.ID(),
					"username": f.user.Login(),
				},
			}).
				Return(mongo.NewCursorFromDocuments(makeEmptyInterface(f.posts), f.err, nil))
			posts = P.FindByUsername(f.user)
		}
		if posts != nil && f.err != nil {
			t.Errorf("case [%d]: expected %s got nil", i, f.err)
			return
		}

		if posts == nil && f.err == nil {
			t.Errorf("case [%d]: unexpected error", i)
			return
		}
		if f.err == nil && !reflect.DeepEqual(f.posts, posts) {
			t.Errorf("case [%d]: expected : %#v\ngot: %#v", i, f.posts, posts)
			return
		}
	}
}

type DeleteCommentTestCase struct {
	postID bson.ObjectId
	post   Post
	target string
	err    error
	resErr error
}

func TestDeleteComment(t *testing.T) {
	id := bson.NewObjectId()
	comments := []*comment.Comment{
		{Id: "1"},
		{Id: "2"},
		{Id: "3"},
	}
	Cases := []DeleteCommentTestCase{
		{
			postID: "NOTBSONID",
			resErr: fmt.Errorf("%w : findById : wrong post id : DeleteComment", myerrors.ErrBadRequest),
		},
		{
			postID: id,
			post: Post{
				Id:       id,
				Comments: comments,
			},
			target: "2",
		},
		{
			postID: id,
			post: Post{
				Id:       id,
				Comments: comments,
			},
			target: "NotExists",
			resErr: fmt.Errorf("%w : DeleteComment", myerrors.ErrComplex{Msg: "post not found"}),
		},
		{
			postID: id,
			post: Post{
				Id:       id,
				Comments: comments,
			},
			target: "1",
			err:    fmt.Errorf("big replace error"),
			resErr: fmt.Errorf("internal error : DeleteComment : big replace error"),
		},
	}
	P := makePosts()

	for i, f := range Cases {
		m := mocks.NewCollectionHelper(t)
		P.All = m
		if f.postID != "NOTBSONID" {
			m.On("FindOne", P.ctx, bson.M{"_id": f.postID}).
				Return(mongo.NewSingleResultFromDocument(f.post, nil, nil))
			if f.target != "NotExists" {
				m.On("ReplaceOne", P.ctx, bson.M{"_id": f.postID}, mock.AnythingOfType("**post.Post")).
					Return(nil, f.err)
			}
		}

		err := P.DeleteComment(f.postID.Hex(), f.target)

		ErrComparing(err, f.resErr, t, i)
	}
}

type CreateCommentTestCase struct {
	postID      bson.ObjectId
	post        Post
	err         error
	resErr      error
	req         *http.Request
	replaceFlag bool
}

func TestCreateComment(t *testing.T) {
	id := bson.NewObjectId()
	comments := []*comment.Comment{
		{Id: "1"},
		{Id: "2"},
		{Id: "3"},
	}
	Cases := []CreateCommentTestCase{
		{
			postID: "NOTBSONID",
			resErr: fmt.Errorf("%w : findById : wrong post id : CreateComment", myerrors.ErrBadRequest),
		},
		{
			postID: id,
			post: Post{
				Id:       id,
				Comments: comments,
			},
			req:         httptest.NewRequest("POST", "/api/post/{id:[a-zA-Z0-9]+}", strings.NewReader(`{"comment":"CreateTestCase"}`)),
			replaceFlag: true,
		},
		{
			postID: id,
			post: Post{
				Id:       id,
				Comments: comments,
			},
			resErr:      fmt.Errorf("Post.CreateComment : Comment Create: internal error : cant get map from post : invalid character 'b' looking for beginning of value"),
			req:         httptest.NewRequest("POST", "/api/post/{id:[a-zA-Z0-9]+}", strings.NewReader(`blablabla`)),
			replaceFlag: false,
		},
		{
			postID: id,
			post: Post{
				Id:       id,
				Comments: comments,
			},
			err:         fmt.Errorf("big replace error"),
			resErr:      fmt.Errorf("internal error : CreateComment : big replace error"),
			req:         httptest.NewRequest("POST", "/api/post/{id:[a-zA-Z0-9]+}", strings.NewReader(`{"comment":"CreateTestCase"}`)),
			replaceFlag: true,
		},
	}
	P := makePosts()

	for i, f := range Cases {
		m := mocks.NewCollectionHelper(t)
		P.All = m
		if f.postID != "NOTBSONID" {
			m.On("FindOne", P.ctx, bson.M{"_id": f.postID}).
				Return(mongo.NewSingleResultFromDocument(f.post, nil, nil))
			if f.replaceFlag {
				m.On("ReplaceOne", P.ctx, bson.M{"_id": f.postID}, mock.AnythingOfType("**post.Post")).
					Return(nil, f.err)
			}
		}
		err := P.CreateComment(f.req, user.User{}, f.postID.Hex())

		ErrComparing(err, f.resErr, t, i)
	}
}

type VoteTestCase struct {
	postID      bson.ObjectId
	vote        int
	userID      string
	post        Post
	err         error
	resErr      error
	replaceFlag bool
}

func TestVote(t *testing.T) {
	id := bson.NewObjectId()
	Votes := []*votes.Vote{
		{UserId: "1", Vote: 1},
		{UserId: "2", Vote: -1},
		{UserId: "3", Vote: -1},
		{UserId: "4", Vote: 1},
	}
	Cases := []VoteTestCase{
		{
			postID: "NOTBSONID",
			resErr: fmt.Errorf("%w : findById : wrong post id : MakeVote", myerrors.ErrBadRequest),
		},
		{
			postID: id,
			post: Post{
				Id:    id,
				Votes: Votes,
			},
			replaceFlag: true,
			userID:      "5",
			vote:        1,
		},
		{
			postID: id,
			post: Post{
				Id:    id,
				Votes: Votes,
			},
			replaceFlag: true,
			userID:      "1",
			vote:        -1,
		},
		{
			postID: id,
			post: Post{
				Id:    id,
				Votes: Votes,
			},
			replaceFlag: true,
			userID:      "2",
			vote:        -1,
		},
		{
			postID: id,
			post: Post{
				Id:    id,
				Votes: Votes[:1],
			},
			replaceFlag: true,
			userID:      "1",
			vote:        0,
		},
		{
			postID: id,
			post: Post{
				Id:    id,
				Votes: Votes,
			},
			err:         fmt.Errorf("big replace error"),
			resErr:      fmt.Errorf("internal error : MakeVote : big replace error"),
			replaceFlag: true,
		},
	}
	P := makePosts()

	for i, f := range Cases {
		m := mocks.NewCollectionHelper(t)
		P.All = m
		if f.postID != "NOTBSONID" {
			m.On("FindOne", P.ctx, bson.M{"_id": f.postID}).
				Return(mongo.NewSingleResultFromDocument(f.post, nil, nil))
			if f.replaceFlag {
				m.On("ReplaceOne", P.ctx, bson.M{"_id": f.postID}, mock.AnythingOfType("**post.Post")).
					Return(nil, f.err)
			}
		}
		err := P.MakeVote(f.postID.Hex(), f.userID, f.vote)

		ErrComparing(err, f.resErr, t, i)
	}
}

type CreateTestCase struct {
	err        error
	resErr     error
	req        *http.Request
	insertFlag bool
}

func TestCreate(t *testing.T) {
	usr := user.User{
		Id:  "1",
		Log: "1",
	}
	Cases := []CreateTestCase{
		{
			req: httptest.NewRequest("POST", "/api/posts",
				strings.NewReader(`{"category":"music","type":"text","title":"create post","text":"create post"}`)),
			insertFlag: true,
		},
		{
			resErr: fmt.Errorf("bad request : create post : wrong parametrs"),
			req: httptest.NewRequest("POST", "/api/posts",
				strings.NewReader(`{"category":"music","title":"create post","text":"create post"}`)),
			insertFlag: false,
		},
		{
			resErr:     fmt.Errorf("create post : internal error : cant get map from post : invalid character 'b' looking for beginning of value"),
			req:        httptest.NewRequest("POST", "/api/posts", strings.NewReader(`blablabla`)),
			insertFlag: false,
		},
		{
			err:    fmt.Errorf("big insert error"),
			resErr: fmt.Errorf("internal error : create post : big insert error"),
			req: httptest.NewRequest("POST", "/api/posts",
				strings.NewReader(`{"category":"music","type":"text","title":"create post","text":"create post"}`)),
			insertFlag: true,
		},
	}
	P := makePosts()

	for i, f := range Cases {
		m := mocks.NewCollectionHelper(t)
		P.All = m
		if f.insertFlag {
			m.On("InsertOne", P.ctx, mock.AnythingOfType("*post.Post")).
				Return(nil, f.err)
		}
		_, err := P.Create(f.req, usr)
		ErrComparing(err, f.resErr, t, i)
	}
}

type GetByAllTestCase struct {
	posts  []*Post
	err    error
	resErr error
}

func TestGetAll(t *testing.T) {
	id := bson.NewObjectId()
	posts := []*Post{
		{Id: id},
		{Id: bson.NewObjectId()},
		{Id: bson.NewObjectId()},
	}
	Cases := []GetByAllTestCase{
		{
			posts: posts,
		},
		{
			err:    fmt.Errorf("big error"),
			posts:  []*Post{nil, nil},
			resErr: fmt.Errorf("internal error : GetAll : WriteNull can only write while positioned on a Element or Value but is positioned on a TopLevel"),
		},
	}
	P := makePosts()

	for i, f := range Cases {
		m := mocks.NewCollectionHelper(t)
		P.All = m
		m.On("Find", P.ctx, bson.M{}).
			Return(mongo.NewCursorFromDocuments(makeEmptyInterface(f.posts), f.err, nil))
		str, err := P.GetAll()

		ErrComparing(err, f.resErr, t, i)
		if f.resErr == nil {
			resStr, _ := MakeJsonFromPostSlice(posts)
			if string(resStr) != string(str) {
				t.Errorf("case [%d]: expected : %s\ngot: %s", i, resStr, str)
			}
		}
	}
}

func makePosts() Posts {
	P := Posts{}
	P.mu = &sync.Mutex{}
	P.ctx = context.Background()
	return P
}

func makeEmptyInterface(f []*Post) (arr []interface{}) {
	for _, v := range f {
		arr = append(arr, v)
	}
	return
}
