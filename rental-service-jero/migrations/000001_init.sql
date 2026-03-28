CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS known_bikes (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rentals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    bicycle_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'finalized', 'cancelled')),
    start_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_time TIMESTAMPTZ,
    duration_seconds INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rentals_user_id ON rentals(user_id);
CREATE INDEX IF NOT EXISTS idx_rentals_bicycle_id ON rentals(bicycle_id);
CREATE INDEX IF NOT EXISTS idx_rentals_status ON rentals(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_rentals_active_user ON rentals(user_id) WHERE status = 'active';
CREATE UNIQUE INDEX IF NOT EXISTS idx_rentals_active_bike ON rentals(bicycle_id) WHERE status = 'active';
