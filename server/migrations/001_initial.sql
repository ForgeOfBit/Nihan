-- Nihan E2EE Messaging Application - Initial Database Schema
-- Migration: 001_initial.sql

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "citext";

-- ============================================================
-- USERS TABLE
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username        CITEXT NOT NULL,
    discriminator   CHAR(4) NOT NULL DEFAULT '0001',
    email           CITEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,
    display_name    VARCHAR(64),
    avatar_url      TEXT,
    is_premium      BOOLEAN NOT NULL DEFAULT FALSE,
    status          VARCHAR(20) NOT NULL DEFAULT 'offline',
    bio             VARCHAR(256),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Each username + discriminator pair must be unique
    CONSTRAINT uq_users_username_discriminator UNIQUE (username, discriminator),

    -- Discriminator must be between 0001 and 9999
    CONSTRAINT chk_discriminator CHECK (
        discriminator ~ '^[0-9]{4}$' AND discriminator <> '0000'
    ),

    -- Status must be one of the allowed values
    CONSTRAINT chk_status CHECK (
        status IN ('online', 'offline', 'idle', 'dnd', 'invisible')
    )
);

-- Index for quick email lookup during login
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

-- Index for searching users by username
CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);

-- ============================================================
-- KEY BUNDLES TABLE (Signal Protocol key bundles for E2EE)
-- ============================================================
CREATE TABLE IF NOT EXISTS key_bundles (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    identity_key        TEXT NOT NULL,
    signed_pre_key      TEXT NOT NULL,
    signed_pre_key_sig  TEXT NOT NULL,
    one_time_pre_keys   JSONB NOT NULL DEFAULT '[]'::JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_key_bundles_user_id ON key_bundles (user_id);

-- ============================================================
-- MESSAGES TABLE (stores encrypted messages)
-- ============================================================
CREATE TABLE IF NOT EXISTS messages (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sender_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ciphertext      TEXT NOT NULL,
    nonce           TEXT NOT NULL,
    ephemeral_key   TEXT,
    message_type    VARCHAR(20) NOT NULL DEFAULT 'text',
    is_read         BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Sender and receiver must be different
    CONSTRAINT chk_no_self_message CHECK (sender_id <> receiver_id),

    -- Message type must be one of the allowed values
    CONSTRAINT chk_message_type CHECK (
        message_type IN ('text', 'image', 'file', 'key_exchange')
    )
);

-- Indexes for retrieving conversation history efficiently
CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages (sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_receiver_id ON messages (receiver_id);
CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages (
    (LEAST(sender_id, receiver_id)),
    (GREATEST(sender_id, receiver_id)),
    created_at DESC
);
CREATE INDEX IF NOT EXISTS idx_messages_unread ON messages (receiver_id, is_read) WHERE NOT is_read;

-- ============================================================
-- FRIENDSHIPS TABLE
-- ============================================================
CREATE TABLE IF NOT EXISTS friendships (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    requester_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    addressee_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Cannot send a friend request to yourself
    CONSTRAINT chk_no_self_friendship CHECK (requester_id <> addressee_id),

    -- Status must be one of the allowed values
    CONSTRAINT chk_friendship_status CHECK (
        status IN ('pending', 'accepted', 'blocked')
    )
);

CREATE INDEX IF NOT EXISTS idx_friendships_requester ON friendships (requester_id);
CREATE INDEX IF NOT EXISTS idx_friendships_addressee ON friendships (addressee_id);
CREATE INDEX IF NOT EXISTS idx_friendships_status ON friendships (status);

-- Only one friendship record per user pair
CREATE UNIQUE INDEX IF NOT EXISTS uq_friendship_pair ON friendships (
    (LEAST(requester_id, addressee_id)),
    (GREATEST(requester_id, addressee_id))
);

-- ============================================================
-- UPDATED_AT TRIGGER FUNCTION
-- ============================================================
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply the updated_at trigger to relevant tables
CREATE TRIGGER set_updated_at_users
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_updated_at_key_bundles
    BEFORE UPDATE ON key_bundles
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_updated_at_friendships
    BEFORE UPDATE ON friendships
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();
