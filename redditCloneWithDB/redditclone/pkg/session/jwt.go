package session

import (
	"fmt"
	"redditclone/pkg/myerrors"
	"redditclone/pkg/user"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var (
	TokenSecret = []byte("eyJpZCI6NDUwLCJsb2dpbiI6InJ2YXNpbHkiLCJuYW1lIjoiVmFzaWx5IFJvbWFub3YiLCJyb2xlIjoidXNlciJ9")
)

func JwtDecode(token string) (user.User, error) {
	hashSecretGetter := func(token *jwt.Token) (interface{}, error) {
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method.Alg() != "HS256" {
			return nil, fmt.Errorf("%w : bad sign method for JWT", myerrors.ErrBadRequest)
		}
		return TokenSecret, nil
	}
	tok, err := jwt.Parse(token, hashSecretGetter)
	if err != nil || !tok.Valid {
		return user.User{}, fmt.Errorf("%w : bad JWT %s", myerrors.ErrBadRequest, token)
	}

	payload, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return user.User{}, fmt.Errorf("%w : no payload in JWT : %s", myerrors.ErrBadRequest, token)
	}
	value := payload["user"]
	switch m := value.(type) {
	case map[string]interface{}:
		params, err := ValidValue(m, []string{"id", "username"})
		if err != nil {
			return user.User{}, fmt.Errorf("%w : bad JWT : %s", myerrors.ErrBadRequest, err.Error())
		}
		return user.User{
			Id:  params[0],
			Log: params[1],
		}, nil
	default:
		return user.User{}, fmt.Errorf("%w : bad JWT, havent got user map", myerrors.ErrBadRequest)
	}
}

func ValidValue(m map[string]interface{}, key []string) ([]string, error) {
	res := make([]string, len(key))
	for i, v := range key {
		value, ok1 := m[v]
		if !ok1 {
			return nil, fmt.Errorf("havent got %s", key)
		}
		switch ret := value.(type) {
		case string:
			res[i] = ret
		default:
			return nil, fmt.Errorf("invalid map")
		}
	}
	return res, nil
}

func MakeJwt(usr user.UserInterface) (string, error) {

	ju := map[string]string{
		"username": usr.Login(),
		"id":       usr.ID(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": ju,
		"iat":  strconv.Itoa(int(time.Now().Unix())),
	})
	tokenString, err := token.SignedString(TokenSecret)

	if err != nil {
		return "", fmt.Errorf("%w : makeJwt : %s", myerrors.ErrInternalError, err)
	}
	return tokenString, nil
}
