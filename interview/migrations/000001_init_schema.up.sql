CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE vacancies (
                           id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                           title        TEXT      NOT NULL,
                           description  TEXT,
                           users_id     UUID      NOT NULL,
                           created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                           updated_at   TIMESTAMPTZ,
                           file_url     TEXT,
                           weight_soft  INT NOT NULL DEFAULT 33 CHECK (weight_soft BETWEEN 0 AND 100),
                           weight_hard  INT NOT NULL DEFAULT 33 CHECK (weight_hard BETWEEN 0 AND 100),
                           weight_case  INT NOT NULL DEFAULT 34 CHECK (weight_case BETWEEN 0 AND 100),
                           text_jsonb   JSONB
);

CREATE TABLE resumes (
                         id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                         vacancy_id   UUID NOT NULL REFERENCES vacancies(id) ON DELETE CASCADE,
                         file_url     TEXT,
                         text_jsonb   JSONB,
                         created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                         result_jsonb JSONB
);

CREATE TABLE interviews (
                            id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                            resume_id    UUID REFERENCES resumes(id)    ON DELETE SET NULL,
                            vacancy_id   UUID NOT NULL REFERENCES vacancies(id) ON DELETE CASCADE,
                            status       TEXT  NOT NULL DEFAULT 'pending',
                            text_jsonb   JSONB,
                            url_token    TEXT UNIQUE,
                            score_jsonb  JSONB,
                            created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                            date_start   TIMESTAMPTZ,
                            updated_at   TIMESTAMPTZ,
                            started_at   TIMESTAMPTZ
);

CREATE INDEX idx_interviews_vacancy ON interviews(vacancy_id);
CREATE INDEX idx_interviews_token   ON interviews(url_token);
