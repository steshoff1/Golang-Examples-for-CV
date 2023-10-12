package votes

type Vote struct {
	UserId string `json:"user" bson:"user"`
	Vote   int    `json:"vote" bson:"vote"`
}

func MakeVote(userId string, vote int) *Vote {
	v := Vote{}
	v.Vote += vote
	v.UserId = userId
	return &v
}
