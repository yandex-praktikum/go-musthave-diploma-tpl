package models

type UserSearchFilter struct {
	Login string `json:"login" query:"login"`
}
