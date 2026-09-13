package model

type BalanceResponse struct {
	Current  float32 `json:"current"`
	Withdraw float32 `json:"withdrawn"`
}
