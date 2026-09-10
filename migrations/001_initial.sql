CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    display_name TEXT NOT NULL CHECK (char_length(display_name) BETWEEN 1 AND 100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rooms (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 120),
    kind TEXT NOT NULL CHECK (kind IN ('direct', 'group')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS room_members (
    room_id TEXT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (room_id, user_id)
);

CREATE TABLE IF NOT EXISTS messages (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    room_id TEXT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    sender_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    body TEXT NOT NULL CHECK (char_length(body) BETWEEN 1 AND 4000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS messages_room_history_idx
    ON messages (room_id, created_at DESC, id DESC);

INSERT INTO users (id, display_name) VALUES
    ('demo-alice', 'Alice'),
    ('demo-bob', 'Bob'),
    ('demo-outsider', 'Outsider')
ON CONFLICT (id) DO NOTHING;

INSERT INTO rooms (id, name, kind)
VALUES ('demo-room', 'Demo Room', 'group')
ON CONFLICT (id) DO NOTHING;

INSERT INTO room_members (room_id, user_id) VALUES
    ('demo-room', 'demo-alice'),
    ('demo-room', 'demo-bob')
ON CONFLICT DO NOTHING;
