CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS rentals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    bicycle_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'finalized', 'cancelled')),
    start_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_time TIMESTAMPTZ,
    duration INTERVAL
);

CREATE INDEX IF NOT EXISTS idx_rentals_user_id ON rentals(user_id);
CREATE INDEX IF NOT EXISTS idx_rentals_status ON rentals(status);
CREATE INDEX IF NOT EXISTS idx_rentals_bicycle_id ON rentals(bicycle_id);

-- Solo puede existir un rental activo por bicicleta
CREATE UNIQUE INDEX IF NOT EXISTS idx_rentals_bicycle_active
    ON rentals(bicycle_id) WHERE status = 'active';

-- Un usuario solo puede tener un rental activo a la vez
CREATE UNIQUE INDEX IF NOT EXISTS idx_rentals_user_active
    ON rentals(user_id) WHERE status = 'active';
