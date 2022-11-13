package domain

type UserDTO struct {
	ID       int    `db:"id"`
	Login    string `db:"login"`
	Password string `db:"password"`
}

type TokenData struct {
	Token string `json:"token"`
	// todo: add expires field
}
