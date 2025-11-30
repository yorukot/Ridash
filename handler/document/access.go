package document

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"ridash/models"
	"ridash/repository"
)

type documentContext struct {
	Document *models.Document
	Folder   *models.Folder
	Team     *models.Team
}

// loadDocumentContext loads document, folder, and team to evaluate permissions.
func loadDocumentContext(ctx context.Context, tx pgx.Tx, docID int64) (documentContext, error) {
	doc, err := repository.GetDocumentByID(ctx, tx, docID)
	if err != nil {
		return documentContext{}, err
	}
	if doc == nil {
		return documentContext{}, nil
	}

	folder, err := repository.GetFolderByID(ctx, tx, doc.FolderID)
	if err != nil {
		return documentContext{}, err
	}
	if folder == nil {
		return documentContext{}, fmt.Errorf("folder %d not found for document %d", doc.FolderID, doc.ID)
	}

	team, err := repository.GetTeamByID(ctx, tx, folder.TeamID)
	if err != nil {
		return documentContext{}, err
	}
	if team == nil {
		return documentContext{}, fmt.Errorf("team %d not found for folder %d", folder.TeamID, folder.ID)
	}

	return documentContext{
		Document: doc,
		Folder:   folder,
		Team:     team,
	}, nil
}

func isTeamOwner(userID int64, team *models.Team) bool {
	return team != nil && team.OwnerID == userID
}
