package database

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS accounts (
		id            UUID PRIMARY KEY DEFAULT uuidv7(),
		name          TEXT NOT NULL,
		email         TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL DEFAULT '',
		created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`,

	`CREATE TABLE IF NOT EXISTS api_keys (
		id         UUID PRIMARY KEY DEFAULT uuidv7(),
		key        TEXT NOT NULL UNIQUE,
		account_id UUID NOT NULL REFERENCES accounts(id),
		name       TEXT NOT NULL DEFAULT '',
		active     BOOLEAN NOT NULL DEFAULT TRUE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`,

	`CREATE TABLE IF NOT EXISTS ledger_accounts (
		id         UUID PRIMARY KEY DEFAULT uuidv7(),
		account_id UUID NOT NULL UNIQUE REFERENCES accounts(id),
		currency   TEXT NOT NULL DEFAULT 'kobo',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`,

	`CREATE TABLE IF NOT EXISTS ledger_balance (
		id                UUID PRIMARY KEY DEFAULT uuidv7(),
		ledger_account_id UUID NOT NULL UNIQUE REFERENCES ledger_accounts(id),
		balance           BIGINT NOT NULL DEFAULT 0,
		version           INT NOT NULL DEFAULT 0,
		created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`,

	`CREATE TABLE IF NOT EXISTS ledger_entries (
		id                UUID PRIMARY KEY DEFAULT uuidv7(),
		ledger_account_id UUID NOT NULL REFERENCES ledger_accounts(id),
		type              TEXT NOT NULL,
		amount            BIGINT NOT NULL,
		reference         TEXT NOT NULL DEFAULT '',
		description       TEXT NOT NULL DEFAULT '',
		created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`,

	`CREATE TABLE IF NOT EXISTS messages (
		id          UUID PRIMARY KEY DEFAULT uuidv7(),
		account_id  UUID NOT NULL REFERENCES accounts(id),
		to_addr     TEXT NOT NULL,
		text        TEXT NOT NULL,
		encoding    TEXT NOT NULL DEFAULT 'gsm',
		source_addr TEXT NOT NULL DEFAULT '',
		parts       INT NOT NULL DEFAULT 0,
		status      TEXT NOT NULL DEFAULT 'queued',
		cost        BIGINT NOT NULL DEFAULT 0,
		created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`,

	`CREATE TABLE IF NOT EXISTS delivery_receipts (
		id          BIGSERIAL PRIMARY KEY,
		message_id  TEXT NOT NULL,
		account_id  TEXT NOT NULL DEFAULT '',
		status      TEXT NOT NULL DEFAULT '',
		error_code  TEXT NOT NULL DEFAULT '',
		raw         TEXT NOT NULL DEFAULT '',
		received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`,

	`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL DEFAULT ''`,

	`CREATE INDEX IF NOT EXISTS idx_api_keys_key ON api_keys(key)`,
	`CREATE INDEX IF NOT EXISTS idx_ledger_balance_account_id ON ledger_balance(ledger_account_id)`,
	`CREATE INDEX IF NOT EXISTS idx_ledger_entries_ledger_account_id ON ledger_entries(ledger_account_id)`,
	`CREATE INDEX IF NOT EXISTS idx_messages_account_id ON messages(account_id)`,
}
