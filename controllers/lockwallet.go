package controllers

import (
	jwt "github.com/emmadal/feeti-module/auth"
	status "github.com/emmadal/feeti-module/status"
	"github.com/emmadal/feeti-wallet/models"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// LockWalletByUser locks a wallet by user
func LockWalletByUser(c *gin.Context) {
	var body models.LockRequest

	// Parse request body
	if err := c.ShouldBindJSON(&body); err != nil {
		status.HandleError(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	// verify user identity with context data
	if body.UserID != jwt.GetUserIDFromGin(c) {
		status.HandleError(c, http.StatusForbidden, "Forbidden request", nil)
		return
	}

	w := models.Wallet{ID: body.WalletID, UserID: body.UserID}

	// Check if the wallet is locked
	if w.WalletIsLocked() {
		status.HandleError(c, http.StatusLocked, "account already locked", nil)
		return
	}

	// Get balance
	wallet, err := w.GetBalance()
	if err != nil {
		status.HandleError(c, http.StatusInternalServerError, "failed to get wallet balance", err)
		return
	}

	// Lock wallet
	if err := w.LockWallet(); err != nil {
		status.HandleError(c, http.StatusInternalServerError, "failed to lock wallet", err)
		return
	}

	// Create a withdrawal log
	go func() {
		walletLog := models.WalletLog{
			UserID:         body.UserID,
			WalletID:       wallet.ID,
			Activity:       "LOCK_WALLET",
			OldBalance:     wallet.Balance,
			NewBalance:     wallet.Balance,
			ActivityAmount: 0,
			Currency:       wallet.Currency,
			Metadata:       `{"source": "lock_wallet"}`,
		}
		if err := walletLog.CreateWalletLog(); err != nil {
			log.Printf("Error creating wallet log: %v\n", err)
		}
	}() // Pass the context to the goroutine

	status.HandleSuccess(c, "wallet locked successfully")
}
