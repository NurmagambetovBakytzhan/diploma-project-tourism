SET search_path TO tourism;

CREATE EXTENSION IF NOT EXISTS vector;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS tour_embeddings
(
    tour_id   uuid PRIMARY KEY,
    embedding vector(384)
);

SET maintenance_work_mem TO '64 GB';

CREATE INDEX ON tour_embeddings
    USING ivfflat (embedding vector_cosine_ops);
