package domain

type UserDTO struct {
	ID       int    `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
}

type TokenData struct {
	Token string `json:"token"`
	// todo: add expires field
}