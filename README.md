# Feeti Wallet

Feeti Wallet is a simple wallet system written in Go. The system currently supports the following features:

- Creating a new wallet
- Getting a wallet balance
- Depositing money in a wallet
- Withdrawing money from a wallet
- Locking a wallet
- Unlocking a wallet

## Setup

1. Install Go on your machine
2. Clone this repository
3. Run `go get` to get all the dependencies
4. Run `go run main.go` to start the server

## API Endpoints

- `GET /v1/api/balance/:userID`: Get the balance of a wallet
- `POST /v1/api/deposit`: Deposit money to a wallet
- `POST /v1/api/withdraw`: Withdraw money from a wallet
- `POST /v1/api/lock`: Lock a wallet
- `POST /v1/api/unlock`: Unlock a wallet

## NATS

The system uses NATS as a message broker. The system listens to the following subjects:

- `wallet.create`: Create a new wallet
- `wallet.deposit`: Deposit money to a wallet
- `wallet.withdraw`: Withdraw money from a wallet
- `wallet.lock`: Lock a wallet
- `wallet.unlock`: Unlock a wallet

## Environment Variables

The system uses the following environment variables:

- `GIN_MODE`: The mode of the server (release or debug)
- `PORT`: The port of the server
- `NATS_URL`: The URL of the NATS server
- `NATS_CLUSTER_ID`: The cluster ID of the NATS server
- `NATS_CLIENT_ID`: The client ID of the NATS server

## Running Tests

To run the tests, run the following command:

```bash
go test -v ./...
```