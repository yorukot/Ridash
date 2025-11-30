package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"ridash/models"
)

func TestEndpointsCoverageExceptOAuth(t *testing.T) {
	ctx := context.Background()

	pool, server, _ := initApp(t, ctx)
	client := newAPIClient(t, server.URL)

	ownerEmail := "owner@example.com"
	memberEmail := "member@example.com"

	client.Register(t, ownerEmail, "password123", "Owner")
	client.Login(t, ownerEmail, "password123")
	ownerToken := client.RefreshAccessToken(t)

	client.Register(t, memberEmail, "password123", "Member")
	client.Login(t, memberEmail, "password123")
	memberToken := client.RefreshAccessToken(t)

	memberID := getUserIDByEmail(t, pool, memberEmail)

	team := client.CreateTeam(t, ownerToken, "Integration Team")
	teams := client.ListTeams(t, ownerToken)
	require.NotEmpty(t, teams)

	fetchedTeam := client.GetTeam(t, ownerToken, team.ID)
	require.Equal(t, team.ID, fetchedTeam.ID)

	updatedTeam := client.UpdateTeam(t, ownerToken, team.ID, "Renamed Team")
	require.Equal(t, "Renamed Team", updatedTeam.Name)

	rootFolder := client.CreateFolder(t, ownerToken, team.ID, "Root Folder", nil)
	childFolder := client.CreateFolder(t, ownerToken, team.ID, "Child Folder", &rootFolder.ID)

	folders := client.ListFolders(t, ownerToken, team.ID)
	require.Len(t, folders, 2)

	fetchedRoot := client.GetFolder(t, ownerToken, team.ID, rootFolder.ID)
	require.Equal(t, rootFolder.ID, fetchedRoot.ID)

	updatedChild := client.UpdateFolder(t, ownerToken, team.ID, childFolder.ID, "Child Folder Updated", nil)
	require.Equal(t, "Child Folder Updated", updatedChild.Name)
	require.Nil(t, updatedChild.ParentFolder)

	document := client.CreateDocument(t, ownerToken, rootFolder.ID, "Secret Plan", models.DocsPermissionPrivate)

	ownerDocs := client.ListDocuments(t, ownerToken)
	require.Len(t, ownerDocs, 1)
	require.Equal(t, document.ID, ownerDocs[0].ID)

	publicDocs := client.ListDocuments(t, "")
	require.Empty(t, publicDocs)

	updatedDoc := client.UpdateDocument(t, ownerToken, document.ID, "Secret Plan v2", models.DocsPermissionPrivate)
	require.Equal(t, "Secret Plan v2", updatedDoc.Name)

	share := client.CreateShare(t, ownerToken, document.ID, memberID, models.DocsSharePermissionRead)
	shares := client.ListShares(t, ownerToken, document.ID)
	require.Len(t, shares, 1)
	require.Equal(t, share.ID, shares[0].ID)

	memberDocs := client.ListDocuments(t, memberToken)
	require.Len(t, memberDocs, 1)
	require.Equal(t, document.ID, memberDocs[0].ID)

	updatedShare := client.UpdateShare(t, ownerToken, document.ID, share.ID, models.DocsSharePermissionWrite)
	require.Equal(t, models.DocsSharePermissionWrite, updatedShare.Roles)

	client.DeleteShare(t, ownerToken, document.ID, share.ID)

	memberDocsAfterDelete := client.ListDocuments(t, memberToken)
	require.Empty(t, memberDocsAfterDelete)

	client.DeleteDocument(t, ownerToken, document.ID)
	require.Empty(t, client.ListDocuments(t, ownerToken))

	client.DeleteFolder(t, ownerToken, team.ID, childFolder.ID)
	client.DeleteFolder(t, ownerToken, team.ID, rootFolder.ID)

	client.DeleteTeam(t, ownerToken, team.ID)
	require.Empty(t, client.ListTeams(t, ownerToken))
}
