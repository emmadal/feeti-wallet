package controllers

import (
	"context"
	status "github.com/emmadal/feeti-module/status"
	"github.com/emmadal/feeti-wallet/helpers"
	"github.com/emmadal/feeti-wallet/models"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

// GetBalanceByUser get balance by user
func GetBalanceByUser(c *gin.Context) {
	// get user_id from url
	userID := c.Param("userID")
	if userID == "" {
		status.HandleError(c, http.StatusBadRequest, "userID is required", nil)
		return
	}

	// check if user_id is a number
	if !helpers.IsNumericRequestID(userID) {
		status.HandleError(c, http.StatusBadRequest, "incorrect parameters", nil)
		return
	}

	// convert userID to int64
	userIDInt64, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		status.HandleError(c, http.StatusBadRequest, "invalid parameters", nil)
		return
	}

	wallet := models.Wallet{UserID: userIDInt64}

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
	go func(c context.Context) {
		walletLog := models.WalletLog{
			UserID:         userIDInt64,
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
	}(c.Request.Context())

	status.HandleSuccessData(c, "balance retrieved successfully", response)
}
