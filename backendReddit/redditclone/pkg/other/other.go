package other

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"redditclone/pkg/myerrors"
)

type Ctxval string

var CtxValue Ctxval = "token"

func GetMapFromPost(r *http.Request) (map[string]string, error) {
	respBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("%w : cant get map from post : %s", myerrors.ErrInternalError, err.Error())
	}

	m := make(map[string]string)
	err = json.Unmarshal(respBody, &m)
	if err != nil {
		return nil, fmt.Errorf("%w : cant get map from post : %s", myerrors.ErrInternalError, err.Error())
	}
	return m, nil
}
