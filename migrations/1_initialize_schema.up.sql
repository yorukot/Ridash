CREATE SCHEMA IF NOT EXISTS "public";

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
    "provider" character varying(50) NOT NULL,
    "created_at" timestamp NOT NULL,
    "updated_at" timestamp NOT NULL,
    PRIMARY KEY ("account_id")
);
-- Indexes
CREATE INDEX "oauth_tokens_idx_oauth_tokens_provider" ON "public"."oauth_tokens" ("provider");

CREATE TABLE "public"."accounts" (
    "id" bigint NOT NULL,
    "provider" character varying(50) NOT NULL,
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

-- Foreign key constraints
-- Schema: public
ALTER TABLE "public"."accounts" ADD CONSTRAINT "fk_accounts_user_id_users_id" FOREIGN KEY("user_id") REFERENCES "public"."users"("id");
ALTER TABLE "public"."oauth_tokens" ADD CONSTRAINT "fk_oauth_tokens_account_id_accounts_id" FOREIGN KEY("account_id") REFERENCES "public"."accounts"("id");
ALTER TABLE "public"."refresh_tokens" ADD CONSTRAINT "fk_refresh_tokens_user_id_users_id" FOREIGN KEY("user_id") REFERENCES "public"."users"("id");