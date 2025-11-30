package e2e

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"ridash/models"
)

func TestAuthAndDocumentFlow(t *testing.T) {
	ctx := context.Background()

	_, server, _ := initApp(t, ctx)
	client := newAPIClient(t, server.URL)

	client.Register(t, "alice@example.com", "password123", "Alice")
	client.Login(t, "alice@example.com", "password123")
	token := client.RefreshAccessToken(t)

	team := client.CreateTeam(t, token, "Team A")
	folder := client.CreateFolder(t, token, team.ID, "Root Folder", nil)

	createdDoc := client.CreateDocument(t, token, folder.ID, "My First Doc", models.DocsPermissionPrivate)
	fetchedDoc := client.GetDocument(t, token, createdDoc.ID)

	require.Equal(t, createdDoc.ID, fetchedDoc.Document.ID)
	require.Equal(t, "My First Doc", fetchedDoc.Document.Name)
	require.Equal(t, models.DocsPermissionPrivate, fetchedDoc.Document.Permission)
	require.Equal(t, "stub-content-"+strconv.FormatInt(createdDoc.ID, 10), fetchedDoc.Content)
	require.Equal(t, int64(1), fetchedDoc.Seq)
}
