package models

import (
	"context"
	"fmt"
	"time"
)

// Wallet is the struct for a wallet
type Wallet struct {
	ID        int64     `json:"id" db:"id,omitempty"`
	UserID    int64     `json:"user_id" db:"user_id" binding:"required,number,gt=0"`
	Balance   int64     `json:"balance" db:"balance"`
	Currency  string    `json:"currency" db:"currency" binding:"alpha,oneof=XAF"`
	Locked    bool      `json:"locked" db:"locked"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at,omitempty"`
}

// WalletLog is the struct for a wallet log
type WalletLog struct {
	ID             int64     `json:"id" db:"id,omitempty"`
	UserID         int64     `json:"user_id" db:"user_id"`
	WalletID       int64     `json:"wallet_id" db:"wallet_id"`
	Activity       string    `json:"activity" db:"activity"`
	OldBalance     int64     `json:"old_balance" db:"old_balance"`
	NewBalance     int64     `json:"new_balance" db:"new_balance"`
	ActivityAmount int64     `json:"activity_amount" db:"activity_amount"`
	Currency       string    `json:"currency" db:"currency"`
	Metadata       string    `json:"metadata" db:"metadata"`
	CreatedAt      time.Time `json:"created_at" db:"created_at,omitempty"`
}

// WalletResponse is the struct for a wallet response
type WalletResponse struct {
	ID       int64  `json:"id"`
	Currency string `json:"currency"`
	Balance  int64  `json:"balance"`
}

// CreateWallet creates a new wallet
func (w *Wallet) CreateWallet() (*Wallet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tx, err := DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var newWallet Wallet

	err = tx.QueryRow(
		ctx,
		`INSERT INTO wallets(user_id) VALUES ($1) RETURNING id, balance, currency`,
		w.UserID,
	).Scan(
		&newWallet.ID,
		&newWallet.Balance,
		&newWallet.Currency,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &newWallet, nil
}

// LockWallet locks a wallet
func (w *Wallet) LockWallet() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tx, err := DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx) // no-op if already committed
	}()

	// Lock wallet
	_, err = tx.Exec(
		ctx,
		`UPDATE wallets SET locked = true WHERE user_id = $1 AND is_active = true`,
		w.UserID,
	)
	if err != nil {
		return err
	}

	// Then commit transaction
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// UnlockWallet unlocks a wallet
func (w *Wallet) UnlockWallet() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tx, err := DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx) // no-op if already committed
	}()

	// Unlock wallet
	_, err = tx.Exec(
		ctx,
		`UPDATE wallets SET locked = false WHERE user_id = $1 AND id = $2 AND is_active = true`,
		w.UserID,
		w.ID,
	)
	if err != nil {
		return err
	}

	// Then commit transaction
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

// GetBalance gets a wallet balance
func (w *Wallet) GetBalance() (*Wallet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	wallet := &Wallet{}

	err := DB.QueryRow(
		ctx,
		`SELECT id, balance, currency FROM wallets WHERE user_id = $1 AND is_active = true`,
		w.UserID,
	).Scan(
		&wallet.ID,
		&wallet.Balance,
		&wallet.Currency,
	)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

// RechargeWallet recharges a wallet
func (w *Wallet) RechargeWallet(amount int64) (*Wallet, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("invalid amount")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tx, err := DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx) // no-op if already committed
	}()

	// Recharge wallet
	var wallet Wallet
	err = tx.QueryRow(
		ctx,
		`UPDATE wallets SET balance = balance + $1 WHERE user_id = $2 AND id = $3 AND is_active = true RETURNING id, balance, currency`, amount, w.UserID, w.ID,
	).Scan(
		&wallet.ID,
		&wallet.Balance,
		&wallet.Currency,
	)
	if err != nil {
		return nil, err
	}

	// Then commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &wallet, nil
}

// WithdrawWallet withdraws from a wallet
func (w *Wallet) WithdrawWallet(amount int64) (*Wallet, error) {
	if amount > w.Balance {
		return nil, fmt.Errorf("insufficient funds")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tx, err := DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx) // no-op if already committed
	}()

	// Withdraw wallet
	var wallet Wallet
	err = tx.QueryRow(
		ctx,
		`UPDATE wallets SET balance = balance - $1 WHERE user_id = $2 AND id = $3 AND is_active = true AND locked = false RETURNING id, balance, currency`,
		amount,
		w.UserID,
		w.ID,
	).Scan(
		&wallet.ID,
		&wallet.Balance,
		&wallet.Currency,
	)
	if err != nil {
		return nil, err
	}

	// Then commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &wallet, nil
}

// WalletIsLocked checks if a wallet is locked
func (w *Wallet) WalletIsLocked() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var locked bool
	err := DB.QueryRow(
		ctx,
		`SELECT locked FROM wallets WHERE user_id = $1 AND id = $2 AND is_active = true`,
		w.UserID,
		w.ID,
	).Scan(
		&locked,
	)
	if err != nil {
		return false
	}
	return locked
}

// DeleteWallet deletes a wallet
func (w *Wallet) DeleteWallet() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tx, err := DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx) // no-op if already committed
	}()

	// Delete wallet
	_, err = tx.Exec(
		ctx,
		`UPDATE wallets SET is_active = false, locked = false WHERE user_id = $1`,
		w.UserID,
	)
	if err != nil {
		return err
	}

	// Then commit transaction
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

// CreateWalletLog creates a new wallet log
func (wl *WalletLog) CreateWalletLog() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tx, err := DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx) // no-op if already committed
	}()

	// Create wallet log
	_, err = tx.Exec(
		ctx,
		`INSERT INTO wallet_logs (user_id, wallet_id, activity, old_balance, new_balance, activity_amount, currency, metadata) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		wl.UserID,
		wl.WalletID,
		wl.Activity,
		wl.OldBalance,
		wl.NewBalance,
		wl.ActivityAmount,
		wl.Currency,
		wl.Metadata,
	)
	if err != nil {
		return err
	}

	// Then commit transaction
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

// GetWalletLogs gets a wallet logs
func (wl *WalletLog) GetWalletLogs() ([]WalletLog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rows, err := DB.Query(
		ctx,
		`SELECT id, user_id, wallet_id, activity, old_balance, new_balance, activity_amount, currency, metadata, created_at FROM wallet_logs ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	walletLogs := make([]WalletLog, 0)
	for rows.Next() {
		walletLog := WalletLog{}
		err := rows.Scan(
			&walletLog.ID,
			&walletLog.UserID,
			&walletLog.WalletID,
			&walletLog.Activity,
			&walletLog.OldBalance,
			&walletLog.NewBalance,
			&walletLog.ActivityAmount,
			&walletLog.Currency,
			&walletLog.Metadata,
			&walletLog.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		walletLogs = append(walletLogs, walletLog)
	}
	return walletLogs, nil
}

// GetWalletLogsByUser gets a wallet logs by user
func (wl *WalletLog) GetWalletLogsByUser() ([]WalletLog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rows, err := DB.Query(
		ctx,
		`SELECT id, user_id, wallet_id, activity, old_balance, new_balance, activity_amount, currency, metadata, created_at FROM wallet_logs WHERE user_id = $1 ORDER BY created_at DESC`,
		wl.UserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	walletLogs := make([]WalletLog, 0)
	for rows.Next() {
		walletLog := WalletLog{}
		err := rows.Scan(
			&walletLog.ID,
			&walletLog.UserID,
			&walletLog.WalletID,
			&walletLog.Activity,
			&walletLog.OldBalance,
			&walletLog.NewBalance,
			&walletLog.ActivityAmount,
			&walletLog.Currency,
			&walletLog.Metadata,
			&walletLog.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		walletLogs = append(walletLogs, walletLog)
	}
	return walletLogs, nil
}
