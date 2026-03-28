package entity

type User struct {
	ID       int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"type" validate:"required,oneof=admin staff"`
}
