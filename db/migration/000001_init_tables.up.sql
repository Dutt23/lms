PRAGMA foreign_keys = 1;
PRAGMA foreign_keys = ON;

CREATE TABLE "books" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT,
  "title" varchar NOT NULL,
  "author" varchar NOT NULL,
  "published_date" timestamp NOT NULL,
  "isbn" varchar UNIQUE NOT NULL,
  "number_of_pages" bigint NOT NULL,
  "cover_image" varchar,
  "language" varchar NOT NULL,
  "available_copies" bigint NOT NULL CHECK ("available_copies" >= 0)
);

CREATE TABLE "members" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT,
  "name" bigint NOT NULL,
  "email" bigint NOT NULL,
  "join_date" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "book_loans" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT,
  "book_id" bigint NOT NULL,
  "member_id" bigint NOT NULL,
  "loan_date" timestamp NOT NULL DEFAULT (now()),
  "return_date" timestamp,
  FOREIGN KEY ("book_id") REFERENCES "books" ("id"),
  FOREIGN KEY ("member_id") REFERENCES "members" ("id")
);

CREATE INDEX "book_author_idx" ON "books" ("author");
CREATE INDEX "book_title_idx" ON "books" ("title");
CREATE INDEX "book_published_idx" ON "books" ("published_date");

CREATE INDEX "members_name_idx" ON "members" ("name");
CREATE INDEX "members_email_idx" ON "members" ("email");

CREATE INDEX  "book_loans_book_id_idx" ON "book_loans" ("book_id");
CREATE INDEX  "book_loans_member_id_idx" ON "book_loans" ("member_id");
CREATE INDEX  "book_loans_loan_date_idx" ON "book_loans" ("loan_date");