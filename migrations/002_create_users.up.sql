CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    team_name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    FOREIGN KEY (team_name) REFERENCES teams(name) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_users_team ON users(team_name);