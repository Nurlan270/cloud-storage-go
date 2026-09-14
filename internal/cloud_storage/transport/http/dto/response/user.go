package response

type RegisterUser struct {
	Username string `json:"username"`
}

type LoginUser struct {
	Username string `json:"username"`
}

type UserMe struct {
	Username string `json:"username"`
}
