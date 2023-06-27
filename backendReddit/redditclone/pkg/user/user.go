package user

type User struct {
	Id   string `json:"id"`
	Log  string `json:"username"`
	Pass string `json:",omitempty"`
}

func (u User) Login() string {
	return u.Log
}

func (u User) ID() string {
	return u.Id
}

func (u User) Password() string {
	return u.Pass
}

type UserInterface interface {
	ID() string
	Login() string
	Password() string
}
