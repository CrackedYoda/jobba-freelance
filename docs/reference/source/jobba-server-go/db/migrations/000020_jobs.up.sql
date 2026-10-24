CREATE TYPE job_status AS ENUM ('DRAFT', 'OPEN', 'DISPUTED', 'ACTIVE', 'COMPLETED', 'CANCELLED', 'CLOSED');
CREATE TYPE job_payment_method AS ENUM ('escrow', 'direct');
CREATE TYPE job_work_type AS ENUM ('onsite', 'remote');
CREATE TYPE job_applicant_type AS ENUM ('all', 'top-rated', 'invite');
CREATE TYPE job_privacy AS ENUM ('public', 'private');

CREATE TABLE jobs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    legacy_id text UNIQUE,
    client_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    job_title text NOT NULL CHECK (char_length(btrim(job_title)) BETWEEN 5 AND 100),
    category text NOT NULL CHECK (btrim(category) <> ''),
    description text NOT NULL CHECK (char_length(btrim(description)) BETWEEN 50 AND 2000),
    attachments text[] NOT NULL DEFAULT '{}',
    min_budget numeric(14,2) NOT NULL CHECK (min_budget >= 0),
    max_budget numeric(14,2) NOT NULL CHECK (max_budget >= min_budget),
    payment_method job_payment_method NOT NULL,
    start_date timestamptz NOT NULL,
    end_date timestamptz NOT NULL CHECK (end_date > start_date),
    work_type job_work_type NOT NULL,
    location text NOT NULL DEFAULT '',
    applicant_type job_applicant_type NOT NULL,
    privacy job_privacy NOT NULL,
    status job_status NOT NULL DEFAULT 'OPEN',
    assigned_worker_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    budget numeric(14,2) CHECK (budget >= 0),
    view_count bigint NOT NULL DEFAULT 0 CHECK (view_count >= 0),
    is_deleted boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT jobs_onsite_location CHECK (work_type <> 'onsite' OR btrim(location) <> ''),
    CONSTRAINT jobs_assignment_budget CHECK (budget IS NULL OR assigned_worker_id IS NOT NULL)
);

CREATE TABLE job_skills (
    job_id uuid NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    skill_id uuid NOT NULL REFERENCES skills(id) ON DELETE RESTRICT,
    PRIMARY KEY (job_id, skill_id)
);

CREATE TABLE saved_jobs (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id uuid NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, job_id)
);

CREATE INDEX jobs_client_status_idx ON jobs (client_id, status) WHERE is_deleted = false;
CREATE INDEX jobs_worker_status_idx ON jobs (assigned_worker_id, status) WHERE is_deleted = false;
CREATE INDEX jobs_public_category_idx ON jobs (category, created_at DESC) WHERE status = 'OPEN' AND privacy = 'public' AND is_deleted = false;
CREATE INDEX job_skills_skill_idx ON job_skills (skill_id, job_id);
CREATE INDEX saved_jobs_user_created_idx ON saved_jobs (user_id, created_at DESC);
