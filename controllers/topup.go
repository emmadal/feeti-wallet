package controllers

import (
	"context"
	"fmt"
	"github.com/emmadal/feeti-wallet/helpers"
	"github.com/emmadal/feeti-wallet/models"

	"github.com/gin-gonic/gin"
	"net/http"
)

type Request struct {
	Amount   int64 `json:"amount" binding:"required,numeric,gt=0,min=100,max=2000000"`
	UserID   int64 `json:"user_id" binding:"required,gt=0,numeric"`
	WalletID int64 `json:"wallet_id" binding:"required,gt=0,numeric"`
}

// TopupWallet processes a wallet topup request
func TopupWallet(c *gin.Context) {
	var body Request

	// parse request body
	if err := c.ShouldBindJSON(&body); err != nil {
		helpers.HandleError(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	// get active balance
	w := models.Wallet{UserID: body.UserID, ID: body.WalletID}
	balance, err := w.GetBalance()
	if err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, "failed to get wallet balance", err)
		return
	}

	// topup wallet
	wallet, err := w.RechargeWallet(body.Amount)
	if err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, "failed to topup wallet", err)
		return
	}

	// Create topup log
	go func(c context.Context) {
		walletLog := models.WalletLog{
			UserID:         body.UserID,
			WalletID:       wallet.ID,
			Activity:       "TOPUP_WALLET",
			OldBalance:     balance.Balance,
			NewBalance:     wallet.Balance,
			ActivityAmount: body.Amount,
			Currency:       wallet.Currency,
			Metadata:       `{"source": "topup"}`,
		}
		if err := walletLog.CreateWalletLog(); err != nil {
			fmt.Printf("Failed to log topup activity for user %d: %v\n", body.UserID, err)
		}
	}(c.Request.Context())

	// return success response
	helpers.HandleSuccessData(c, "wallet topup successful", models.WalletResponse{
		ID:       wallet.ID,
		Currency: wallet.Currency,
		Balance:  wallet.Balance,
	})
}
