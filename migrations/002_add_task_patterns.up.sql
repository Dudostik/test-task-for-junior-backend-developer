CREATE TABLE IF NOT EXISTS task_patterns (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL CHECK (type IN ('daily', 'monthly', 'specific', 'even_odd')),
    interval_days INT,
    day_of_month INT CHECK (day_of_month BETWEEN 1 AND 30),
    even_odd_type TEXT CHECK (even_odd_type IN ('even', 'odd')),
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pattern_specific_dates (
    id BIGSERIAL PRIMARY KEY,
    pattern_id BIGINT NOT NULL REFERENCES task_patterns(id) ON DELETE CASCADE,
    specific_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pattern_specific_dates_pattern_id ON pattern_specific_dates(pattern_id);

ALTER TABLE tasks ADD COLUMN pattern_id BIGINT REFERENCES task_patterns(id) ON DELETE SET NULL;
ALTER TABLE tasks ADD COLUMN is_generated BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE tasks ADD COLUMN original_date DATE;
ALTER TABLE tasks ADD COLUMN parent_task_id BIGINT REFERENCES tasks(id) ON DELETE SET NULL;

CREATE INDEX idx_tasks_pattern_id ON tasks(pattern_id);
CREATE INDEX idx_tasks_original_date ON tasks(original_date);
CREATE INDEX idx_tasks_parent_task_id ON tasks(parent_task_id);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_task_patterns_updated_at
    BEFORE UPDATE ON task_patterns
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();