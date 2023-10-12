package session

import (
	"database/sql"
	"fmt"
	"net/http"
	"redditclone/pkg/myerrors"
	"redditclone/pkg/other"
	"redditclone/pkg/user"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type UsersRepoInterface interface {
	Init() error
	AddSession(usr user.UserInterface) (string, error)
	Valid(sessionID string) (bool, error)
	Find(username string) (user.UserInterface, error)
	CreateUser(r *http.Request) (user.UserInterface, error)
	Destroy()
}

type SqlUsers struct {
	DB *sql.DB
}

func (u *SqlUsers) Init() error {
	dsn := "root:love@tcp(localhost:3306)/golang?"
	dsn += "charset=utf8"
	dsn += "&interpolateParams=true"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("users Init : %w", err)
	}

	db.SetMaxOpenConns(100)

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("users Init : %w", err)
	}
	u.DB = db
	return nil
}

func (u *SqlUsers) Destroy() {
	u.DB.Close()
}

func (u SqlUsers) CreateUser(r *http.Request) (user.UserInterface, error) {
	m, err := other.GetMapFromPost(r)
	r.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("%w : sign up : %s", myerrors.ErrInternalError, err.Error())
	}

	username := m["username"]
	pass := m["password"]
	if pass == "" || username == "" {
		return nil, fmt.Errorf("%w : password or username dont not exist", myerrors.ErrBadRequest)
	}

	if _, err := u.Find(username); err == nil {
		return nil, myerrors.ErrComplex{
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

	result, err := u.DB.Exec(
		"INSERT INTO users (`username`, `password`) VALUES (?, ?)",
		username,
		pass,
	)
	if err != nil {
		return nil, fmt.Errorf("%w : CreateUser : DB error : %s", myerrors.ErrInternalError, err.Error())
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("%w : CreateUser : DB error : %s", myerrors.ErrInternalError, err.Error())
	}

	return &user.User{
		Id:   strconv.Itoa(int(id)),
		Pass: pass,
		Log:  username,
	}, nil
}

func (u SqlUsers) Find(username string) (user.UserInterface, error) {
	usr := user.User{}
	var id int
	err := u.DB.
		QueryRow("SELECT id, username, password FROM users WHERE username = ?", username).
		Scan(&id, &usr.Log, &usr.Pass)
	if err != nil {
		return user.User{}, fmt.Errorf("%w : user not found", myerrors.ErrComplex{Msg: "user not found", StatusCode: http.StatusUnauthorized})
	}
	usr.Id = strconv.Itoa(id)
	return usr, nil
}

func (s SqlUsers) Valid(sessionID string) (bool, error) {
	var exp int64
	err := s.DB.
		QueryRow("SELECT exp FROM sessions WHERE sessionID = ?", sessionID).
		Scan(&exp)
	if err != nil {
		return false, err
	}
	if time.Duration(exp) <= time.Duration(time.Now().Unix()) {
		_, err = s.DB.Exec(
			"DELETE FROM sessions WHERE sessionID = ?",
			sessionID,
		)
		if err != nil {
			return false, fmt.Errorf("%w : valid session : %s", myerrors.ErrInternalError, err)
		}
		return false, nil
	}
	return true, nil
}

func (s SqlUsers) AddSession(usr user.UserInterface) (string, error) {
	expiration := time.Now().AddDate(0, 0, 7)
	token, err := MakeJwt(user.User{
		Id:   usr.ID(),
		Log:  usr.Login(),
		Pass: usr.Password(),
	})
	if err != nil {
		return "", fmt.Errorf("%w : create user", err)
	}

	_, err = s.DB.Exec(
		"INSERT INTO sessions (`exp`, `sessionID`) VALUES (?, ?)",
		expiration.Unix(),
		token,
	)
	if err != nil {
		return "", fmt.Errorf("%w : AddSession : %s", myerrors.ErrInternalError, err.Error())
	}
	return token, nil
}
