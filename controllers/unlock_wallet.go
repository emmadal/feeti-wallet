package controllers

import (
	"context"
	"github.com/emmadal/feeti-wallet/helpers"
	"github.com/emmadal/feeti-wallet/models"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type UnLockRequest struct {
	UserID   int64 `json:"user_id" binding:"required,gt=0,numeric"`
	WalletID int64 `json:"wallet_id" binding:"required,gt=0,numeric"`
}

// UnLockWalletByUser unlocks the user wallet
func UnLockWalletByUser(c *gin.Context) {
	var body UnLockRequest

	// Parse request body
	if err := c.ShouldBindJSON(&body); err != nil {
		helpers.HandleError(c, http.StatusBadRequest, "invalid request", err)
		return
	}
	w := models.Wallet{ID: body.WalletID, UserID: body.UserID}

	// Check if the wallet is locked
	if !w.WalletIsLocked() {
		helpers.HandleError(c, http.StatusNotFound, "account not locked", nil)
		return
	}

	// Get balance
	wallet, err := w.GetBalance()
	if err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, "failed to get wallet balance", err)
		return
	}

	// Unlock wallet
	if err := w.UnlockWallet(); err != nil {
		helpers.HandleError(c, http.StatusInternalServerError, "failed to unlock wallet", err)
		return
	}

	// Create a withdrawal log
	go func(ctx context.Context) {
		walletLog := models.WalletLog{
			UserID:         body.UserID,
			WalletID:       wallet.ID,
			Activity:       "UNLOCK_WALLET",
			OldBalance:     wallet.Balance,
			NewBalance:     wallet.Balance,
			ActivityAmount: 0,
			Currency:       wallet.Currency,
			Metadata:       `{"source": "unlock_wallet"}`,
		}
		if err := walletLog.CreateWalletLog(); err != nil {
			log.Printf("Error creating wallet log: %v\n", err)
		}
	}(c.Request.Context()) // Pass the context to the goroutine

	helpers.HandleSuccess(c, "wallet unlocked successfully")
}
