package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/emmadal/feeti-wallet/models"
	"github.com/nats-io/nats.go"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	nc   *nats.Conn
	once sync.Once
)

// NatsConfig holds the configuration options for NATS
type NatsConfig struct {
	URL           string
	MaxReconnects int
	ReconnectWait time.Duration
	Replicas      int
}

// defaultNatsConfig returns default configuration for NATS
func defaultNatsConfig() NatsConfig {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}
	return NatsConfig{
		URL:           natsURL,
		MaxReconnects: 5,
		ReconnectWait: 5 * time.Second,
		Replicas:      1,
	}
}

// NatsConnect initializes the NATS connection
func NatsConnect() error {
	var connectErr error

	once.Do(func() {
		// Load configuration from environment
		config := defaultNatsConfig()

		// Connect to NATS
		nc, connectErr = nats.Connect(
			config.URL,
			nats.RetryOnFailedConnect(true),
			nats.MaxReconnects(config.MaxReconnects),
			nats.ReconnectWait(config.ReconnectWait),
			nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
				log.Println("NATS disconnected")
			}),
			nats.ReconnectHandler(func(nc *nats.Conn) {
				log.Printf("NATS reconnection attempt")
			}),
			nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
				fmt.Printf("Nats error: %v\n", err)
			}),
			nats.ClosedHandler(func(nc *nats.Conn) {
				log.Println("NATS connection closed")
			}),
		)

		if connectErr != nil {
			fmt.Printf("Failed to connect to NATS: %v\n", connectErr)
			return
		}
		fmt.Println("Successfully connected to NATS")

		// Only start subscribers if everything is set up correctly
		if connectErr == nil {
			subscribeToCreateWallet()
			subscribeToDisableWallet()
			subscribeToGetBalance()
		}
	})

	return connectErr
}

// DrainNatsConnection drains and closes the NATS connection
func DrainNatsConnection() error {
	if nc == nil {
		return nil
	}
	log.Println("Draining NATS connection...")
	return nc.Drain()
}

// ResponsePayload represents the standard response structure
type ResponsePayload struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// subscribeToCreateWallet creates a wallet when a message is received
func subscribeToCreateWallet() {
	// Subscribe to the "wallet.create" subject
	sub, err := nc.Subscribe("wallet.create", func(msg *nats.Msg) {
		startTime := time.Now()
		fmt.Printf("Received message [%s] on subject %s\n", string(msg.Data), msg.Subject)

		// Parse the message payload
		userId := string(msg.Data)
		id, err := strconv.ParseInt(userId, 10, 64)
		if err != nil {
			fmt.Printf("Invalid user ID: %s\n", userId)
			sendResponse(msg, ResponsePayload{
				Success: false,
				Error:   "Invalid user ID: must be a number",
			})
			return
		}

		// Create a wallet with a retry mechanism
		wallet := models.Wallet{UserID: id}
		newWallet, err := wallet.CreateWallet()
		fmt.Println("Wallet created: ", *newWallet)
		if err != nil {
			log.Printf("Failed to create wallet for user id [%d]: %v\n", id, err)
			sendResponse(msg, ResponsePayload{
				Success: false,
				Error:   fmt.Sprintf("Failed to create wallet for user id [%d]: %v", id, err),
			})
			return
		}
		fmt.Printf("Wallet for user id [%d] created successfully in  %v\n", id, time.Since(startTime))

		// Send success response
		sendResponse(msg, ResponsePayload{
			Success: true,
			Data:    newWallet,
		})
	})
	if err != nil {
		fmt.Printf("Failed to subscribe to subject: %v\n", err)
		return
	}

	// Keep subscription active - don't auto-unsubscribe
	if err := sub.SetPendingLimits(-1, -1); err != nil {
		fmt.Printf("Failed to set pending limits: %v\n", err)
	}
}

func subscribeToDisableWallet() {
	// Subscribe to the "wallet.disable" subject
	sub, err := nc.Subscribe("wallet.disable", func(msg *nats.Msg) {
		startTime := time.Now()
		fmt.Printf("Received message [%s] on subject %s\n", string(msg.Data), msg.Subject)

		// Parse the message payload
		userId := string(msg.Data)
		id, err := strconv.ParseInt(userId, 10, 64)
		if err != nil {
			fmt.Printf("Invalid user ID: %s\n", userId)
			sendResponse(msg, ResponsePayload{
				Success: false,
				Error:   "Invalid user ID: must be a number",
			})
			return
		}

		// Create a wallet with a retry mechanism
		wallet := models.Wallet{UserID: id}
		err = wallet.DeleteWallet()
		if err != nil {
			log.Printf("Failed to disable wallet for user id [%d]: %v\n", id, err)
			sendResponse(msg, ResponsePayload{
				Success: false,
				Error:   fmt.Sprintf("Failed to disable wallet for user id [%d]: %v", id, err),
			})
			return
		}
		fmt.Printf("Wallet for user id [%d] disabled successfully in  %v\n", id, time.Since(startTime))

		// Send success response
		sendResponse(msg, ResponsePayload{
			Success: true,
			Data:    nil,
		})
	})

	if err != nil {
		fmt.Printf("Failed to subscribe to subject: %v\n", err)
		return
	}

	// Keep subscription active - don't auto-unsubscribe
	if err := sub.SetPendingLimits(-1, -1); err != nil {
		fmt.Printf("Failed to set pending limits: %v\n", err)
	}
}

func subscribeToGetBalance() {
	// Subscribe to the "wallet.balance" subject
	sub, err := nc.Subscribe("wallet.balance", func(msg *nats.Msg) {
		startTime := time.Now()
		fmt.Printf("Received message [%s] on subject %s\n", string(msg.Data), msg.Subject)

		// Parse the message payload
		userId := string(msg.Data)
		id, err := strconv.ParseInt(userId, 10, 64)
		if err != nil {
			fmt.Printf("Invalid user ID: %s\n", userId)
			sendResponse(msg, ResponsePayload{
				Success: false,
				Error:   "Invalid user ID: must be a number",
			})
			return
		}

		// Create a wallet with a retry mechanism
		wallet := models.Wallet{UserID: id}
		balance, err := wallet.GetBalance()
		if err != nil {
			log.Printf("Failed to get balance for user id [%d]: %v\n", id, err)
			sendResponse(msg, ResponsePayload{
				Success: false,
				Error:   fmt.Sprintf("Failed to get balance for user id [%d]: %v", id, err),
			})
			return
		}
		fmt.Printf("Balance for user id [%d] retrieved successfully in  %v\n", id, time.Since(startTime))

		// Send success response
		sendResponse(msg, ResponsePayload{
			Success: true,
			Data:    balance,
		})
	})

	if err != nil {
		fmt.Printf("Failed to subscribe to subject: %v\n", err)
		return
	}

	// Keep subscription active - don't auto-unsubscribe
	if err := sub.SetPendingLimits(-1, -1); err != nil {
		fmt.Printf("Failed to set pending limits: %v\n", err)
	}
}

// sendResponse sends a structured response to the NATS message
func sendResponse(msg *nats.Msg, payload ResponsePayload) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal response: %v\n", err)
		// Try to send a simplified error if marshaling fails
		simpleErr, _ := json.Marshal(ResponsePayload{Success: false, Error: "Internal server error"})
		_ = msg.Respond(simpleErr)
		return
	}
	if err := msg.Respond(data); err != nil {
		log.Printf("Failed to send response: %v\n", err)
	}
}
