package session

import (
	"fmt"
	"net/http"
	"redditclone/pkg/myerrors"
	"redditclone/pkg/other"
	"redditclone/pkg/user"
	"strconv"
	"sync"
)

type Users struct {
	sessions map[string]user.UserInterface
	mu       *sync.RWMutex
	ind      uint32
}

func (u *Users) Init() {
	u.sessions = make(map[string]user.UserInterface)
	u.mu = &sync.RWMutex{}
	u.ind = 0
}

func (u *Users) CreateUser(r *http.Request) (string, error) {
	m, err := other.GetMapFromPost(r)
	r.Body.Close()
	if err != nil {
		return "", fmt.Errorf("%w : sign up : %s", myerrors.ErrInternalError, err.Error())
	}

	username := m["username"]
	pass := m["password"]
	if pass == "" || username == "" {
		return "", fmt.Errorf("%w : password or username dont not exist", myerrors.ErrBadRequest)
	}

	u.mu.Lock()
	defer u.mu.Unlock()
	if _, ok := u.sessions[username]; ok {
		return "", myerrors.ErrComplex{
			Errors: []myerrors.ErrSupport{
				{
					Location: "body",
					Value:    username,
					Param:    "username",
					Msg:      "already exists",
				},
			},
		}
	}

	usr := user.User{
		Id:   strconv.Itoa(int(u.ind)),
		Log:  username,
		Pass: pass,
	}

	token, err := MakeJwt(usr)
	if err != nil {
		return "", fmt.Errorf("%w : password or username dont not exist", myerrors.ErrBadRequest)
	}
	u.ind++

	u.sessions[usr.Log] = usr
	return token, nil
}

func (u *Users) Find(username string) (user.UserInterface, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	if usr, ok := u.sessions[username]; ok {
		return usr, nil
	}
	return user.User{}, fmt.Errorf("%w : user not found", myerrors.ErrComplex{Msg: "user not found", StatusCode: http.StatusUnauthorized})
}
