package models

type User struct {
	ID        int    `json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Contact   string `json:"contact"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Status    string `json:"status"`
}
