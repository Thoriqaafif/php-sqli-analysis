CREATE TABLE IF NOT EXISTS projects (
    id uuid PRIMARY KEY,
    name VARCHAR (100) NOT NULL,
    num_of_files INT DEFAULT 0,
    detected_vulns INT DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    scanned BOOLEAN DEFAULT false
);