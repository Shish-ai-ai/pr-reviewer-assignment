CREATE TABLE teams (
                       team_name VARCHAR(100) PRIMARY KEY,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE users (
                       user_id VARCHAR(100) PRIMARY KEY,
                       username VARCHAR(100) NOT NULL,
                       team_name VARCHAR(100) NOT NULL REFERENCES teams(team_name) ON DELETE CASCADE,
                       is_active BOOLEAN NOT NULL DEFAULT true,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE pull_requests (
                               pull_request_id VARCHAR(100) PRIMARY KEY,
                               pull_request_name VARCHAR(255) NOT NULL,
                               author_id VARCHAR(100) NOT NULL REFERENCES users(user_id),
                               status VARCHAR(20) NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
                               assigned_reviewers JSONB,
                               created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                               merged_at TIMESTAMP
);

CREATE INDEX idx_users_team_name ON users(team_name);
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_users_team_active ON users(team_name, is_active);
CREATE INDEX idx_pull_requests_author_id ON pull_requests(author_id);
CREATE INDEX idx_pull_requests_status ON pull_requests(status);