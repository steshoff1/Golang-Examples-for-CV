package user

type User struct {
	Id   string `json:"id" bson:"id"`
	Log  string `json:"username" bson:"username"`
	Pass string `json:",omitempty" bson:",omitempty"`
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
