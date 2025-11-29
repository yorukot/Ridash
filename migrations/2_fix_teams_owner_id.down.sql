ALTER TABLE "public"."teams" DROP CONSTRAINT IF EXISTS "fk_teams_owner_id_users_id";
ALTER TABLE "public"."teams" RENAME COLUMN "owner_id" TO "onwer_id";
ALTER TABLE "public"."teams" ADD CONSTRAINT "fk_teams_onwer_id_users_id" FOREIGN KEY ("onwer_id") REFERENCES "public"."users"("id");
