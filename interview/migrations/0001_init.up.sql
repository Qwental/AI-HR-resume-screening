-- Расширение для UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица вакансий
CREATE TABLE IF NOT EXISTS vacancies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title TEXT NOT NULL,
    description TEXT,
    users_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    file_url TEXT,
    weight_soft INT DEFAULT 30 CHECK (weight_soft >= 0 AND weight_soft <= 100),
    weight_hard INT DEFAULT 50 CHECK (weight_hard >= 0 AND weight_hard <= 100),
    weight_case INT DEFAULT 20 CHECK (weight_case >= 0 AND weight_case <= 100),
    text_jsonb JSONB
    );

CREATE INDEX IF NOT EXISTS idx_vacancies_users_id ON vacancies(users_id);

-- Таблица резюме
CREATE TABLE IF NOT EXISTS resumes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vacancy_id UUID NOT NULL,
    file_url TEXT,
    text_jsonb JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_resumes_vacancy_id ON resumes(vacancy_id);

-- Таблица собеседований
CREATE TABLE IF NOT EXISTS interviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resume_id UUID,
    vacancy_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    text_jsonb JSONB,
    audio_url TEXT,
    url_token TEXT UNIQUE,
    score_jsonb JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    date_start TIMESTAMP,
    updated_at TIMESTAMP
    );

CREATE INDEX IF NOT EXISTS idx_interviews_vacancy_id ON interviews(vacancy_id);
CREATE INDEX IF NOT EXISTS idx_interviews_resume_id ON interviews(resume_id);
