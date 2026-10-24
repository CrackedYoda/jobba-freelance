CREATE TYPE application_status AS ENUM ('PENDING', 'SHORTLISTED', 'REJECTED', 'ACCEPTED');

CREATE TABLE applications (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    legacy_id text UNIQUE,
    job_id uuid NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    worker_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    cover_letter text NOT NULL CHECK (char_length(btrim(cover_letter)) >= 10),
    bid_amount numeric(14,2) NOT NULL CHECK (bid_amount >= 1),
    work_duration text NOT NULL CHECK (btrim(work_duration) <> ''),
    status application_status NOT NULL DEFAULT 'PENDING',
    attachments text[] NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (job_id, worker_id)
);

CREATE INDEX applications_job_status_created_idx ON applications (job_id, status, created_at DESC);
CREATE INDEX applications_worker_status_created_idx ON applications (worker_id, status, created_at DESC);
