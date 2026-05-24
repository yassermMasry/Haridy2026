ALTER TABLE receipt_vouchers ADD COLUMN IF NOT EXISTS tenant_id BIGINT REFERENCES tenants(id);
ALTER TABLE payment_vouchers ADD COLUMN IF NOT EXISTS tenant_id BIGINT REFERENCES tenants(id);

CREATE INDEX IF NOT EXISTS idx_receipt_vouchers_tenant_id ON receipt_vouchers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_payment_vouchers_tenant_id ON payment_vouchers(tenant_id);
