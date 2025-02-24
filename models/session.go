package models

type Session struct {
	Id    string `json:"id"`
	User  string `json:"user"`
	Email string `json:"email"`
	App   string `json:"app"`
}
