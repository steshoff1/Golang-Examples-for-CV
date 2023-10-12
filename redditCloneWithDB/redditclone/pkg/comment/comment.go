package comment

import (
	"fmt"
	"net/http"
	"redditclone/pkg/myerrors"
	"redditclone/pkg/other"
	"redditclone/pkg/user"
	"strconv"
	"time"
)

type Comment struct {
	Created string    `json:"created"`
	User    user.User `json:"author" bson:"author"`
	Body    string    `json:"body" bson:"body"`
	Id      string    `json:"id" bson:"id"`
}

func CreateComment(r *http.Request, user user.User, id uint32) (*Comment, error) {
	m, err := other.GetMapFromPost(r)
	r.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("Comment Create: %w", err)
	}

	body := m["comment"]
	if body == "" {
		return nil, fmt.Errorf("%w : empty comment body", myerrors.ErrComplex{
			Errors: []myerrors.ErrSupport{
				{
					Location: "body",
					Param:    "comment",
					Msg:      "is required",
				},
			},
		})
	}

	if len(body) >= 2000 {
		return nil, fmt.Errorf("%w : empty comment body", myerrors.ErrComplex{
			Errors: []myerrors.ErrSupport{
				{
					Location: "body",
					Param:    "comment",
					Msg:      "must be at most 2000 characters long",
					Value:    body,
				},
			},
		})
	}

	comment := &Comment{
		User:    user,
		Id:      strconv.Itoa(int(id)),
		Body:    body,
		Created: time.Now().Local().Format(time.RFC3339),
	}
	return comment, nil
}
