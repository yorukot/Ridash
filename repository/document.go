package repository

import (
	"context"
	"ridash/models"

	"github.com/jackc/pgx/v5"
)

// CreateDocument inserts a new document record.
func CreateDocument(ctx context.Context, tx pgx.Tx, doc models.Document) error {
	query := `INSERT INTO documents (id, owner_id, name, premission, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := tx.Exec(ctx, query,
		doc.ID,
		doc.OwnerID,
		doc.Name,
		doc.Permission,
		doc.CreatedAt,
		doc.UpdatedAt,
	)

	return err
}

// GetDocumentByID retrieves a document by its ID.
func GetDocumentByID(ctx context.Context, tx pgx.Tx, id int64) (*models.Document, error) {
	query := `SELECT id, owner_id, name, premission, created_at, updated_at
	          FROM documents
	          WHERE id = $1
	          LIMIT 1`

	var doc models.Document
	err := tx.QueryRow(ctx, query, id).Scan(
		&doc.ID,
		&doc.OwnerID,
		&doc.Name,
		&doc.Permission,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &doc, nil
}

// ListDocumentsForUser returns documents visible to the given user.
// If userID is nil, only public/public_write documents are returned.
func ListDocumentsForUser(ctx context.Context, tx pgx.Tx, userID *int64) ([]models.Document, error) {
	if userID == nil {
		query := `SELECT id, owner_id, name, premission, created_at, updated_at
		          FROM documents
		          WHERE premission IN ('public', 'public_write')
		          ORDER BY created_at DESC`

		rows, err := tx.Query(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var documents []models.Document
		for rows.Next() {
			var doc models.Document
			if err := rows.Scan(&doc.ID, &doc.OwnerID, &doc.Name, &doc.Permission, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
				return nil, err
			}
			documents = append(documents, doc)
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}

		return documents, nil
	}

	query := `SELECT DISTINCT d.id, d.owner_id, d.name, d.premission, d.created_at, d.updated_at
	          FROM documents d
	          LEFT JOIN docs_shares s ON s.document_id = d.id AND s.user_id = $1
	          WHERE d.owner_id = $1
	             OR d.premission IN ('public', 'public_write')
	             OR s.id IS NOT NULL
	          ORDER BY d.created_at DESC`

	rows, err := tx.Query(ctx, query, *userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []models.Document
	for rows.Next() {
		var doc models.Document
		if err := rows.Scan(&doc.ID, &doc.OwnerID, &doc.Name, &doc.Permission, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return documents, nil
}

// UpdateDocument updates name and permission for a document.
func UpdateDocument(ctx context.Context, tx pgx.Tx, id int64, name string, permission models.DocsPermission, updatedAt any) error {
	query := `UPDATE documents
	          SET name = $1, premission = $2, updated_at = $3
	          WHERE id = $4`

	_, err := tx.Exec(ctx, query, name, permission, updatedAt, id)
	return err
}

// DeleteDocument removes a document by ID.
func DeleteDocument(ctx context.Context, tx pgx.Tx, id int64) error {
	query := `DELETE FROM documents WHERE id = $1`
	_, err := tx.Exec(ctx, query, id)
	return err
}

// DeleteSharesByDocument removes all share rows for a document.
func DeleteSharesByDocument(ctx context.Context, tx pgx.Tx, documentID int64) error {
	query := `DELETE FROM docs_shares WHERE document_id = $1`
	_, err := tx.Exec(ctx, query, documentID)
	return err
}

// ListSharesByDocument lists all shares for a given document.
func ListSharesByDocument(ctx context.Context, tx pgx.Tx, documentID int64) ([]models.DocsShare, error) {
	query := `SELECT id, document_id, user_id, roles
	          FROM docs_shares
	          WHERE document_id = $1
	          ORDER BY id ASC`

	rows, err := tx.Query(ctx, query, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shares []models.DocsShare
	for rows.Next() {
		var share models.DocsShare
		if err := rows.Scan(&share.ID, &share.DocumentID, &share.UserID, &share.Roles); err != nil {
			return nil, err
		}
		shares = append(shares, share)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return shares, nil
}

// GetShareByID retrieves a share by its ID.
func GetShareByID(ctx context.Context, tx pgx.Tx, shareID int64) (*models.DocsShare, error) {
	query := `SELECT id, document_id, user_id, roles
	          FROM docs_shares
	          WHERE id = $1
	          LIMIT 1`

	var share models.DocsShare
	err := tx.QueryRow(ctx, query, shareID).Scan(
		&share.ID,
		&share.DocumentID,
		&share.UserID,
		&share.Roles,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &share, nil
}

// GetShareByDocumentAndUser returns a share if a user has access to a document.
func GetShareByDocumentAndUser(ctx context.Context, tx pgx.Tx, documentID, userID int64) (*models.DocsShare, error) {
	query := `SELECT id, document_id, user_id, roles
	          FROM docs_shares
	          WHERE document_id = $1 AND user_id = $2
	          LIMIT 1`

	var share models.DocsShare
	err := tx.QueryRow(ctx, query, documentID, userID).Scan(
		&share.ID,
		&share.DocumentID,
		&share.UserID,
		&share.Roles,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &share, nil
}

// CreateShare inserts a new share row.
func CreateShare(ctx context.Context, tx pgx.Tx, share models.DocsShare) error {
	query := `INSERT INTO docs_shares (id, document_id, user_id, roles)
	          VALUES ($1, $2, $3, $4)`

	_, err := tx.Exec(ctx, query,
		share.ID,
		share.DocumentID,
		share.UserID,
		share.Roles,
	)

	return err
}

// UpdateShare updates the permission for an existing share.
func UpdateShare(ctx context.Context, tx pgx.Tx, shareID int64, permission models.DocsSharePermission) error {
	query := `UPDATE docs_shares
	          SET roles = $1
	          WHERE id = $2`

	_, err := tx.Exec(ctx, query, permission, shareID)
	return err
}

// DeleteShare removes a share by ID.
func DeleteShare(ctx context.Context, tx pgx.Tx, shareID int64) error {
	query := `DELETE FROM docs_shares WHERE id = $1`
	_, err := tx.Exec(ctx, query, shareID)
	return err
}
