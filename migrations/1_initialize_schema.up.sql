CREATE SCHEMA IF NOT EXISTS "public";

CREATE TYPE "auth_provider" AS ENUM ('email', 'google');
CREATE TYPE "docs_permission" AS ENUM ('private', 'public', 'public_write');
CREATE TYPE "docs_share_permission" AS ENUM ('read', 'write');
CREATE TYPE "role" AS ENUM ('owner', 'admin', 'member');

CREATE TABLE "public"."docs_shares" (
    "id" bigint NOT NULL,
    "document_id" bigint NOT NULL,
    "user_id" bigint NOT NULL,
    "roles" docs_share_permission,
    PRIMARY KEY ("id")
);

CREATE TABLE "public"."documents" (
    "id" bigint NOT NULL,
    "folder_id" bigint NOT NULL,
    "name" text NOT NULL,
    "premission" docs_permission NOT NULL,
    "created_at" timestamp NOT NULL,
    "updated_at" timestamp NOT NULL,
    CONSTRAINT "pk_documents_id" PRIMARY KEY ("id")
);

CREATE TABLE "public"."refresh_tokens" (
    "id" bigint NOT NULL,
    "user_id" bigint NOT NULL,
    "token" text NOT NULL UNIQUE,
    "user_agent" text,
    "ip" inet,
    "used_at" timestamp,
    "created_at" timestamp NOT NULL,
    PRIMARY KEY ("id")
);
-- Indexes
CREATE UNIQUE INDEX "refresh_tokens_refresh_tokens_token_key" ON "public"."refresh_tokens" ("token");
CREATE INDEX "refresh_tokens_idx_refresh_tokens_token" ON "public"."refresh_tokens" ("token");
CREATE INDEX "refresh_tokens_idx_refresh_tokens_created_at" ON "public"."refresh_tokens" ("created_at");
CREATE INDEX "refresh_tokens_idx_refresh_tokens_user_id" ON "public"."refresh_tokens" ("user_id");

CREATE TABLE "public"."oauth_tokens" (
    "account_id" bigint NOT NULL,
    "access_token" text NOT NULL,
    "refresh_token" text,
    "expiry" timestamp NOT NULL,
    "token_type" character varying(50) NOT NULL,
    "provider" auth_provider NOT NULL,
    "created_at" timestamp NOT NULL,
    "updated_at" timestamp NOT NULL,
    CONSTRAINT "pk_oauth_tokens_account_id" PRIMARY KEY ("account_id")
);
-- Indexes
CREATE INDEX "oauth_tokens_idx_oauth_tokens_provider" ON "public"."oauth_tokens" ("provider");

CREATE TABLE "public"."accounts" (
    "id" bigint NOT NULL,
    "provider" auth_provider NOT NULL,
    "provider_user_id" character varying(255) NOT NULL,
    "user_id" bigint NOT NULL,
    "email" character varying(255) NOT NULL,
    "created_at" timestamp NOT NULL,
    "updated_at" timestamp NOT NULL,
    PRIMARY KEY ("id")
);
-- Indexes
CREATE UNIQUE INDEX "accounts_accounts_provider_provider_user_id_key" ON "public"."accounts" ("provider", "provider_user_id");
CREATE INDEX "accounts_idx_accounts_provider" ON "public"."accounts" ("provider");
CREATE UNIQUE INDEX "accounts_accounts_provider_email_key" ON "public"."accounts" ("provider", "email");
CREATE INDEX "accounts_idx_accounts_email" ON "public"."accounts" ("email");
CREATE INDEX "accounts_idx_accounts_user_id" ON "public"."accounts" ("user_id");

CREATE TABLE "public"."users" (
    "id" bigint NOT NULL,
    "password_hash" text,
    "avatar" text,
    "display_name" text NOT NULL,
    "created_at" timestamp with time zone NOT NULL,
    "updated_at" timestamp with time zone NOT NULL,
    PRIMARY KEY ("id")
);

CREATE TABLE "public"."folders" (
    "id" bigint NOT NULL,
    "team_id" bigint NOT NULL,
    "name" text NOT NULL,
    "parent_folder" bigint,
    "created_at" timestamp NOT NULL,
    "updated_at" timestamp NOT NULL,
    PRIMARY KEY ("id")
);

CREATE TABLE "public"."team_members" (
    "id" bigint NOT NULL,
    "team_id" bigint NOT NULL,
    "user_id" bigint NOT NULL,
    "role" role NOT NULL,
    "updated_at" timestamp NOT NULL,
    "created_at" timestamp NOT NULL,
    PRIMARY KEY ("id")
);

CREATE TABLE "public"."teams" (
    "id" bigint NOT NULL,
    "owner_id" bigint NOT NULL,
    "name" varchar(50) NOT NULL,
    "updated_at" timestamp NOT NULL,
    "created_at" timestamp NOT NULL,
    PRIMARY KEY ("id")
);

-- Foreign key constraints
-- Schema: public
ALTER TABLE "public"."accounts" ADD CONSTRAINT "fk_accounts_user_id_users_id" FOREIGN KEY("user_id") REFERENCES "public"."users"("id");
ALTER TABLE "public"."docs_shares" ADD CONSTRAINT "fk_docs_shares_document_id_documents_id" FOREIGN KEY("document_id") REFERENCES "public"."documents"("id");
ALTER TABLE "public"."documents" ADD CONSTRAINT "fk_documents_folder_id_folders_id" FOREIGN KEY("folder_id") REFERENCES "public"."folders"("id");
ALTER TABLE "public"."folders" ADD CONSTRAINT "fk_folders_team_id_teams_id" FOREIGN KEY("team_id") REFERENCES "public"."teams"("id");
ALTER TABLE "public"."oauth_tokens" ADD CONSTRAINT "fk_oauth_tokens_account_id_accounts_id" FOREIGN KEY("account_id") REFERENCES "public"."accounts"("id");
ALTER TABLE "public"."refresh_tokens" ADD CONSTRAINT "fk_refresh_tokens_user_id_users_id" FOREIGN KEY("user_id") REFERENCES "public"."users"("id");
ALTER TABLE "public"."team_members" ADD CONSTRAINT "fk_team_members_team_id_teams_id" FOREIGN KEY("team_id") REFERENCES "public"."teams"("id");
ALTER TABLE "public"."team_members" ADD CONSTRAINT "fk_team_members_user_id_users_id" FOREIGN KEY("user_id") REFERENCES "public"."users"("id");
ALTER TABLE "public"."teams" ADD CONSTRAINT "fk_teams_owner_id_users_id" FOREIGN KEY("owner_id") REFERENCES "public"."users"("id");
