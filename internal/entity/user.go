package entity

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	ApiKey   string `json:"api_key"`
	Role     string `json:"type" validate:"required,oneof=admin staff"`
}
