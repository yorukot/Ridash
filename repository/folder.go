package repository

import (
	"context"
	"database/sql"
	"ridash/models"

	"github.com/jackc/pgx/v5"
)

// CreateFolder inserts a new folder record
func CreateFolder(ctx context.Context, tx pgx.Tx, folder models.Folder) error {
	query := `INSERT INTO folders (id, team_id, name, parent_folder, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := tx.Exec(ctx, query,
		folder.ID,
		folder.TeamID,
		folder.Name,
		folder.ParentFolder,
		folder.CreatedAt,
		folder.UpdatedAt,
	)

	return err
}

// GetFolderByIDAndTeamID retrieves a folder ensuring it belongs to the given team
func GetFolderByIDAndTeamID(ctx context.Context, tx pgx.Tx, folderID, teamID int64) (*models.Folder, error) {
	query := `SELECT id, team_id, name, parent_folder, created_at, updated_at
	          FROM folders
	          WHERE id = $1 AND team_id = $2
	          LIMIT 1`

	var folder models.Folder
	var parentFolder sql.NullInt64

	err := tx.QueryRow(ctx, query, folderID, teamID).Scan(
		&folder.ID,
		&folder.TeamID,
		&folder.Name,
		&parentFolder,
		&folder.CreatedAt,
		&folder.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if parentFolder.Valid {
		folder.ParentFolder = &parentFolder.Int64
	}

	return &folder, nil
}

// GetFoldersByTeamID retrieves all folders for a team
func GetFoldersByTeamID(ctx context.Context, tx pgx.Tx, teamID int64) ([]models.Folder, error) {
	query := `SELECT id, team_id, name, parent_folder, created_at, updated_at
	          FROM folders
	          WHERE team_id = $1
	          ORDER BY created_at ASC`

	rows, err := tx.Query(ctx, query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []models.Folder

	for rows.Next() {
		var folder models.Folder
		var parentFolder sql.NullInt64

		if err := rows.Scan(
			&folder.ID,
			&folder.TeamID,
			&folder.Name,
			&parentFolder,
			&folder.CreatedAt,
			&folder.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if parentFolder.Valid {
			folder.ParentFolder = &parentFolder.Int64
		}

		folders = append(folders, folder)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return folders, nil
}

// UpdateFolder updates folder fields for a given folder and team
func UpdateFolder(ctx context.Context, tx pgx.Tx, folderID, teamID int64, name string, parentFolder *int64, updatedAt any) error {
	query := `UPDATE folders
	          SET name = $1, parent_folder = $2, updated_at = $3
	          WHERE id = $4 AND team_id = $5`

	_, err := tx.Exec(ctx, query, name, parentFolder, updatedAt, folderID, teamID)
	return err
}

// DeleteFolder removes a folder by ID and team
func DeleteFolder(ctx context.Context, tx pgx.Tx, folderID, teamID int64) error {
	query := `DELETE FROM folders WHERE id = $1 AND team_id = $2`
	_, err := tx.Exec(ctx, query, folderID, teamID)
	return err
}
