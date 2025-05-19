package controllers

import (
	"context"
	status "github.com/emmadal/feeti-module/status"
	"github.com/emmadal/feeti-wallet/models"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// UnLockWalletByUser unlocks the user wallet
func UnLockWalletByUser(c *gin.Context) {
	var body models.UnLockRequest

	// Parse request body
	if err := c.ShouldBindJSON(&body); err != nil {
		status.HandleError(c, http.StatusBadRequest, "invalid request", err)
		return
	}
	w := models.Wallet{ID: body.WalletID, UserID: body.UserID}

	// Check if the wallet is locked
	if !w.WalletIsLocked() {
		status.HandleError(c, http.StatusNotFound, "account not locked", nil)
		return
	}

	// Get balance
	wallet, err := w.GetBalance()
	if err != nil {
		status.HandleError(c, http.StatusInternalServerError, "failed to get wallet balance", err)
		return
	}

	// Unlock wallet
	if err := w.UnlockWallet(); err != nil {
		status.HandleError(c, http.StatusInternalServerError, "failed to unlock wallet", err)
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

	status.HandleSuccess(c, "wallet unlocked successfully")
}
