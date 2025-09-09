-- Seed data for development/testing

-- Insert a dummy user
INSERT INTO users (id, email, created_at, updated_at) VALUES
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'test@example.com', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Insert a dummy repository for the user
INSERT INTO repos (id, user_id, owner, name, etag, last_checked_at, created_at, updated_at) VALUES
    ('b1cde0f1-1234-5678-90ab-cdef01234567', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'octocat', 'Spoon-Knife', '', NOW() - INTERVAL '1 day', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Insert a dummy subscription for the user to the repository
INSERT INTO subscriptions (id, repo_id, user_id, channel, created_at, updated_at) VALUES
    ('c2def102-abcd-efab-cdef-123456789012', 'b1cde0f1-1234-5678-90ab-cdef01234567', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'some_telegram_chat_id', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
