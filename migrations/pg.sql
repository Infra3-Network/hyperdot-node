-- DROP TABLE IF EXISTS hyperdot_query_engines;
-- CREATE TABLE IF NOT EXSITS hyperdot_query_engines (
--   id SERIAL PRIMARY KEY,
--   name VARCHAR(255) NOT NULL,
--   description TEXT,
--   connection_params JSON,
--   created_at TIMESTAMP NOT NULL DEFAULT NOW(),
--   updated_at TIMESTAMP NOT NULL DEFAULT NOW()
-- );
--
-- INSERT INTO hyperdot_query_engines (name, description, connection_params)
-- VALUES ('bigquery', 'Google BigQuery', '{"project_id": "substrate-etl"}');
--
-- DROP TABLE IF EXISTS hyperdot_dataset

DROP TABLE IF EXISTS hyperdot_users
CREATE TABLE IF NOT EXISTS hyperdot_users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL,
    password VARCHAR(100) NOT NULL
);