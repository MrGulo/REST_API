CREATE TYPE task_status AS ENUM ('created', 'in_progress', 'completed', 'expired');

CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       username VARCHAR(255) UNIQUE NOT NULL,
                       password_hash VARCHAR(255) NOT NULL,
                       created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tasks (
                       id SERIAL PRIMARY KEY,
                       title VARCHAR(255) NOT NULL,
                       description TEXT,
                       status task_status DEFAULT 'created',
                       deadline TIMESTAMP WITH TIME ZONE NOT NULL,
                       creator_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                       responsible_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
                       created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);