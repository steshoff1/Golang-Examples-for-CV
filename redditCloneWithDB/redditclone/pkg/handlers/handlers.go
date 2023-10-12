package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"redditclone/pkg/myerrors"
	"redditclone/pkg/other"
	"redditclone/pkg/post"
	"redditclone/pkg/session"
	"redditclone/pkg/user"

	"github.com/gorilla/mux"
)

type Api struct {
	U        session.UsersRepoInterface
	P        post.PostsInterface
	Errorlog *log.Logger
	Infolog  *log.Logger
}

func (a *Api) Init() error {
	a.U = &session.SqlUsers{}
	err := a.U.Init()
	if err != nil {
		return fmt.Errorf("Api init : %w", err)
	}
	a.P = &post.Posts{}
	err = a.P.Init()
	if err != nil {
		return fmt.Errorf("Api init : %w", err)
	}
	a.Errorlog = log.New(os.Stdout, "[error]: ", log.LUTC|log.Ldate|log.Ltime)
	a.Infolog = log.New(os.Stdout, "[info]: ", log.LUTC|log.Ldate|log.Ltime)
	return nil
}

func (a *Api) Destroy() {
	a.U.Destroy()
	a.P.Destroy()
}

func (a *Api) errorHandler(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, myerrors.ErrInternalError):
		http.Error(w, myerrors.ErrInternalError.Error(), http.StatusInternalServerError)
	case errors.Is(err, myerrors.ErrBadRequest):
		http.Error(w, myerrors.ErrBadRequest.Error(), http.StatusBadRequest)
	case errors.As(err, &myerrors.ErrComplex{}):
		tmpErr := myerrors.MyUnwrap(err).(myerrors.ErrComplex)
		if tmpErr.StatusCode != 0 {
			http.Error(w, tmpErr.Error(), tmpErr.StatusCode)
		} else {
			http.Error(w, tmpErr.Error(), http.StatusUnprocessableEntity)
		}
	}
	a.Errorlog.Println(err)
}

func (a *Api) Write(w http.ResponseWriter, str []byte) {
	_, err := w.Write([]byte(str))
	if err != nil {
		a.errorHandler(w, fmt.Errorf("%w : %s", myerrors.ErrInternalError, err))
	}
}

func (a Api) RegHand(w http.ResponseWriter, r *http.Request) {
	usr, err := a.U.CreateUser(r)
	if err != nil {
		a.errorHandler(w, err)
		return
	}
	a.Infolog.Println("user created")

	token, err := a.U.AddSession(usr)
	if err != nil {
		a.errorHandler(w, err)
		return
	}
	a.Write(w, []byte(fmt.Sprintf(`{"token" : "%s"}`, token)))
}

func (a Api) SignHand(w http.ResponseWriter, r *http.Request) {
	m, err := other.GetMapFromPost(r)
	r.Body.Close()
	if err != nil {
		a.errorHandler(w, err)
		return
	}

	username := m["username"]
	password := m["password"]
	if username == "" || password == "" {
		http.Error(w, myerrors.ErrComplex{Msg: "user not found"}.Error(), http.StatusUnauthorized)
		a.Errorlog.Println("SignHand : username or password empty")
		return
	}

	usr, err := a.U.Find(username)
	if err != nil {
		a.errorHandler(w, err)
		return
	}

	if usr.Password() != password {
		http.Error(w, myerrors.ErrComplex{Msg: "invalid password"}.Error(), http.StatusUnauthorized)
		a.Infolog.Println("SignHand : wrong password")
		return
	}

	token, err := a.U.AddSession(usr)
	if err != nil {
		a.errorHandler(w, err)
		return
	}
	a.Infolog.Println("user loged")
	a.Write(w, []byte(fmt.Sprintf(`{"token" : "%s"}`, token)))
}

func (a Api) PostCreate(w http.ResponseWriter, r *http.Request) {
	usr, ok := r.Context().Value(other.CtxValue).(user.User)
	if !ok {
		a.errorHandler(w, fmt.Errorf("%w : can't get user from context", myerrors.ErrInternalError))
		return
	}
	post, err := a.P.Create(r, usr)

	if err != nil {
		a.errorHandler(w, err)
		return
	}
	a.Infolog.Printf("post created %s\n", post)
	a.Write(w, post)
}

func (a Api) Posts(w http.ResponseWriter, r *http.Request) {
	posts, err := a.P.GetAll()
	if err != nil {
		a.errorHandler(w, err)
		return
	}
	a.Infolog.Println("All posts")
	a.Write(w, posts)
}

func (a Api) OnePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	post, err := a.P.GetOne(id, 1)
	if err != nil {
		a.errorHandler(w, err)
		return
	}
	a.Infolog.Printf("Send one post with id : %s", id)
	a.Write(w, post)
}

func (a Api) GetCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["category"]
	posts := a.P.FindByCategory(id)
	json, err := post.MakeJsonFromPostSlice(posts)
	if err != nil {
		a.errorHandler(w, err)
		return
	}
	a.Infolog.Printf("posts by category : %s", id)
	a.Write(w, json)
}

func (a Api) GetPostsByUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	usr, err := a.U.Find(id)
	if err != nil {
		a.errorHandler(w, fmt.Errorf("%w : GetPostsByUser", err))
		return
	}
	posts := a.P.FindByUsername(usr)
	json, err := post.MakeJsonFromPostSlice(posts)
	if err != nil {
		a.errorHandler(w, err)
		return
	}
	a.Infolog.Printf("posts by user : %s", id)
	a.Write(w, json)
}

func (a Api) DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	err := a.P.Delete(id)
	if err != nil {
		a.errorHandler(w, err)
		return
	}
	a.Infolog.Printf("deleted post with id : %s", id)
	a.Write(w, []byte(`{"message": "success"}`))
}

func (a Api) CreateComment(w http.ResponseWriter, r *http.Request) {
	usr, ok := r.Context().Value(other.CtxValue).(user.User)
	if !ok {
		a.errorHandler(w, fmt.Errorf("%w : can't get user from context", myerrors.ErrInternalError))
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]
	err := a.P.CreateComment(r, usr, id)
	if err != nil {
		a.errorHandler(w, err)
		return
	}

	post, err := a.P.GetOne(id, 0)
	if err != nil {
		a.errorHandler(w, err)
		return
	}
	a.Infolog.Printf("create comment to post :%s", id)
	a.Write(w, post)
}

func (a Api) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postId := vars["id"]
	commentId := vars["commentId"]
	err := a.P.DeleteComment(postId, commentId)
	if err != nil {
		a.errorHandler(w, err)
		return
	}
	post, err := a.P.GetOne(postId, 0)
	if err != nil {
		a.errorHandler(w, err)
		return
	}
	a.Infolog.Printf("delete comment %s from post %s", commentId, postId)
	a.Write(w, post)
}

func (a Api) MakeVote(w http.ResponseWriter, r *http.Request) {
	usr, ok := r.Context().Value(other.CtxValue).(user.User)
	if !ok {
		a.errorHandler(w, fmt.Errorf("%w : can't get user from context", myerrors.ErrInternalError))
		return
	}
	vars := mux.Vars(r)
	postId := vars["id"]
	var err error
	switch vars["vote"] {
	case "upvote":
		err = a.P.MakeVote(postId, usr.Id, 1)
	case "downvote":
		err = a.P.MakeVote(postId, usr.Id, -1)
	case "unvote":
		err = a.P.MakeVote(postId, usr.Id, 0)
	}
	if err != nil {
		a.errorHandler(w, err)
		return
	}
	post, err := a.P.GetOne(postId, 0)
	if err != nil {
		a.errorHandler(w, err)
		return
	}
	a.Infolog.Printf("make vote to post %s", postId)
	a.Write(w, post)
}
