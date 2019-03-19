-- migrate:up
CREATE TABLE IF NOT EXISTS contacts (
    name text, 
    email text unique
);

-- migrate:down

DROP TABLE IF EXISTS contacts;