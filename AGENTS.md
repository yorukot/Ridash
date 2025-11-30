# Repository Guidelines

## Project Structure & Module Organization
- `cmd/main.go`: entrypoint wiring Echo server, routes, logging, and config.
- `handler/`, `router/`, `middleware/`: HTTP handlers, grouped routing, and cross-cutting concerns.
- `models/`, `repository/`: data structures and DB access; `db/` manages the pgx pool plus migrations; `migrations/*.sql` store schema changes.
- `utils/` holds config, ID generator, and logger helpers; `docs/` contains generated Swagger artifacts (regenerated via `make generate-docs`).
- `compose.yaml` + `db/` help run Postgres locally; `tmp/` stores build outputs and should stay untracked.

## Build, Test, and Development Commands
- `make build`: compile the API to `tmp/radish`.
- `make run`: build then start the server on `$APP_PORT` (default 8000).
- `make dev`: live reload with Air (install `air` locally).
- `make lint`: run `gofmt`, `go vet`, and `golint`.
- `make generate-docs`: regenerate Swagger docs from annotations in `cmd/main.go` and handlers.
- `go test ./...`: execute all tests (add as you build coverage).

## Coding Style & Naming Conventions
- Keep code `gofmt` clean (tabs, standard Go brace style); lint before committing.
- Mirror package boundaries to folders above; avoid cyclic dependencies and keep handlers thin by delegating to repositories/services.
- Request/response structs use exported PascalCase fields with JSON tags; helpers stay camelCase unexported.
- Wrap errors with context and prefer structured logging via `zap.L()` and the middleware logger.

## Testing Guidelines
- Place `_test.go` files beside the code; prefer table-driven cases.
- Cover handlers with Echo test utilities; mock pgx repositories to avoid real DB when unit testing.
- Migrations run automatically on startup via golang-migrate—ensure new migrations are reversible and named sequentially.
- Require tests for bug fixes and new endpoints; target meaningful scenario coverage rather than raw percentages.

## Commit & Pull Request Guidelines
- Follow Conventional Commits (`feat(scope): ...`, `fix: ...`, `chore: ...`) as seen in history.
- PRs should describe behavior changes, migrations, env var updates, and include sample requests/responses or screenshots for API adjustments.
- Link issues when available; keep PRs small and focused; update docs (`docs/`, `AGENTS.md`) and mention if Swagger was regenerated.
- Run `make lint` and `go test ./...` before requesting review.

## Environment & Security Notes
- Copy `.env.example` to `.env`; set Postgres creds, SMTP values, Google OAuth keys, JWT secret, and document manager settings.
- DB user must allow migrations on start; use `compose.yaml` for a local stack when testing.
- Keep secrets out of commits and logs; rely on environment files or CI secrets instead.
