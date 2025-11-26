package repository

import (
	"context"
	"ridash/models"

	"github.com/jackc/pgx/v5"
)

// CreateRefreshToken creates a new refresh token in the database
func CreateRefreshToken(ctx context.Context, tx pgx.Tx, token models.RefreshToken) error {
	query := `INSERT INTO refresh_tokens (id, user_id, token, user_agent, ip, used_at, created_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := tx.Exec(ctx, query,
		token.ID,
		token.UserID,
		token.Token,
		token.UserAgent,
		token.IP,
		token.UsedAt,
		token.CreatedAt,
	)

	return err
}
