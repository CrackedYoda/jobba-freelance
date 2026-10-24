CREATE TYPE chat_message_type AS ENUM ('text', 'image', 'file', 'audio');

CREATE TABLE conversations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    legacy_id text UNIQUE,
    job_id uuid NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    participant_a uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    participant_b uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_message_id uuid,
    last_message_at timestamptz NOT NULL DEFAULT now(),
    is_active boolean NOT NULL DEFAULT true,
    is_urgent boolean NOT NULL DEFAULT false,
    marked_urgent_by uuid REFERENCES users(id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT conversations_distinct_participants CHECK (participant_a <> participant_b),
    CONSTRAINT conversations_canonical_participants CHECK (participant_a::text < participant_b::text),
    UNIQUE (job_id, participant_a, participant_b)
);

CREATE TABLE conversation_members (
    conversation_id uuid NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    archived_at timestamptz,
    read_through_at timestamptz NOT NULL DEFAULT '-infinity',
    unread_count integer NOT NULL DEFAULT 0 CHECK (unread_count >= 0),
    PRIMARY KEY (conversation_id, user_id)
);

CREATE TABLE messages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    legacy_id text UNIQUE,
    conversation_id uuid NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    receiver_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    message_type chat_message_type NOT NULL DEFAULT 'text',
    content text,
    file_url text,
    file_name text,
    file_size bigint CHECK (file_size IS NULL OR file_size >= 0),
    delivered_at timestamptz NOT NULL DEFAULT now(),
    read_at timestamptz,
    deleted_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT messages_distinct_participants CHECK (sender_id <> receiver_id),
    CONSTRAINT messages_payload CHECK (
      (message_type = 'text' AND content IS NOT NULL AND btrim(content) <> '' AND file_url IS NULL)
      OR (message_type IN ('image','file','audio') AND file_url IS NOT NULL AND btrim(file_url) <> '')
    )
);

ALTER TABLE conversations ADD CONSTRAINT conversations_last_message_fk
  FOREIGN KEY (last_message_id) REFERENCES messages(id) ON DELETE SET NULL DEFERRABLE INITIALLY DEFERRED;

CREATE INDEX conversations_participant_a_recent_idx ON conversations(participant_a, last_message_at DESC, id DESC);
CREATE INDEX conversations_participant_b_recent_idx ON conversations(participant_b, last_message_at DESC, id DESC);
CREATE INDEX conversation_members_user_recent_idx ON conversation_members(user_id, conversation_id);
CREATE INDEX messages_conversation_recent_idx ON messages(conversation_id, created_at DESC, id DESC) WHERE deleted_at IS NULL;
CREATE INDEX messages_unread_idx ON messages(conversation_id, receiver_id, created_at) WHERE read_at IS NULL AND deleted_at IS NULL;

CREATE FUNCTION chat_add_members() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  INSERT INTO conversation_members(conversation_id,user_id) VALUES(NEW.id,NEW.participant_a),(NEW.id,NEW.participant_b);
  RETURN NEW;
END $$;
CREATE TRIGGER conversations_add_members AFTER INSERT ON conversations FOR EACH ROW EXECUTE FUNCTION chat_add_members();
