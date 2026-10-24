CREATE TABLE saved_workers (
    client_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    worker_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY(client_id, worker_id), CHECK(client_id <> worker_id)
);

CREATE TYPE offer_status AS ENUM ('PENDING','ACCEPTED','REJECTED');
CREATE TABLE offers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(), job_id uuid NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    client_id uuid NOT NULL REFERENCES users(id), worker_id uuid NOT NULL REFERENCES users(id),
    budget numeric(14,2) NOT NULL CHECK (budget >= 0), title text, description text,
    start_date timestamptz, end_date timestamptz, location text,
    status offer_status NOT NULL DEFAULT 'PENDING', created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK(end_date IS NULL OR start_date IS NULL OR end_date > start_date)
);
CREATE INDEX offers_worker_status_idx ON offers(worker_id,status,created_at DESC);
CREATE INDEX offers_job_worker_idx ON offers(job_id,worker_id);

CREATE TABLE reviews (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(), legacy_id text UNIQUE,
    job_id uuid NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    reviewer_id uuid NOT NULL REFERENCES users(id), reviewee_id uuid NOT NULL REFERENCES users(id),
    rating smallint NOT NULL CHECK(rating BETWEEN 1 AND 5), comment text CHECK(char_length(comment) <= 500),
    created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(job_id,reviewer_id), CHECK(reviewer_id <> reviewee_id)
);
CREATE INDEX reviews_reviewee_created_idx ON reviews(reviewee_id,created_at DESC);
