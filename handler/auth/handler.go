package auth

import (
	"ridash/utils/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthHandler struct {
	DB *pgxpool.Pool
	OAuthConfig *config.OAuthConfig
}