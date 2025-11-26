package repository

import (
	"context"
	"errors"
	"ridash/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

// GetAccountByEmail retrieves an account by email address
func GetAccountByEmail(ctx context.Context, tx pgx.Tx, email string) (*models.Account, error) {
	query := `SELECT id, provider, provider_user_id, user_id, email, created_at, updated_at
	          FROM accounts
	          WHERE email = $1
	          LIMIT 1`

	var account models.Account
	err := tx.QueryRow(ctx, query, email).Scan(
		&account.ID,
		&account.Provider,
		&account.ProviderUserID,
		&account.UserID,
		&account.Email,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // Not an error, just not found
	}

	if err != nil {
		return nil, err
	}

	return &account, nil
}

// GetAccountWithUserByProviderUserID retrieves the account and its associated user
func GetAccountWithUserByProviderUserID(ctx context.Context, db pgx.Tx, provider models.Provider, providerUserID string) (*models.Account, *models.User, error) {
	query := `
		SELECT
			a.id AS "a.id", a.provider AS "a.provider", a.provider_user_id AS "a.provider_user_id", a.user_id AS "a.user_id",
			u.id AS "u.id", u.created_at AS "u.created_at", u.updated_at AS "u.updated_at"
		FROM accounts a
		JOIN users u ON a.user_id = u.id
		WHERE a.provider = $1 AND a.provider_user_id = $2
	`

	// Using aliases to scan into both Account and User
	var result struct {
		A models.Account `db:"a"`
		U models.User    `db:"u"`
	}

	err := pgxscan.Get(ctx, db, &result, query, provider, providerUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, nil
	} else if err != nil {
		return nil, nil, err
	}

	return &result.A, &result.U, nil
}

// CreateAccount creates a new account
func CreateAccount(ctx context.Context, tx pgx.Tx, account models.Account) error {
	query := `INSERT INTO accounts (id, provider, provider_user_id, user_id, email, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := tx.Exec(ctx, query,
		account.ID,
		account.Provider,
		account.ProviderUserID,
		account.UserID,
		account.Email,
		account.CreatedAt,
		account.UpdatedAt,
	)

	return err
}

// CreateUserAndAccount creates a new user and associated account in a transaction
func CreateUserAndAccount(ctx context.Context, tx pgx.Tx, user models.User, account models.Account) error {
	// Insert user
	userQuery := `INSERT INTO users (id, password_hash, display_name, avatar, created_at, updated_at)
	              VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := tx.Exec(ctx, userQuery,
		user.ID,
		user.PasswordHash,
		user.DisplayName,
		user.Avatar,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Insert account
	accountQuery := `INSERT INTO accounts (id, provider, provider_user_id, user_id, email, created_at, updated_at)
	                 VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = tx.Exec(ctx, accountQuery,
		account.ID,
		account.Provider,
		account.ProviderUserID,
		account.UserID,
		account.Email,
		account.CreatedAt,
		account.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

// CreateOAuthToken creates a new OAuth token
func CreateOAuthToken(ctx context.Context, db pgx.Tx, oauthToken models.OAuthToken) error {
	query := `
		INSERT INTO oauth_tokens (
			account_id,
			access_token,
			refresh_token,
			expiry,
			token_type,
			provider,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (account_id)
		DO UPDATE SET
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			expiry = EXCLUDED.expiry,
			token_type = EXCLUDED.token_type,
			updated_at = EXCLUDED.updated_at
	`

	_, err := db.Exec(ctx,
		query,
		oauthToken.AccountID,
		oauthToken.AccessToken,
		oauthToken.RefreshToken,
		oauthToken.Expiry,
		oauthToken.TokenType,
		oauthToken.Provider,
		oauthToken.CreatedAt,
		oauthToken.UpdatedAt,
	)
	return err
}
