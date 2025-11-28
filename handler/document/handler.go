package document

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type DocumentHandler struct {
	DB *pgxpool.Pool
}
