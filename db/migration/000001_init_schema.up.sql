-- TODO: make owner and name a composite key

CREATE TABLE "entries" (
  "id" SERIAL,
  "owner" varchar NOT NULL,
  "name" varchar NOT NULL,
  "due_date" timestamptz NOT NULL,
  "amount" bigint NOT NULL DEFAULT 0,
  "category" varchar,
  PRIMARY KEY("owner", "name")
);

-- CREATE INDEX ON "users" ("username");

CREATE INDEX ON "entries" ("owner");

COMMENT ON COLUMN "entries"."amount" IS 'must be positive';

-- ALTER TABLE "entries" ADD FOREIGN KEY ("owner") REFERENCES "users" ("username")