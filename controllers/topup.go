package controllers

import (
	"fmt"
	jwt "github.com/emmadal/feeti-module/auth"
	status "github.com/emmadal/feeti-module/status"
	"github.com/emmadal/feeti-wallet/models"

	"github.com/gin-gonic/gin"
	"net/http"
)

// TopupWallet processes a wallet topup request
func TopupWallet(c *gin.Context) {
	var body models.Request

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
	balance, err := w.GetBalance()
	if err != nil {
		status.HandleError(c, http.StatusInternalServerError, "failed to get wallet balance", err)
		return
	}

	// topup wallet
	wallet, err := w.RechargeWallet(body.Amount)
	if err != nil {
		status.HandleError(c, http.StatusInternalServerError, "failed to topup wallet", err)
		return
	}

	// Create topup log
	go func() {
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
	}()

	// return success response
	status.HandleSuccessData(c, "wallet topup successful", models.WalletResponse{
		ID:       wallet.ID,
		Currency: wallet.Currency,
		Balance:  wallet.Balance,
	})
}
