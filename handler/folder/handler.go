package folder

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type FolderHandler struct {
	DB *pgxpool.Pool
}
