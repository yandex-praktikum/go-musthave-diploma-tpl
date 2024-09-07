package register

type RequestBody struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
