package controllers

import (
	"context"
	"fmt"
	"github.com/emmadal/feeti-wallet/helpers"
	"github.com/emmadal/feeti-wallet/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type WithdrawRequest struct {
	Amount   int64 `json:"amount" binding:"required,numeric,gt=0,min=100,max=2000000"`
	UserID   int64 `json:"user_id" binding:"required,gt=0,numeric"`
	WalletID int64 `json:"wallet_id" binding:"required,gt=0,numeric"`
}

// WithdrawWallet processes a wallet withdraw request
func WithdrawWallet(c *gin.Context) {
	var body WithdrawRequest

	// parse request body
	if err := c.ShouldBindJSON(&body); err != nil {
		helpers.HandleError(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	// get active balance
	w := models.Wallet{UserID: body.UserID, ID: body.WalletID}

	// Check if the wallet is locked
	if w.WalletIsLocked() {
		helpers.HandleError(c, http.StatusLocked, "account locked", nil)
		return
	}

	// get balance
	balance, err := w.GetBalance()
	if err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, "failed to get wallet balance", err)
		return
	}

	// validate withdrawal amount
	if body.Amount > balance.Balance {
		helpers.HandleError(c, http.StatusUnauthorized, "insufficient balance", nil)
		return
	}

	// withdraw wallet
	w.Balance = balance.Balance
	wallet, err := w.WithdrawWallet(body.Amount)
	if err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, "failed to withdraw wallet", err)
		return
	}

	// publish event to NATS
	//go func(c context.Context) {
	//	data, _ := json.Marshal(wallet)
	//	helpers.JStream.Publish(c, "wallet.transactions", data)
	//}(c.Request.Context())

	// Create a withdrawal log
	go func(c context.Context) {
		walletLog := models.WalletLog{
			UserID:         body.UserID,
			WalletID:       wallet.ID,
			Activity:       "WITHDRAWAL",
			OldBalance:     balance.Balance,
			NewBalance:     wallet.Balance,
			ActivityAmount: body.Amount,
			Currency:       wallet.Currency,
			Metadata:       `{"source": "withdrawal"}`,
		}
		if err := walletLog.CreateWalletLog(); err != nil {
			fmt.Printf("Failed to create withdrawal log: %v\n", err)
		}
	}(c.Request.Context())

	// return response
	helpers.HandleSuccessData(c, "withdrawal successful", models.WalletResponse{
		ID:       wallet.ID,
		Currency: wallet.Currency,
		Balance:  wallet.Balance,
	})
}
