package main

import (
	"net/http"
	"redditclone/pkg/handlers"
	"redditclone/pkg/middleware"

	"github.com/gorilla/mux"
)

func main() {
	a := handlers.Api{}
	err := a.Init()
	if err != nil {
		panic(err)
	}
	defer a.Destroy()
	mmux := mux.NewRouter()
	authMux := mux.NewRouter()
	mmux.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	mmux.Handle("/", http.FileServer(http.Dir("./static/html/")))
	mmux.HandleFunc("/api/register", a.RegHand).Methods("POST")
	mmux.HandleFunc("/api/login", a.SignHand).Methods("POST")
	mmux.HandleFunc("/api/posts/", a.Posts).Methods("GET")
	mmux.HandleFunc("/api/posts/{category:music|funny|videos|programming|news|fashion}", a.GetCategory).Methods("GET")
	mmux.HandleFunc("/api/post/{id:[a-zA-Z0-9]+}", a.OnePost).Methods("GET")
	mmux.HandleFunc("/api/user/{id:[a-zA-Z0-9]+}", a.GetPostsByUser).Methods("GET")
	authMux.HandleFunc("/api/post/{id:[a-zA-Z0-9]+}", a.DeletePost).Methods("DELETE")
	authMux.HandleFunc("/api/post/{id:[a-zA-Z0-9]+}", a.CreateComment).Methods("POST")
	authMux.HandleFunc("/api/posts", a.PostCreate).Methods("POST")
	authMux.HandleFunc("/api/post/{id:[a-zA-Z0-9]+}/{commentId:[0-9]+}", a.DeleteComment).Methods("DELETE")
	authMux.HandleFunc("/api/post/{id:[a-zA-Z0-9]+}/{vote:unvote|downvote|upvote}", a.MakeVote).Methods("GET")
	mmux.UseEncodedPath().NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.Infolog.Println(`redirected to "/"`)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	authHand := middleware.Auth(&a, authMux)
	mmux.PathPrefix("/api/").Handler(authHand)

	r := middleware.AccessLog(&a, mmux)
	r = middleware.Panic(&a, r)

	err = http.ListenAndServe("localhost:8080", r)
	if err != nil {
		a.Destroy()
		a.Errorlog.Fatalf("cant start server : %s", err)
	}
}
