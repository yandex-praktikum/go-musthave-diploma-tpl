package dto

type UserCredential struct {
	Login    string `json: "login"`
	Password string `json: "password"`
}

type UserData struct {
	Uuid     string `json: "uuid"`
	Login    string `json: "login"`
	Password string `json: "password"`
}

func (c *UserCredential) ToUserData(uuid string) *UserData {
	return &UserData{
		Uuid:     uuid,
		Login:    c.Login,
		Password: c.Password,
	}
}
