package controllers

import (
	"fmt"
	jwt "github.com/emmadal/feeti-module/auth"
	status "github.com/emmadal/feeti-module/status"
	"github.com/emmadal/feeti-wallet/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

// WithdrawWallet processes a wallet withdraw request
func WithdrawWallet(c *gin.Context) {
	var body models.WithdrawRequest

	// parse request body
	if err := c.ShouldBindJSON(&body); err != nil {
		status.HandleError(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	// verify user identity with context data
	if body.UserID != jwt.GetUserIDFromGin(c) {
		status.HandleError(c, http.StatusForbidden, "Forbidden request", nil)
		return
	}

	// get active balance
	w := models.Wallet{UserID: body.UserID, ID: body.WalletID}

	// Check if the wallet is locked
	if w.WalletIsLocked() {
		status.HandleError(c, http.StatusLocked, "account locked", nil)
		return
	}

	// get balance
	balance, err := w.GetBalance()
	if err != nil {
		status.HandleError(c, http.StatusInternalServerError, "failed to get wallet balance", err)
		return
	}

	// validate withdrawal amount
	if body.Amount > balance.Balance {
		status.HandleError(c, http.StatusUnauthorized, "insufficient balance", nil)
		return
	}

	// withdraw wallet
	w.Balance = balance.Balance
	wallet, err := w.WithdrawWallet(body.Amount)
	if err != nil {
		status.HandleError(c, http.StatusInternalServerError, "failed to withdraw wallet", err)
		return
	}

	// publish event to NATS
	//go func(c context.Context) {
	//	data, _ := json.Marshal(wallet)
	//	helpers.JStream.Publish(c, "wallet.transactions", data)
	//}(c.Request.Context())

	// Create a withdrawal log
	go func() {
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
	}()

	// return response
	status.HandleSuccessData(c, "withdrawal successful", models.WalletResponse{
		ID:       wallet.ID,
		Currency: wallet.Currency,
		Balance:  wallet.Balance,
	})
}
