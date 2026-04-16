ALTER TABLE autopilot
    ADD COLUMN IF NOT EXISTS retry_on_blocked boolean NOT NULL DEFAULT false;
