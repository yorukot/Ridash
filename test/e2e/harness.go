package e2e

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"ridash/db"
	customMiddleware "ridash/middleware"
	"ridash/models"
	"ridash/repository"
	"ridash/router"
	"ridash/utils/config"
	"ridash/utils/docmanager"
	"ridash/utils/id"
	"ridash/utils/logger"
)

type docManagerStub struct {
	URL     string
	server  *httptest.Server
	editCh  chan struct{}
	tickets map[int64][]string
}

func startDocManagerStub(t *testing.T) *docManagerStub {
	t.Helper()

	stub := &docManagerStub{
		editCh:  make(chan struct{}, 1),
		tickets: make(map[int64][]string),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/documents/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/documents/")
		parts := strings.Split(path, "/")
		if len(parts) == 0 || parts[0] == "" {
			http.NotFound(w, r)
			return
		}

		docID := parts[0]
		if len(parts) > 1 && parts[1] == "ticket" {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			idVal, _ := strconv.ParseInt(docID, 10, 64)
			stub.tickets[idVal] = append(stub.tickets[idVal], "ticket-"+docID)
			writeJSON(t, w, http.StatusOK, docmanager.IssueTicketResponse{
				Ticket:    "ticket-" + docID,
				ExpiresAt: time.Now().Add(time.Minute * 5).Unix(),
			})
			return
		}

		switch r.Method {
		case http.MethodGet:
			idVal, _ := strconv.ParseInt(docID, 10, 64)
			writeJSON(t, w, http.StatusOK, docmanager.DocumentContent{
				DocID:   idVal,
				Content: "stub-content-" + docID,
				Seq:     1,
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/edit", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		select {
		case stub.editCh <- struct{}{}:
		default:
		}
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	stub.URL = server.URL
	stub.server = server
	return stub
}

func writeJSON(t *testing.T, w http.ResponseWriter, status int, payload any) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	require.NoError(t, json.NewEncoder(w).Encode(payload))
}

func startPostgresContainer(t *testing.T, ctx context.Context) *postgres.PostgresContainer {
	t.Helper()

	pg, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("timescale/timescaledb:latest-pg17"),
		postgres.WithDatabase("ridash"),
		postgres.WithUsername("ridash"),
		postgres.WithPassword("ridash-this-is-a-really-long-password"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp").WithStartupTimeout(90*time.Second)),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = pg.Terminate(context.Background())
	})

	return pg
}

func switchToProjectRoot(t *testing.T) func() {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)

	root := filepath.Clean(filepath.Join(wd, "..", ".."))
	require.NoError(t, os.Chdir(root))

	return func() { _ = os.Chdir(wd) }
}

func setTestEnv(t *testing.T, ctx context.Context, pg *postgres.PostgresContainer, docStub *docManagerStub) {
	t.Helper()

	host, err := pg.Host(ctx)
	require.NoError(t, err)

	port, err := pg.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	envs := map[string]string{
		"APP_ENV":                  "dev",
		"APP_NAME":                 "ridash-e2e",
		"APP_MACHINE_ID":           "1",
		"APP_PORT":                 "18000",
		"DB_HOST":                  host,
		"DB_PORT":                  port.Port(),
		"DB_USER":                  "ridash",
		"DB_PASSWORD":              "ridash-this-is-a-really-long-password",
		"DB_NAME":                  "ridash",
		"DB_SSL_MODE":              "disable",
		"SMTP_HOST":                "localhost",
		"SMTP_PORT":                "1025",
		"SMTP_USERNAME":            "user",
		"SMTP_PASSWORD":            "pass",
		"SMTP_FROM":                "noreply@example.com",
		"GOOGLE_CLIENT_ID":         "test-client-id",
		"GOOGLE_CLIENT_SECRET":     "test-client-secret",
		"GOOGLE_REDIRECT_URL":      "http://localhost/callback",
		"OAUTH_STATE_EXPIRES_AT":   "600",
		"ACCESS_TOKEN_EXPIRES_AT":  "900",
		"REFRESH_TOKEN_EXPIRES_AT": strconv.Itoa(int(time.Hour.Seconds())),
		"JWT_SECRET_KEY":           "test-secret-key",
		"FRONTEND_DOMAIN":          "127.0.0.1",
		"DOC_MANAGER_BASE_URL":     docStub.URL,
		"DOC_MANAGER_API_TOKEN":    "stub-token",
	}

	for key, val := range envs {
		require.NoError(t, os.Setenv(key, val))
	}
}

func startAppServer(t *testing.T, pool *pgxpool.Pool) *httptest.Server {
	t.Helper()

	e := echo.New()
	e.Use(customMiddleware.ZapLogger(zap.L()))
	e.Use(echomw.Recover())

	api := e.Group("/api")
	router.AuthRouter(api, pool)
	router.TeamRouter(api, pool)
	router.FolderRouter(api, pool)
	router.DocumentRouter(api, pool)

	server := httptest.NewServer(e)
	t.Cleanup(server.Close)
	return server
}

type apiClient struct {
	baseURL string
	client  *http.Client
}

func newAPIClient(t *testing.T, baseURL string) *apiClient {
	t.Helper()

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	return &apiClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 15 * time.Second,
			Jar:     jar,
		},
	}
}

func (c *apiClient) doJSON(t *testing.T, method, path, token string, body any) *http.Response {
	t.Helper()

	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		require.NoError(t, err)
		payload = strings.NewReader(string(encoded))
	}

	req, err := http.NewRequest(method, c.baseURL+path, payload)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.client.Do(req)
	require.NoError(t, err)
	return resp
}

type successResponse[T any] struct {
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func decodeSuccess[T any](t *testing.T, resp *http.Response, target *successResponse[T]) {
	t.Helper()
	defer resp.Body.Close()

	require.NoError(t, json.NewDecoder(resp.Body).Decode(target))
}

func (c *apiClient) Register(t *testing.T, email, password, displayName string) {
	t.Helper()

	resp := c.doJSON(t, http.MethodPost, "/api/auth/register", "", map[string]string{
		"email":        email,
		"password":     password,
		"display_name": displayName,
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	decodeSuccess(t, resp, &successResponse[struct{}]{})
}

func (c *apiClient) Login(t *testing.T, email, password string) {
	t.Helper()

	resp := c.doJSON(t, http.MethodPost, "/api/auth/login", "", map[string]string{
		"email":    email,
		"password": password,
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	decodeSuccess(t, resp, &successResponse[struct{}]{})
}

func (c *apiClient) RefreshAccessToken(t *testing.T) string {
	t.Helper()

	resp := c.doJSON(t, http.MethodPost, "/api/auth/refresh", "", struct{}{})
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var parsed successResponse[map[string]string]
	decodeSuccess(t, resp, &parsed)

	token, ok := parsed.Data["access_token"]
	require.True(t, ok, "access_token missing in response")
	return token
}

func (c *apiClient) CreateTeam(t *testing.T, token, name string) models.Team {
	t.Helper()

	resp := c.doJSON(t, http.MethodPost, "/api/teams", token, map[string]string{
		"name": name,
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var parsed successResponse[models.Team]
	decodeSuccess(t, resp, &parsed)
	return parsed.Data
}

func (c *apiClient) ListTeams(t *testing.T, token string) []models.Team {
	t.Helper()

	resp := c.doJSON(t, http.MethodGet, "/api/teams", token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var parsed successResponse[[]models.Team]
	decodeSuccess(t, resp, &parsed)
	return parsed.Data
}

func (c *apiClient) GetTeam(t *testing.T, token string, teamID int64) models.Team {
	t.Helper()

	resp := c.doJSON(t, http.MethodGet, "/api/teams/"+strconv.FormatInt(teamID, 10), token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var parsed successResponse[models.Team]
	decodeSuccess(t, resp, &parsed)
	return parsed.Data
}

func (c *apiClient) UpdateTeam(t *testing.T, token string, teamID int64, name string) models.Team {
	t.Helper()

	resp := c.doJSON(t, http.MethodPut, "/api/teams/"+strconv.FormatInt(teamID, 10), token, map[string]string{
		"name": name,
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var parsed successResponse[models.Team]
	decodeSuccess(t, resp, &parsed)
	return parsed.Data
}

func (c *apiClient) DeleteTeam(t *testing.T, token string, teamID int64) {
	t.Helper()

	resp := c.doJSON(t, http.MethodDelete, "/api/teams/"+strconv.FormatInt(teamID, 10), token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	decodeSuccess(t, resp, &successResponse[struct{}]{})
}

func (c *apiClient) CreateFolder(t *testing.T, token string, teamID int64, name string, parentFolder *int64) models.Folder {
	t.Helper()

	body := map[string]any{
		"name": name,
	}
	if parentFolder != nil {
		body["parent_folder"] = parentFolder
	}

	resp := c.doJSON(t, http.MethodPost, "/api/teams/"+strconv.FormatInt(teamID, 10)+"/folders", token, body)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var parsed successResponse[models.Folder]
	decodeSuccess(t, resp, &parsed)
	return parsed.Data
}

func (c *apiClient) ListFolders(t *testing.T, token string, teamID int64) []models.Folder {
	t.Helper()

	resp := c.doJSON(t, http.MethodGet, "/api/teams/"+strconv.FormatInt(teamID, 10)+"/folders", token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var parsed successResponse[[]models.Folder]
	decodeSuccess(t, resp, &parsed)
	return parsed.Data
}

func (c *apiClient) GetFolder(t *testing.T, token string, teamID, folderID int64) models.Folder {
	t.Helper()

	resp := c.doJSON(t, http.MethodGet, "/api/teams/"+strconv.FormatInt(teamID, 10)+"/folders/"+strconv.FormatInt(folderID, 10), token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var parsed successResponse[models.Folder]
	decodeSuccess(t, resp, &parsed)
	return parsed.Data
}

func (c *apiClient) UpdateFolder(t *testing.T, token string, teamID, folderID int64, name string, parentFolder *int64) models.Folder {
	t.Helper()

	body := map[string]any{
		"name": name,
	}
	if parentFolder != nil {
		body["parent_folder"] = parentFolder
	}

	resp := c.doJSON(t, http.MethodPut, "/api/teams/"+strconv.FormatInt(teamID, 10)+"/folders/"+strconv.FormatInt(folderID, 10), token, body)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var parsed successResponse[models.Folder]
	decodeSuccess(t, resp, &parsed)
	return parsed.Data
}

func (c *apiClient) DeleteFolder(t *testing.T, token string, teamID, folderID int64) {
	t.Helper()

	resp := c.doJSON(t, http.MethodDelete, "/api/teams/"+strconv.FormatInt(teamID, 10)+"/folders/"+strconv.FormatInt(folderID, 10), token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	decodeSuccess(t, resp, &successResponse[struct{}]{})
}

func (c *apiClient) CreateDocument(t *testing.T, token string, folderID int64, name string, permission models.DocsPermission) models.Document {
	t.Helper()

	resp := c.doJSON(t, http.MethodPost, "/api/documents", token, map[string]any{
		"folder_id":  strconv.FormatInt(folderID, 10),
		"name":       name,
		"permission": string(permission),
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var parsed successResponse[models.Document]
	decodeSuccess(t, resp, &parsed)
	return parsed.Data
}

func (c *apiClient) GetDocument(t *testing.T, token string, id int64) models.DocumentWithContent {
	t.Helper()

	resp := c.doJSON(t, http.MethodGet, "/api/documents/"+strconv.FormatInt(id, 10), token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var parsed successResponse[models.DocumentWithContent]
	decodeSuccess(t, resp, &parsed)
	return parsed.Data
}

func (c *apiClient) ListDocuments(t *testing.T, token string) []models.Document {
	t.Helper()

	resp := c.doJSON(t, http.MethodGet, "/api/documents", token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var parsed successResponse[[]models.Document]
	decodeSuccess(t, resp, &parsed)
	return parsed.Data
}

func (c *apiClient) UpdateDocument(t *testing.T, token string, id int64, name string, permission models.DocsPermission) models.Document {
	t.Helper()

	resp := c.doJSON(t, http.MethodPut, "/api/documents/"+strconv.FormatInt(id, 10), token, map[string]any{
		"name":       name,
		"permission": string(permission),
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var parsed successResponse[models.Document]
	decodeSuccess(t, resp, &parsed)
	return parsed.Data
}

func (c *apiClient) DeleteDocument(t *testing.T, token string, id int64) {
	t.Helper()

	resp := c.doJSON(t, http.MethodDelete, "/api/documents/"+strconv.FormatInt(id, 10), token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	decodeSuccess(t, resp, &successResponse[struct{}]{})
}

func (c *apiClient) ListShares(t *testing.T, token string, documentID int64) []models.DocsShare {
	t.Helper()

	resp := c.doJSON(t, http.MethodGet, "/api/documents/"+strconv.FormatInt(documentID, 10)+"/shares", token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var parsed successResponse[[]models.DocsShare]
	decodeSuccess(t, resp, &parsed)
	return parsed.Data
}

func (c *apiClient) CreateShare(t *testing.T, token string, documentID, userID int64, roles models.DocsSharePermission) models.DocsShare {
	t.Helper()

	resp := c.doJSON(t, http.MethodPost, "/api/documents/"+strconv.FormatInt(documentID, 10)+"/shares", token, map[string]any{
		"user_id": strconv.FormatInt(userID, 10),
		"roles":   string(roles),
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var parsed successResponse[models.DocsShare]
	decodeSuccess(t, resp, &parsed)
	return parsed.Data
}

func (c *apiClient) UpdateShare(t *testing.T, token string, documentID, shareID int64, roles models.DocsSharePermission) models.DocsShare {
	t.Helper()

	resp := c.doJSON(t, http.MethodPut, "/api/documents/"+strconv.FormatInt(documentID, 10)+"/shares/"+strconv.FormatInt(shareID, 10), token, map[string]any{
		"roles": string(roles),
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var parsed successResponse[models.DocsShare]
	decodeSuccess(t, resp, &parsed)
	return parsed.Data
}

func (c *apiClient) DeleteShare(t *testing.T, token string, documentID, shareID int64) {
	t.Helper()

	resp := c.doJSON(t, http.MethodDelete, "/api/documents/"+strconv.FormatInt(documentID, 10)+"/shares/"+strconv.FormatInt(shareID, 10), token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	decodeSuccess(t, resp, &successResponse[struct{}]{})
}

func initApp(t *testing.T, ctx context.Context) (*pgxpool.Pool, *httptest.Server, *docManagerStub) {
	t.Helper()

	restoreCwd := switchToProjectRoot(t)
	t.Cleanup(restoreCwd)

	docStub := startDocManagerStub(t)
	pg := startPostgresContainer(t, ctx)
	setTestEnv(t, ctx, pg, docStub)

	logger.InitLogger()
	config.ResetForTests()
	_, err := config.InitConfig()
	require.NoError(t, err)
	require.NoError(t, id.Init())

	pool, err := db.InitDatabase()
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	server := startAppServer(t, pool)
	return pool, server, docStub
}

func getUserIDByEmail(t *testing.T, pool *pgxpool.Pool, email string) int64 {
	t.Helper()

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	user, err := repository.GetUserByEmail(ctx, tx, email)
	require.NoError(t, err)
	require.NotNil(t, user, "user not found for email %s", email)

	require.NoError(t, tx.Commit(ctx))

	return user.ID
}

func waitForDocEdit(t *testing.T, stub *docManagerStub) {
	t.Helper()

	select {
	case <-stub.editCh:
	case <-time.After(2 * time.Second):
		t.Fatal("document edit endpoint was not called")
	}
}
