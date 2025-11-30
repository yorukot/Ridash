package e2e

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	"ridash/models"
)

func TestDocumentEditProxy(t *testing.T) {
	ctx := context.Background()

	_, server, docStub := initApp(t, ctx)
	client := newAPIClient(t, server.URL)

	client.Register(t, "owner2@example.com", "password123", "Owner2")
	client.Login(t, "owner2@example.com", "password123")
	token := client.RefreshAccessToken(t)

	team := client.CreateTeam(t, token, "Proxy Team")
	folder := client.CreateFolder(t, token, team.ID, "Proxy Folder", nil)
	doc := client.CreateDocument(t, token, folder.ID, "Proxy Doc", models.DocsPermissionPrivate)

	resp := client.doJSON(t, http.MethodGet, "/api/documents/"+strconv.FormatInt(doc.ID, 10)+"/socket", token, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from websocket proxy, got %d", resp.StatusCode)
	}

	waitForDocEdit(t, docStub)
}
