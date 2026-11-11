-- Extension UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. ENUM LOG CATEGORY
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'log_category') THEN
        CREATE TYPE log_category AS ENUM ('CODE', 'MEETING', 'BUGFIX', 'RESEARCH', 'DEPLOY', 'TASK', 'MISC');
    END IF;
END$$;

-- 2. TABEL ACCOUNTS
CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    telegram_chat_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(100),
    github_username VARCHAR(100),
    github_token TEXT,
    timezone VARCHAR(50) NOT NULL DEFAULT 'Asia/Jakarta',
    standup_hour INT NOT NULL DEFAULT 8,
    standup_minute INT NOT NULL DEFAULT 30,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. TABEL MONITORED_REPOS
CREATE TABLE IF NOT EXISTS monitored_repos (
    id SERIAL PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    repo_name VARCHAR(200) NOT NULL,
    branch VARCHAR(100) NOT NULL DEFAULT 'main',
    last_synced_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(account_id, repo_name)
);

-- 4. TABEL WORK_LOGS
CREATE TABLE IF NOT EXISTS work_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    category log_category NOT NULL DEFAULT 'TASK',
    source VARCHAR(50) NOT NULL DEFAULT 'MANUAL',
    repo_name VARCHAR(200),
    title VARCHAR(255) NOT NULL,
    details TEXT,
    commit_hash VARCHAR(64),
    log_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_work_logs_acc_date ON work_logs(account_id, log_date DESC);
CREATE INDEX IF NOT EXISTS idx_work_logs_acc_cat ON work_logs(account_id, category);
CREATE UNIQUE INDEX IF NOT EXISTS idx_work_logs_unique_commit ON work_logs(account_id, commit_hash) WHERE commit_hash IS NOT NULL AND commit_hash != '';

-- 5. TABEL TODOS
CREATE TABLE IF NOT EXISTS todos (
    id SERIAL PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 6. TABEL BLOCKERS
CREATE TABLE IF NOT EXISTS blockers (
    id SERIAL PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    is_resolved BOOLEAN NOT NULL DEFAULT FALSE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
