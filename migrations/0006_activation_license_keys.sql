ALTER TABLE license_keys ADD COLUMN IF NOT EXISTS key_hash VARCHAR(120);
ALTER TABLE license_keys ADD COLUMN IF NOT EXISTS plan_code VARCHAR(50) NOT NULL DEFAULT 'yearly';
ALTER TABLE license_keys ADD COLUMN IF NOT EXISTS max_operations BIGINT NOT NULL DEFAULT 250;
ALTER TABLE license_keys ADD COLUMN IF NOT EXISTS used_at TIMESTAMPTZ;
ALTER TABLE license_keys ADD COLUMN IF NOT EXISTS tenant_id BIGINT REFERENCES tenants(id);
ALTER TABLE license_keys ADD COLUMN IF NOT EXISTS status VARCHAR(30) NOT NULL DEFAULT 'active';
ALTER TABLE license_keys ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ;

UPDATE license_keys
SET key_hash = LOWER(key)
WHERE (key_hash IS NULL OR key_hash = '')
  AND key IS NOT NULL
  AND key <> ''
  AND key ~* '^[0-9a-f]{64}$';

UPDATE license_keys
SET key_hash = 'legacy-invalid-' || id,
    status = 'legacy_invalid'
WHERE (key_hash IS NULL OR key_hash = '');

ALTER TABLE license_keys ALTER COLUMN key_hash SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_license_keys_key_hash ON license_keys(key_hash);
