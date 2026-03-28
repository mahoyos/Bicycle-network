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

-- Pending deletes: stores delete requests for bicycles currently rented
CREATE TABLE IF NOT EXISTS pending_deletes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bicycle_id UUID NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_pending_deletes_bicycle_id ON pending_deletes(bicycle_id);
CREATE INDEX IF NOT EXISTS idx_pending_deletes_processed ON pending_deletes(processed) WHERE processed = FALSE;
