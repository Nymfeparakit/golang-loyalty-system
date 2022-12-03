package handlers

type registrationInput struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginInput struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type BalanceWithdrawalInput struct {
	OrderNumber string  `json:"order" binding:"required"`
	Sum         float32 `json:"sum" binding:"required,numeric,gt=0"`
}
