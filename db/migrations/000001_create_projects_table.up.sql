CREATE TABLE IF NOT EXISTS projects (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR (100) NOT NULL,
    num_of_files INT DEFAULT 0,
    detected_vulns INT DEFAULT 0,
    scan_time NUMERIC(5, 2) DEFAULT 0.0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);