ALTER TABLE "public"."teams" DROP CONSTRAINT IF EXISTS "fk_teams_onwer_id_users_id";
ALTER TABLE "public"."teams" RENAME COLUMN "onwer_id" TO "owner_id";
ALTER TABLE "public"."teams" ADD CONSTRAINT "fk_teams_owner_id_users_id" FOREIGN KEY ("owner_id") REFERENCES "public"."users"("id");
