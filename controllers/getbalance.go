package controllers

import (
	jwt "github.com/emmadal/feeti-module/auth"
	status "github.com/emmadal/feeti-module/status"
	"github.com/emmadal/feeti-wallet/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
)

// GetBalanceByUser get balance by user
func GetBalanceByUser(c *gin.Context) {
	// get user_id from url
	ctxUserID := c.Param("userID")
	if ctxUserID == "" {
		status.HandleError(c, http.StatusBadRequest, "userID is required", nil)
		return
	}

	// parse userID and verify user identity with context data
	userID := uuid.MustParse(ctxUserID)
	if userID != jwt.GetUserIDFromGin(c) {
		status.HandleError(c, http.StatusForbidden, "Forbidden request", nil)
		return
	}

	wallet := models.Wallet{UserID: userID}

	// get wallet balance
	wl, err := wallet.GetBalance()
	if err != nil {
		status.HandleError(c, http.StatusInternalServerError, "failed to get wallet", err)
		return
	}
	response := models.WalletResponse{
		ID:       wl.ID,
		Currency: wl.Currency,
		Balance:  wl.Balance,
	}

	// record wallet log
	go func() {
		walletLog := models.WalletLog{
			UserID:         userID,
			WalletID:       wl.ID,
			Activity:       "GET_BALANCE",
			OldBalance:     wl.Balance,
			NewBalance:     wl.Balance,
			ActivityAmount: 0,
			Currency:       wl.Currency,
			Metadata:       `{"source": "get_balance"}`,
		}
		if err := walletLog.CreateWalletLog(); err != nil {
			log.Printf("Error creating wallet log: %v\n", err)
		}
	}()

	status.HandleSuccessData(c, "balance retrieved successfully", response)
}
