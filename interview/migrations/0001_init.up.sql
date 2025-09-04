CREATE TABLE "users"(
                        "id" INTEGER NOT NULL,
                        "username" VARCHAR(100) NOT NULL,
                        "surname" VARCHAR(100) NOT NULL,
                        "email" VARCHAR(255) NOT NULL,
                        "password_hash" VARCHAR(255) NOT NULL,
                        "role" VARCHAR(255) CHECK
        (
            "role" IN(
                'hr_specialist',
                'candidate',
                'admin'
            )
        ) NOT NULL,
                        "is_active" BOOLEAN NULL DEFAULT 'DEFAULT TRUE',
                        "created_at" TIMESTAMP(0)
                            WITH
        TIME zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        "updated_at" TIMESTAMP(0)
                            WITH
        TIME zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);
ALTER TABLE
    "users" ADD PRIMARY KEY("id");
ALTER TABLE
    "users" ADD CONSTRAINT "users_email_unique" UNIQUE("email");
CREATE INDEX "users_role_index" ON
    "users"("role");
CREATE TABLE "tokens"(
                         "id" INTEGER NOT NULL,
                         "user_id" INTEGER NOT NULL,
                         "token_hash" VARCHAR(255) NOT NULL,
                         "expires_at" TIMESTAMP(0) WITH
        TIME zone NOT NULL,
                         "is_revoked" BOOLEAN NULL DEFAULT 'DEFAULT FALSE',
                         "created_at" TIMESTAMP(0)
                             WITH
        TIME zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
                         "revoked_at" TIMESTAMP(0)
                             WITH
        TIME zone NULL
);
ALTER TABLE
    "tokens" ADD PRIMARY KEY("id");
CREATE INDEX "tokens_user_id_index" ON
    "tokens"("user_id");
ALTER TABLE
    "tokens" ADD CONSTRAINT "tokens_token_hash_unique" UNIQUE("token_hash");
CREATE TABLE "vacancies"(
                            "id" UUID NOT NULL DEFAULT 'DEFAULT UUID_GENERATE_V4 ( )',
                            "title" TEXT NOT NULL,
                            "description" TEXT NULL,
                            "users_id" UUID NOT NULL,
                            "created_at" TIMESTAMP(0) WITH
        TIME zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
                            "updated_at" TIMESTAMP(0)
                                WITH
        TIME zone NULL,
                            "file_url" TEXT NULL,
                            "weight_soft" INTEGER NOT NULL DEFAULT '33',
                            "weight_hard" INTEGER NOT NULL DEFAULT '33',
                            "weight_case" INTEGER NOT NULL DEFAULT '34',
                            "text_jsonb" jsonb NULL
);
ALTER TABLE
    "vacancies" ADD PRIMARY KEY("id");
CREATE TABLE "resumes"(
                          "id" UUID NOT NULL DEFAULT 'DEFAULT UUID_GENERATE_V4 ( )',
                          "vacancy_id" UUID NOT NULL,
                          "file_url" TEXT NULL,
                          "text_jsonb" jsonb NULL,
                          "status" TEXT NULL,
                          "mail" TEXT NULL,
                          "created_at" TIMESTAMP(0) WITH
        TIME zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
                          "result_jsonb" jsonb NULL
);
ALTER TABLE
    "resumes" ADD PRIMARY KEY("id");
CREATE TABLE "interviews"(
                             "id" UUID NOT NULL DEFAULT 'DEFAULT UUID_GENERATE_V4 ( )',
                             "resume_id" UUID NULL,
                             "vacancy_id" UUID NOT NULL,
                             "status" TEXT NOT NULL DEFAULT 'pending',
                             "text_jsonb" jsonb NULL,
                             "url_token" TEXT NULL,
                             "score_jsonb" jsonb NULL,
                             "created_at" TIMESTAMP(0) WITH
        TIME zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
                             "date_start" TIMESTAMP(0)
                                 WITH
        TIME zone NULL,
                             "updated_at" TIMESTAMP(0)
                                 WITH
        TIME zone NULL,
                             "started_at" TIMESTAMP(0)
                                 WITH
        TIME zone NULL
);
ALTER TABLE
    "interviews" ADD PRIMARY KEY("id");
CREATE INDEX "interviews_vacancy_id_index" ON
    "interviews"("vacancy_id");
ALTER TABLE
    "interviews" ADD CONSTRAINT "interviews_url_token_unique" UNIQUE("url_token");
ALTER TABLE
    "vacancies" ADD CONSTRAINT "vacancies_users_id_foreign" FOREIGN KEY("users_id") REFERENCES "users"("id");
ALTER TABLE
    "tokens" ADD CONSTRAINT "tokens_user_id_foreign" FOREIGN KEY("user_id") REFERENCES "users"("id");
ALTER TABLE
    "resumes" ADD CONSTRAINT "resumes_vacancy_id_foreign" FOREIGN KEY("vacancy_id") REFERENCES "vacancies"("id");
ALTER TABLE
    "interviews" ADD CONSTRAINT "interviews_vacancy_id_foreign" FOREIGN KEY("vacancy_id") REFERENCES "vacancies"("id");
ALTER TABLE
    "interviews" ADD CONSTRAINT "interviews_resume_id_foreign" FOREIGN KEY("resume_id") REFERENCES "resumes"("id");