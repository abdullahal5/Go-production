CREATE TABLE listings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    title Text Not NULL,
    description TEXT NOT NULL,
    price BIGINT NOT NULL,
    city Text NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)