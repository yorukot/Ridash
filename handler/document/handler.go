package document

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"ridash/utils/docmanager"
)

type DocumentHandler struct {
	DB         *pgxpool.Pool
	DocManager *docmanager.Client
}
