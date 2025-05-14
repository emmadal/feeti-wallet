package models

import (
	"context"
	"time"
)

// createTables create tables
func createTables() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	queries := []string{
		`CREATE TABLE IF NOT EXISTS wallets (
			id SERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			balance BIGINT DEFAULT 0 NOT NULL,
			currency VARCHAR(3) DEFAULT 'XAF' NOT NULL,
    		locked BOOLEAN DEFAULT FALSE,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS wallet_logs (
    		id SERIAL PRIMARY KEY,
    		user_id BIGINT NOT NULL,
    		wallet_id BIGINT NOT NULL,
    		activity VARCHAR(50) NOT NULL, -- 'creation', 'balance_check', 'debit', etc.
    		old_balance BIGINT NOT NULL,
    		new_balance BIGINT NOT NULL,
    		activity_amount BIGINT NOT NULL,
    		currency VARCHAR(3) DEFAULT 'XAF' NOT NULL,
    		metadata JSONB, -- optional extra info (e.g. payment method, ref number)
    		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		
			CONSTRAINT fk_log_wallet FOREIGN KEY (wallet_id)
				REFERENCES wallets (id)
				ON DELETE CASCADE
				ON UPDATE CASCADE
    	);`,
		`CREATE INDEX IF NOT EXISTS idx_wallet_logs_user_id ON wallet_logs (user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_wallets_user_lookup ON wallets (user_id, locked, is_active);`,
	}
	for _, query := range queries {
		if _, err := DB.Exec(ctx, query); err != nil {
			return err
		}
	}
	return nil
}
