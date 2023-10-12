package middleware

import (
	"context"
	"net/http"
	"redditclone/pkg/handlers"
	"redditclone/pkg/myerrors"
	"redditclone/pkg/other"
	"redditclone/pkg/session"
	"time"
)

func Auth(a *handlers.Api, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			a.Errorlog.Println("empty token")
			http.Error(w, myerrors.ErrComplex{Msg: "user not found"}.Error(), http.StatusUnauthorized)
			return
		}
		token = token[7:]

		ok, err := a.U.Valid(token)
		if err != nil {
			a.Errorlog.Printf("cant validate user : %s\n", err)
			http.Error(w, myerrors.ErrInternalError.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			a.Infolog.Println("session is expired or doesnt exist")
			http.Error(w, myerrors.ErrComplex{Msg: "You need to login"}.Error(), http.StatusUnauthorized)
			return
		}

		usr, err := session.JwtDecode(token)
		if err != nil {
			a.Errorlog.Println("invalid token")
			http.Error(w, myerrors.ErrComplex{Msg: "user not found"}.Error(), http.StatusUnauthorized)
			return
		}
		ctx := r.Context()
		next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, other.CtxValue, usr)))
	})
}

func AccessLog(a *handlers.Api, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		a.Infolog.Printf(`New "%s" request with remote addr %s, url : %s, time : %d`, r.Method, r.RemoteAddr, r.URL.Path, time.Since(start).Milliseconds())
	})
}

func Panic(a *handlers.Api, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				a.Errorlog.Println("recovered", err)
				http.Error(w, myerrors.ErrInternalError.Error(), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
