-- database/init.sql

-- ТАБЛИЦА ПОЛЬЗОВАТЕЛЕЙ
CREATE TABLE IF NOT EXISTS users (
                                     id SERIAL PRIMARY KEY,
                                     username VARCHAR(100) NOT NULL,
                                     surname VARCHAR(100) NOT NULL,
                                     email VARCHAR(255) NOT NULL UNIQUE,
                                     password_hash VARCHAR(255) NOT NULL,
                                     role VARCHAR(255) NOT NULL CHECK (role IN ('hr_specialist','candidate','admin')),
                                     is_active BOOLEAN NOT NULL DEFAULT TRUE,
                                     created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                     updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- ТАБЛИЦА ТОКЕНОВ
CREATE TABLE IF NOT EXISTS tokens (
                                      id SERIAL PRIMARY KEY,
                                      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                      token_hash VARCHAR(255) NOT NULL UNIQUE,
                                      expires_at TIMESTAMPTZ NOT NULL,
                                      is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
                                      created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                      revoked_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_tokens_user_id ON tokens(user_id);

-- ТАБЛИЦА ВАКАНСИЙ
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS vacancies (
                                         id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                         title TEXT NOT NULL,
                                         description TEXT,
                                         users_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                         created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                         updated_at TIMESTAMPTZ,
                                         file_url TEXT,
                                         storage_key VARCHAR(255),
                                         weight_soft INT NOT NULL DEFAULT 33,
                                         weight_hard INT NOT NULL DEFAULT 33,
                                         weight_case INT NOT NULL DEFAULT 34,
                                         text_jsonb JSONB
);
CREATE INDEX IF NOT EXISTS idx_vacancies_storage_key ON vacancies(storage_key);

-- ТАБЛИЦА РЕЗЮМЕ
CREATE TABLE IF NOT EXISTS resumes (
                                       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                       vacancy_id UUID NOT NULL REFERENCES vacancies(id) ON DELETE CASCADE,
                                       file_url TEXT,
                                       storage_key VARCHAR(255),
                                       text TEXT,
                                       status TEXT,
                                       mail TEXT,
                                       created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                       result_jsonb JSONB,
                                       resume_analysis_jsonb JSONB

);
CREATE INDEX IF NOT EXISTS idx_resumes_storage_key ON resumes(storage_key);

-- ТАБЛИЦА ИНТЕРВЬЮ
CREATE TABLE IF NOT EXISTS interviews (
                                          id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                          resume_id UUID REFERENCES resumes(id) ON DELETE SET NULL,
                                          vacancy_id UUID NOT NULL REFERENCES vacancies(id) ON DELETE CASCADE,
                                          status TEXT NOT NULL DEFAULT 'pending',
                                          text_jsonb JSONB,
                                          url_token TEXT UNIQUE,
                                          score_jsonb JSONB,
                                          created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                          date_start TIMESTAMPTZ,
                                          updated_at TIMESTAMPTZ,
                                          started_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_interviews_vacancy_id ON interviews(vacancy_id);