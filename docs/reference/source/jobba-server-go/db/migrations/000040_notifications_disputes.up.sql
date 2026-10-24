CREATE TYPE notification_type AS ENUM ('info','warning','success','error','alert');
CREATE TABLE notifications (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), recipient_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 sender_id uuid REFERENCES users(id) ON DELETE SET NULL, title text NOT NULL CHECK (char_length(btrim(title)) BETWEEN 1 AND 200),
 message text NOT NULL CHECK (char_length(btrim(message)) BETWEEN 1 AND 2000), type notification_type NOT NULL DEFAULT 'info',
 is_read boolean NOT NULL DEFAULT false, link text, metadata jsonb NOT NULL DEFAULT '{}', created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now());
CREATE INDEX notifications_recipient_created_idx ON notifications(recipient_id,created_at DESC,id DESC);
CREATE INDEX notifications_unread_idx ON notifications(recipient_id,created_at DESC) WHERE is_read=false;

CREATE TYPE dispute_reason AS ENUM ('NON_PAYMENT','POOR_QUALITY','UNRESPONSIVE','SCOPE_CREEP','OTHER');
CREATE TYPE dispute_status AS ENUM ('OPEN','UNDER_REVIEW','RESOLVED','CLOSED_FAVOR_WORKER','CLOSED_FAVOR_CLIENT');
CREATE TABLE disputes (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), job_id uuid NOT NULL REFERENCES jobs(id) ON DELETE RESTRICT,
 initiator_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT, responder_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
 reason dispute_reason NOT NULL, description text NOT NULL CHECK (char_length(btrim(description)) BETWEEN 1 AND 3000),
 evidence_urls text[] NOT NULL DEFAULT '{}', status dispute_status NOT NULL DEFAULT 'OPEN', admin_assigned_id uuid REFERENCES users(id) ON DELETE SET NULL,
 resolution_message text CHECK (resolution_message IS NULL OR char_length(btrim(resolution_message)) BETWEEN 1 AND 5000), resolved_at timestamptz,
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(), CHECK (initiator_id<>responder_id),
 CHECK ((status IN ('RESOLVED','CLOSED_FAVOR_WORKER','CLOSED_FAVOR_CLIENT'))=(resolved_at IS NOT NULL)));
CREATE UNIQUE INDEX disputes_one_active_job_idx ON disputes(job_id) WHERE status IN ('OPEN','UNDER_REVIEW');
CREATE INDEX disputes_party_idx ON disputes(initiator_id,responder_id,created_at DESC);
CREATE INDEX disputes_admin_idx ON disputes(admin_assigned_id,status,created_at DESC);
CREATE TABLE dispute_audit_log (id bigserial PRIMARY KEY, dispute_id uuid NOT NULL REFERENCES disputes(id) ON DELETE CASCADE,
 actor_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT, action text NOT NULL CHECK (action IN ('CREATED','EVIDENCE_ADDED','ASSIGNED','RESOLVED')),
 from_status dispute_status, to_status dispute_status, details jsonb NOT NULL DEFAULT '{}', created_at timestamptz NOT NULL DEFAULT now());
CREATE INDEX dispute_audit_dispute_idx ON dispute_audit_log(dispute_id,created_at,id);
