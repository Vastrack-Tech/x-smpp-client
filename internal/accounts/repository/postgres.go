package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"x-smpp-client/internal/database"
	"x-smpp-client/internal/models"
)

type PostgresRepo struct {
	db *database.DB
}

func NewPostgresRepo(db *database.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) CreateAccount(ctx context.Context, a *models.Account) error {
	err := r.db.Pool.QueryRow(ctx,
		`INSERT INTO accounts (name, email, password_hash, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		a.Name, a.Email, a.PasswordHash, a.CreatedAt, a.UpdatedAt,
	).Scan(&a.ID)
	return err
}

func (r *PostgresRepo) GetAccount(ctx context.Context, id string) (*models.Account, error) {
	a := &models.Account{}
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, name, email, password_hash, created_at, updated_at FROM accounts WHERE id = $1`, id,
	).Scan(&a.ID, &a.Name, &a.Email, &a.PasswordHash, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, err
	}
	return a, nil
}

func (r *PostgresRepo) GetAccountByEmail(ctx context.Context, email string) (*models.Account, error) {
	a := &models.Account{}
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, name, email, password_hash, created_at, updated_at FROM accounts WHERE email = $1`, email,
	).Scan(&a.ID, &a.Name, &a.Email, &a.PasswordHash, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, err
	}
	return a, nil
}

func (r *PostgresRepo) CreateAPIKey(ctx context.Context, k *models.APIKey) error {
	err := r.db.Pool.QueryRow(ctx,
		`INSERT INTO api_keys (key, account_id, name, active, created_at)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		k.Key, k.AccountID, k.Name, k.Active, k.CreatedAt,
	).Scan(&k.ID)
	return err
}

func (r *PostgresRepo) GetAPIKey(ctx context.Context, key string) (*models.APIKey, error) {
	k := &models.APIKey{}
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, key, account_id, name, active, created_at FROM api_keys WHERE key = $1 AND active = TRUE`, key,
	).Scan(&k.ID, &k.Key, &k.AccountID, &k.Name, &k.Active, &k.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("api key not found")
		}
		return nil, err
	}
	return k, nil
}

func (r *PostgresRepo) ListAPIKeys(ctx context.Context, accountID string) ([]models.APIKey, error) {
	rows, err := r.db.Pool.Query(ctx,
		`SELECT id, key, account_id, name, active, created_at FROM api_keys WHERE account_id = $1 ORDER BY created_at DESC`, accountID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []models.APIKey
	for rows.Next() {
		var k models.APIKey
		if err := rows.Scan(&k.ID, &k.Key, &k.AccountID, &k.Name, &k.Active, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (r *PostgresRepo) CreateLedgerAccount(ctx context.Context, la *models.LedgerAccount) error {
	err := r.db.Pool.QueryRow(ctx,
		`INSERT INTO ledger_accounts (account_id, currency, created_at, updated_at)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		la.AccountID, la.Currency, la.CreatedAt, la.UpdatedAt,
	).Scan(&la.ID)
	return err
}

func (r *PostgresRepo) CreateLedgerBalance(ctx context.Context, lb *models.LedgerBalance) error {
	err := r.db.Pool.QueryRow(ctx,
		`INSERT INTO ledger_balance (ledger_account_id, balance, version, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		lb.LedgerAccountID, lb.Balance, lb.Version, lb.CreatedAt, lb.UpdatedAt,
	).Scan(&lb.ID)
	return err
}

func (r *PostgresRepo) GetLedgerAccount(ctx context.Context, accountID string) (*models.LedgerAccount, error) {
	la := &models.LedgerAccount{}
	err := r.db.Pool.QueryRow(ctx,
		`SELECT id, account_id, currency, created_at, updated_at
		 FROM ledger_accounts WHERE account_id = $1`, accountID,
	).Scan(&la.ID, &la.AccountID, &la.Currency, &la.CreatedAt, &la.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("ledger account not found for account %s", accountID)
		}
		return nil, err
	}
	return la, nil
}

func (r *PostgresRepo) GetLedgerBalance(ctx context.Context, accountID string) (*models.LedgerBalance, error) {
	lb := &models.LedgerBalance{}
	err := r.db.Pool.QueryRow(ctx,
		`SELECT lb.id, lb.ledger_account_id, lb.balance, lb.version, lb.created_at, lb.updated_at
		 FROM ledger_balance lb
		 JOIN ledger_accounts la ON la.id = lb.ledger_account_id
		 WHERE la.account_id = $1`, accountID,
	).Scan(&lb.ID, &lb.LedgerAccountID, &lb.Balance, &lb.Version, &lb.CreatedAt, &lb.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("ledger balance not found for account %s", accountID)
		}
		return nil, err
	}
	return lb, nil
}

func (r *PostgresRepo) ApplyBalanceChange(ctx context.Context, accountID string, change int64, entryType, reference, description string) (*models.LedgerEntry, error) {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	la, lb, err := r.lockLedger(ctx, tx, accountID)
	if err != nil {
		return nil, err
	}

	newBalance := lb.Balance + change
	if newBalance < 0 {
		return nil, fmt.Errorf("insufficient balance: have %d, need %d", lb.Balance, -change)
	}

	tag, err := tx.Exec(ctx,
		`UPDATE ledger_balance SET balance = $1, version = version + 1, updated_at = $2
		 WHERE id = $3 AND version = $4`,
		newBalance, time.Now(), lb.ID, lb.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("update ledger balance: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return nil, fmt.Errorf("balance update conflict, retry")
	}

	e := &models.LedgerEntry{
		LedgerAccountID: la.ID,
		Type:            entryType,
		Amount:          change,
		Reference:       reference,
		Description:     description,
		CreatedAt:       time.Now(),
	}
	err = tx.QueryRow(ctx,
		`INSERT INTO ledger_entries (ledger_account_id, type, amount, reference, description, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		e.LedgerAccountID, e.Type, e.Amount, e.Reference, e.Description, e.CreatedAt,
	).Scan(&e.ID)
	if err != nil {
		return nil, fmt.Errorf("insert ledger entry: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return e, nil
}

func (r *PostgresRepo) lockLedger(ctx context.Context, tx pgx.Tx, accountID string) (*models.LedgerAccount, *models.LedgerBalance, error) {
	la := &models.LedgerAccount{}
	lb := &models.LedgerBalance{}
	err := tx.QueryRow(ctx,
		`SELECT la.id, la.account_id, la.currency, la.created_at, la.updated_at,
		        lb.id, lb.ledger_account_id, lb.balance, lb.version, lb.created_at, lb.updated_at
		 FROM ledger_accounts la
		 JOIN ledger_balance lb ON lb.ledger_account_id = la.id
		 WHERE la.account_id = $1
		 FOR UPDATE OF lb`, accountID,
	).Scan(
		&la.ID, &la.AccountID, &la.Currency, &la.CreatedAt, &la.UpdatedAt,
		&lb.ID, &lb.LedgerAccountID, &lb.Balance, &lb.Version, &lb.CreatedAt, &lb.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, fmt.Errorf("ledger not found for account %s", accountID)
		}
		return nil, nil, err
	}
	return la, lb, nil
}

func (r *PostgresRepo) GetLedgerEntries(ctx context.Context, ledgerAccountID string, limit, offset int) ([]models.LedgerEntry, error) {
	rows, err := r.db.Pool.Query(ctx,
		`SELECT id, ledger_account_id, type, amount, reference, description, created_at
		 FROM ledger_entries WHERE ledger_account_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		ledgerAccountID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.LedgerEntry
	for rows.Next() {
		var e models.LedgerEntry
		if err := rows.Scan(&e.ID, &e.LedgerAccountID, &e.Type, &e.Amount, &e.Reference, &e.Description, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (r *PostgresRepo) CreateMessage(ctx context.Context, m *models.Message) error {
	err := r.db.Pool.QueryRow(ctx,
		`INSERT INTO messages (account_id, to_addr, text, encoding, source_addr, status, cost, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		m.AccountID, m.To, m.Text, m.Encoding, m.SourceAddr, m.Status, m.Cost, m.CreatedAt, m.UpdatedAt,
	).Scan(&m.ID)
	return err
}

func (r *PostgresRepo) UpdateMessageStatus(ctx context.Context, id, status string, parts int, cost int64) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE messages SET status = $1, parts = $2, cost = $3, updated_at = NOW() WHERE id = $4`,
		status, parts, cost, id,
	)
	return err
}

func (r *PostgresRepo) SaveDeliveryReceipt(ctx context.Context, d *models.DeliveryReceipt) error {
	_, err := r.db.Pool.Exec(ctx,
		`INSERT INTO delivery_receipts (message_id, account_id, status, error_code, raw, received_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		d.MessageID, d.AccountID, d.Status, d.ErrorCode, d.Raw, d.ReceivedAt,
	)
	return err
}
