package handlers

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"redditclone/pkg/myerrors"
	"redditclone/pkg/other"
	"redditclone/pkg/post"
	mocksP "redditclone/pkg/post/repo_mocks"
	mocksU "redditclone/pkg/session/mocks"
	"redditclone/pkg/user"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
)

func InitTests() Api {
	var a Api
	var buf bytes.Buffer
	mockWriter := bufio.NewWriter(&buf)
	a.Errorlog = log.New(mockWriter, "", 0)
	a.Infolog = log.New(mockWriter, "", 0)
	return a
}

type PostsTestCase struct {
	posts          []*post.Post
	message        string
	PostId         string
	usr            user.User
	r              *http.Request
	err            error
	resError       error
	respStatusCode int
	flag           bool
	notExpectFlag  bool
}

func TestGetAll(t *testing.T) {
	a := InitTests()
	Cases := []PostsTestCase{
		{
			message:        "posts string",
			respStatusCode: 200,
		},
		{
			message:        "posts string",
			err:            fmt.Errorf("%w : error", myerrors.ErrInternalError),
			resError:       myerrors.ErrInternalError,
			respStatusCode: 500,
		},
		{

			message:        "posts string",
			err:            fmt.Errorf("%w : error", myerrors.ErrBadRequest),
			resError:       myerrors.ErrBadRequest,
			respStatusCode: 400,
		},
		{
			message:        "posts string",
			err:            fmt.Errorf("%w : error", myerrors.ErrComplex{Msg: "error"}),
			resError:       myerrors.ErrComplex{Msg: "error"},
			respStatusCode: 422,
		},
	}

	for i, ptc := range Cases {
		w := httptest.NewRecorder()
		m := mocksP.NewPostsInterface(t)
		a.P = m
		m.On("GetAll").
			Return([]byte(ptc.message), ptc.err)

		a.Posts(w, ptc.r)
		ErrComparing(w, ptc, t, i)
	}
}

func TestOnePost(t *testing.T) {
	a := InitTests()
	Cases := []PostsTestCase{
		{
			PostId:         "abc0",
			r:              mux.SetURLVars(httptest.NewRequest("GET", "/api/post/id:{abc3}", nil), map[string]string{"id": "abc0"}),
			respStatusCode: 200,
			message:        "posts string",
		},
		{
			r:              mux.SetURLVars(httptest.NewRequest("GET", "/api/post/id:{abc3}", nil), map[string]string{"id": "abc1"}),
			PostId:         "abc1",
			err:            myerrors.ErrInternalError,
			resError:       myerrors.ErrInternalError,
			respStatusCode: 500,
		},
		{
			r:              mux.SetURLVars(httptest.NewRequest("GET", "/api/post/id:{abc3}", nil), map[string]string{"id": "abc2"}),
			PostId:         "abc2",
			err:            myerrors.ErrBadRequest,
			resError:       myerrors.ErrBadRequest,
			respStatusCode: 400,
		},
		{
			r:              mux.SetURLVars(httptest.NewRequest("GET", "/api/post/id:{abc3}", nil), map[string]string{"id": "abc3"}),
			PostId:         "abc3",
			err:            myerrors.ErrComplex{Msg: "error"},
			resError:       myerrors.ErrComplex{Msg: "error"},
			respStatusCode: 422,
		},
	}
	for i, ptc := range Cases {
		w := httptest.NewRecorder()
		m := mocksP.NewPostsInterface(t)
		a.P = m
		m.On("GetOne", ptc.PostId, uint32(1)).
			Return([]byte(ptc.message), ptc.err)

		a.OnePost(w, ptc.r)
		ErrComparing(w, ptc, t, i)
	}
}

func TestGetCategory(t *testing.T) {
	a := InitTests()
	fakePosts := []*post.Post{
		{Id: "test"},
		{Id: "test2"},
		{Id: "test3"},
	}
	fakePostStr, err := post.MakeJsonFromPostSlice(fakePosts)
	if err != nil {
		t.Fatal("test error")
	}
	Cases := []PostsTestCase{
		{
			r:              mux.SetURLVars(httptest.NewRequest("GET", "/api/post/", nil), map[string]string{"category": "cat0"}),
			posts:          fakePosts,
			PostId:         "cat0",
			message:        string(fakePostStr),
			respStatusCode: 200,
		},
		{
			r:              mux.SetURLVars(httptest.NewRequest("GET", "/api/post/", nil), map[string]string{"category": "cat1"}),
			posts:          []*post.Post{nil, nil},
			PostId:         "cat1",
			message:        "[null,null]",
			respStatusCode: 200,
		},
	}
	for i, ptc := range Cases {
		w := httptest.NewRecorder()
		m := mocksP.NewPostsInterface(t)
		a.P = m
		m.On("FindByCategory", ptc.PostId).
			Return(ptc.posts)

		a.GetCategory(w, ptc.r)
		ErrComparing(w, ptc, t, i)
	}
}

func TestGetByUser(t *testing.T) {
	a := InitTests()
	fakePosts := []*post.Post{
		{Id: "test"},
		{Id: "test2"},
		{Id: "test3"},
	}
	fakePostStr, err := post.MakeJsonFromPostSlice(fakePosts)
	if err != nil {
		t.Fatal("test error")
	}
	Cases := []PostsTestCase{
		{
			r:              mux.SetURLVars(httptest.NewRequest("GET", "/api/post/", nil), map[string]string{"id": "1"}),
			posts:          fakePosts,
			usr:            user.User{Id: "1"},
			flag:           true,
			message:        string(fakePostStr),
			respStatusCode: 200,
		},
		{
			r:              mux.SetURLVars(httptest.NewRequest("GET", "/api/post/", nil), map[string]string{"id": "1"}),
			posts:          []*post.Post{nil, nil},
			usr:            user.User{Id: "1"},
			err:            myerrors.ErrBadRequest,
			flag:           false,
			resError:       myerrors.ErrBadRequest,
			respStatusCode: 400,
		},
	}

	for i, ptc := range Cases {
		w := httptest.NewRecorder()
		m := mocksP.NewPostsInterface(t)
		m2 := mocksU.NewUsersRepoInterface(t)
		a.U = m2
		a.P = m

		m2.On("Find", ptc.usr.Id).
			Return(ptc.usr, ptc.err)

		if ptc.flag {
			m.On("FindByUsername", ptc.usr).
				Return(ptc.posts)
		}
		a.GetPostsByUser(w, ptc.r)
		ErrComparing(w, ptc, t, i)
	}
}

func TestDeletePost(t *testing.T) {
	a := InitTests()
	Cases := []PostsTestCase{
		{
			r:              mux.SetURLVars(httptest.NewRequest("DELETE", "/api/post/", nil), map[string]string{"id": "abc0"}),
			PostId:         "abc0",
			err:            nil,
			message:        `{"message": "success"}`,
			respStatusCode: 200,
		},
		{
			r:              mux.SetURLVars(httptest.NewRequest("DELETE", "/api/post/", nil), map[string]string{"id": "abc1"}),
			PostId:         "abc1",
			err:            fmt.Errorf("%w : error", myerrors.ErrInternalError),
			resError:       myerrors.ErrInternalError,
			respStatusCode: 500,
		},
		{
			r:              mux.SetURLVars(httptest.NewRequest("DELETE", "/api/post/", nil), map[string]string{"id": "abc2"}),
			PostId:         "abc2",
			err:            fmt.Errorf("%w : error", myerrors.ErrBadRequest),
			resError:       myerrors.ErrBadRequest,
			respStatusCode: 400,
		},
		{
			r:              mux.SetURLVars(httptest.NewRequest("DELETE", "/api/post/", nil), map[string]string{"id": "abc3"}),
			PostId:         "abc3",
			err:            fmt.Errorf("%w : error", myerrors.ErrComplex{Msg: "error"}),
			resError:       myerrors.ErrComplex{Msg: "error"},
			respStatusCode: 422,
		},
	}
	for i, ptc := range Cases {
		w := httptest.NewRecorder()
		m := mocksP.NewPostsInterface(t)
		a.P = m
		m.On("Delete", ptc.PostId).
			Return(ptc.err)

		a.DeletePost(w, ptc.r)
		ErrComparing(w, ptc, t, i)
	}
}

func TestVote(t *testing.T) {
	a := InitTests()
	ctx := context.WithValue(context.Background(), other.CtxValue, user.User{Id: "1"})
	Cases := []PostsTestCase{
		{
			r: mux.SetURLVars(httptest.NewRequest("GET", "/api/post/", nil).
				WithContext(ctx),
				map[string]string{
					"id":   "1",
					"vote": "downvote",
				}),
			PostId:         "1",
			usr:            user.User{Id: "1"},
			flag:           true,
			message:        "posts string",
			respStatusCode: 200,
		},
		{
			r: mux.SetURLVars(httptest.NewRequest("GET", "/api/post/", nil).
				WithContext(ctx),
				map[string]string{
					"id":   "2",
					"vote": "unvote",
				}),
			PostId:         "2",
			usr:            user.User{Id: "1"},
			flag:           true,
			message:        "posts string",
			respStatusCode: 200,
		},
		{
			PostId: "3",
			usr: user.User{
				Id: "1",
			},
			r: mux.SetURLVars(httptest.NewRequest("GET", "/api/post/", nil).
				WithContext(ctx),
				map[string]string{
					"id":   "3",
					"vote": "upvote",
				}),
			err: myerrors.ErrComplex{
				Msg:        "error",
				StatusCode: 404,
			},
			flag:           false,
			resError:       fmt.Errorf(`{"message":"error"}`),
			respStatusCode: 404,
		},
		{
			r: mux.SetURLVars(httptest.NewRequest("GET", "/api/post/", nil).
				WithContext(ctx),
				map[string]string{
					"id":   "4",
					"vote": "upvote",
				}),
			PostId: "4",
			usr:    user.User{Id: "1"},
			err: myerrors.ErrComplex{
				Msg:        "error",
				StatusCode: 404,
			},
			flag:           true,
			resError:       fmt.Errorf(`{"message":"error"}`),
			respStatusCode: 404,
		},
		{
			r:              httptest.NewRequest("GET", "/api/post/", nil).WithContext(context.Background()),
			flag:           false,
			notExpectFlag:  true,
			resError:       myerrors.ErrInternalError,
			respStatusCode: 500,
		},
	}
	for i, ptc := range Cases {
		w := httptest.NewRecorder()
		m := mocksP.NewPostsInterface(t)
		a.P = m

		if !ptc.notExpectFlag {
			if ptc.flag {
				m.On("MakeVote", ptc.PostId, ptc.usr.Id, mock.AnythingOfType("int")).
					Return(nil)
				m.On("GetOne", ptc.PostId, uint32(0)).
					Return([]byte(ptc.message), ptc.err)
			} else {
				m.On("MakeVote", ptc.PostId, ptc.usr.Id, mock.AnythingOfType("int")).
					Return(ptc.err)
			}
		}
		a.MakeVote(w, ptc.r)
		ErrComparing(w, ptc, t, i)
	}
}

func TestDeleteComment(t *testing.T) {
	a := InitTests()
	Cases := []PostsTestCase{
		{
			r: mux.SetURLVars(httptest.NewRequest("DELETE", "/api/post/", nil), map[string]string{
				"id":        "1",
				"commentId": "12",
			}),
			PostId:         "1",
			flag:           true,
			message:        "posts string",
			respStatusCode: 200,
		},
		{
			r: mux.SetURLVars(httptest.NewRequest("DELETE", "/api/post/", nil), map[string]string{
				"id":        "2",
				"commentId": "12",
			}),
			PostId: "2",
			err: myerrors.ErrComplex{
				Msg:        "error",
				StatusCode: 404,
			},
			resError:       fmt.Errorf(`{"message":"error"}`),
			flag:           false,
			respStatusCode: 404,
		},
		{
			r: mux.SetURLVars(httptest.NewRequest("DELETE", "/api/post/", nil), map[string]string{
				"id":        "2",
				"commentId": "12",
			}),
			PostId: "2",
			err: myerrors.ErrComplex{
				Msg:        "error",
				StatusCode: 404,
			},
			resError:       fmt.Errorf(`{"message":"error"}`),
			flag:           true,
			respStatusCode: 404,
		},
	}
	for i, ptc := range Cases {
		w := httptest.NewRecorder()
		m := mocksP.NewPostsInterface(t)
		a.P = m

		if ptc.flag {
			m.On("DeleteComment", ptc.PostId, "12").
				Return(nil)
			m.On("GetOne", ptc.PostId, uint32(0)).
				Return([]byte(ptc.message), ptc.err)
		} else {
			m.On("DeleteComment", ptc.PostId, "12").
				Return(ptc.err)
		}
		a.DeleteComment(w, ptc.r)
		ErrComparing(w, ptc, t, i)
	}
}

func TestCreateComment(t *testing.T) {
	a := InitTests()
	ctx := context.WithValue(context.Background(), other.CtxValue, user.User{Id: "1"})
	Cases := []PostsTestCase{
		{
			r: mux.SetURLVars(httptest.NewRequest("POST", "/api/post/", nil).
				WithContext(ctx),
				map[string]string{
					"id": "1",
				}),
			PostId:         "1",
			usr:            user.User{Id: "1"},
			flag:           true,
			message:        "posts string",
			respStatusCode: 200,
		},
		{
			r: mux.SetURLVars(httptest.NewRequest("POST", "/api/post/", nil).
				WithContext(ctx),
				map[string]string{
					"id": "3",
				}),
			PostId: "3",
			usr:    user.User{Id: "1"},
			err: myerrors.ErrComplex{
				Msg:        "error",
				StatusCode: 404,
			},
			flag:           false,
			resError:       fmt.Errorf(`{"message":"error"}`),
			respStatusCode: 404,
		},
		{
			r: mux.SetURLVars(httptest.NewRequest("POST", "/api/post/", nil).
				WithContext(ctx),
				map[string]string{
					"id": "4",
				}),
			PostId: "4",
			usr:    user.User{Id: "1"},
			err: myerrors.ErrComplex{
				Msg:        "error",
				StatusCode: 404,
			},
			flag:           true,
			resError:       fmt.Errorf(`{"message":"error"}`),
			respStatusCode: 404,
		},
		{
			r:              httptest.NewRequest("POST", "/api/post/", nil).WithContext(context.Background()),
			flag:           false,
			notExpectFlag:  true,
			resError:       myerrors.ErrInternalError,
			respStatusCode: 500,
		},
	}
	for i, ptc := range Cases {
		w := httptest.NewRecorder()
		m := mocksP.NewPostsInterface(t)
		a.P = m

		if !ptc.notExpectFlag {
			if ptc.flag {
				m.On("CreateComment", ptc.r, ptc.usr, ptc.PostId).
					Return(nil)
				m.On("GetOne", ptc.PostId, uint32(0)).
					Return([]byte(ptc.message), ptc.err)
			} else {
				m.On("CreateComment", ptc.r, ptc.usr, ptc.PostId).
					Return(ptc.err)
			}
		}
		a.CreateComment(w, ptc.r)
		ErrComparing(w, ptc, t, i)
	}
}

func TestCreatePost(t *testing.T) {
	a := InitTests()
	ctx := context.WithValue(context.Background(), other.CtxValue, user.User{Id: "1"})
	Cases := []PostsTestCase{
		{
			r:              httptest.NewRequest("POST", "/api/posts", nil).WithContext(ctx),
			usr:            user.User{Id: "1"},
			message:        "posts string",
			respStatusCode: 200,
		},
		{
			r:   httptest.NewRequest("POST", "/api/posts", nil).WithContext(ctx),
			usr: user.User{Id: "1"},
			err: myerrors.ErrComplex{
				Msg:        "error",
				StatusCode: 404,
			},
			resError:       fmt.Errorf(`{"message":"error"}`),
			respStatusCode: 404,
		},
		{
			r:              httptest.NewRequest("POST", "/api/posts", nil).WithContext(context.Background()),
			notExpectFlag:  true,
			resError:       myerrors.ErrInternalError,
			respStatusCode: 500,
		},
	}
	for i, ptc := range Cases {
		w := httptest.NewRecorder()
		m := mocksP.NewPostsInterface(t)
		a.P = m

		if !ptc.notExpectFlag {
			m.On("Create", ptc.r, ptc.usr).
				Return([]byte(ptc.message), ptc.err)
		}
		a.PostCreate(w, ptc.r)
		ErrComparing(w, ptc, t, i)
	}
}

func TestRegister(t *testing.T) {
	a := InitTests()
	Cases := []PostsTestCase{
		{
			r:              httptest.NewRequest("POST", "/api/register", nil),
			usr:            user.User{Id: "created"},
			flag:           true,
			message:        `{"token" : "created"}`,
			respStatusCode: 200,
		},
		{
			r:   httptest.NewRequest("POST", "/api/register", nil),
			usr: user.User{Id: "1"},
			err: myerrors.ErrComplex{
				Msg:        "error",
				StatusCode: 404,
			},
			resError:       fmt.Errorf(`{"message":"error"}`),
			flag:           false,
			respStatusCode: 404,
		},
		{
			r:   httptest.NewRequest("POST", "/api/register", nil),
			usr: user.User{Id: "1"},
			err: myerrors.ErrComplex{
				Msg:        "error",
				StatusCode: 404,
			},
			resError:       fmt.Errorf(`{"message":"error"}`),
			flag:           true,
			respStatusCode: 404,
		},
	}
	for i, ptc := range Cases {
		w := httptest.NewRecorder()
		m := mocksU.NewUsersRepoInterface(t)
		a.U = m

		if ptc.flag {
			m.On("CreateUser", ptc.r).
				Return(ptc.usr, nil)
			m.On("AddSession", ptc.usr).
				Return(ptc.usr.Id, ptc.err)
		} else {
			m.On("CreateUser", ptc.r).
				Return(ptc.usr, ptc.err)
		}
		a.RegHand(w, ptc.r)
		ErrComparing(w, ptc, t, i)
	}
}

func TestLogin(t *testing.T) {
	a := InitTests()
	Cases := []PostsTestCase{
		{
			r: httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"password":"pass","username":"loged"}`)),
			usr: user.User{
				Log:  "loged",
				Pass: "pass",
			},
			flag:           true,
			message:        `{"token" : "loged"}`,
			respStatusCode: 200,
		},
		{
			r: httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"password":"pass","username":"loged"}`)),
			usr: user.User{
				Log:  "loged",
				Pass: "outerpass",
			},
			resError:       myerrors.ErrComplex{Msg: "invalid password"},
			respStatusCode: 401,
		},
		{
			r:              httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"password":"pass","username":""}`)),
			usr:            user.User{Log: "loged"},
			notExpectFlag:  true,
			resError:       myerrors.ErrComplex{Msg: "user not found"},
			respStatusCode: 401,
		},
		{
			r:              httptest.NewRequest("POST", "/api/login", nil),
			notExpectFlag:  true,
			resError:       myerrors.ErrInternalError,
			respStatusCode: 500,
		},
		{
			r:   httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"password":"pass","username":"loged"}`)),
			usr: user.User{Log: "loged"},
			err: myerrors.ErrComplex{
				Msg:        "error",
				StatusCode: 404,
			},
			flag:           false,
			resError:       fmt.Errorf(`{"message":"error"}`),
			respStatusCode: 404,
		},
		{
			r: httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"password":"pass","username":"loged"}`)),
			usr: user.User{
				Log:  "loged",
				Pass: "pass",
			},
			err: myerrors.ErrComplex{
				Msg:        "error",
				StatusCode: 404,
			},
			flag:           true,
			resError:       fmt.Errorf(`{"message":"error"}`),
			respStatusCode: 404,
		},
	}

	for i, ptc := range Cases {
		w := httptest.NewRecorder()
		m := mocksU.NewUsersRepoInterface(t)
		a.U = m
		if !ptc.notExpectFlag {
			if ptc.flag {
				m.On("Find", ptc.usr.Log).
					Return(ptc.usr, nil)
				m.On("AddSession", ptc.usr).
					Return(ptc.usr.Log, ptc.err)
			} else {
				m.On("Find", ptc.usr.Log).
					Return(ptc.usr, ptc.err)
			}
		}
		a.SignHand(w, ptc.r)
		ErrComparing(w, ptc, t, i)
	}
}

func ErrComparing(w *httptest.ResponseRecorder, ptc PostsTestCase, t *testing.T, i int) {
	resp := w.Result()
	byteBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("test error")
	}
	body := string(byteBody)

	if resp.StatusCode != ptc.respStatusCode {
		t.Errorf("case [%d]: expected %d got %d", i, ptc.respStatusCode, resp.StatusCode)
		return
	}

	if ptc.resError != nil && ptc.resError.Error() != body[:len(body)-1] {
		t.Errorf("case [%d]: expected : %s\ngot: %s", i, ptc.resError, body[:len(body)-1])
		return
	}

	if ptc.resError == nil {
		if body != ptc.message {
			t.Errorf("case [%d]: expected : %s\ngot: %s", i, ptc.message, body)
		}
	}
}
