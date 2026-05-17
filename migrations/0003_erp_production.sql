CREATE TABLE IF NOT EXISTS branches (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(140) NOT NULL UNIQUE,
  code VARCHAR(40) NOT NULL UNIQUE,
  address VARCHAR(255),
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS warehouses (
  id BIGSERIAL PRIMARY KEY,
  branch_id BIGINT NOT NULL REFERENCES branches(id),
  name VARCHAR(140) NOT NULL,
  code VARCHAR(40) NOT NULL UNIQUE,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_warehouses_branch_id ON warehouses(branch_id);

ALTER TABLE users ADD COLUMN IF NOT EXISTS current_branch_id BIGINT REFERENCES branches(id);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS branch_id BIGINT REFERENCES branches(id);
ALTER TABLE suppliers ADD COLUMN IF NOT EXISTS branch_id BIGINT REFERENCES branches(id);
ALTER TABLE treasuries ADD COLUMN IF NOT EXISTS branch_id BIGINT REFERENCES branches(id);
ALTER TABLE stock_movements ADD COLUMN IF NOT EXISTS branch_id BIGINT REFERENCES branches(id);
ALTER TABLE stock_movements ADD COLUMN IF NOT EXISTS warehouse_id BIGINT REFERENCES warehouses(id);
ALTER TABLE sales_invoices ADD COLUMN IF NOT EXISTS branch_id BIGINT REFERENCES branches(id);
ALTER TABLE sales_invoices ADD COLUMN IF NOT EXISTS warehouse_id BIGINT REFERENCES warehouses(id);
ALTER TABLE purchase_invoices ADD COLUMN IF NOT EXISTS branch_id BIGINT REFERENCES branches(id);
ALTER TABLE purchase_invoices ADD COLUMN IF NOT EXISTS warehouse_id BIGINT REFERENCES warehouses(id);

CREATE TABLE IF NOT EXISTS item_warehouse_balances (
  id BIGSERIAL PRIMARY KEY,
  item_id BIGINT NOT NULL REFERENCES items(id),
  warehouse_id BIGINT NOT NULL REFERENCES warehouses(id),
  quantity NUMERIC(14,3) NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ,
  CONSTRAINT idx_item_warehouse UNIQUE(item_id, warehouse_id)
);
CREATE INDEX IF NOT EXISTS idx_item_warehouse_balances_quantity ON item_warehouse_balances(quantity);

CREATE TABLE IF NOT EXISTS warehouse_transfers (
  id BIGSERIAL PRIMARY KEY,
  number VARCHAR(40) NOT NULL UNIQUE,
  from_warehouse_id BIGINT NOT NULL REFERENCES warehouses(id),
  to_warehouse_id BIGINT NOT NULL REFERENCES warehouses(id),
  item_id BIGINT NOT NULL REFERENCES items(id),
  quantity NUMERIC(14,3) NOT NULL,
  user_id BIGINT NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS permissions (
  id BIGSERIAL PRIMARY KEY,
  code VARCHAR(80) NOT NULL UNIQUE,
  description VARCHAR(255),
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS rbac_roles (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(80) NOT NULL UNIQUE,
  description VARCHAR(255),
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS role_permissions (
  rbac_role_id BIGINT NOT NULL REFERENCES rbac_roles(id) ON DELETE CASCADE,
  permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  PRIMARY KEY (rbac_role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS user_roles (
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  rbac_role_id BIGINT NOT NULL REFERENCES rbac_roles(id) ON DELETE CASCADE,
  PRIMARY KEY (user_id, rbac_role_id)
);

CREATE TABLE IF NOT EXISTS sales_returns (
  id BIGSERIAL PRIMARY KEY,
  number VARCHAR(40) NOT NULL UNIQUE,
  invoice_id BIGINT NOT NULL REFERENCES sales_invoices(id),
  branch_id BIGINT REFERENCES branches(id),
  warehouse_id BIGINT REFERENCES warehouses(id),
  total NUMERIC(14,2) NOT NULL DEFAULT 0,
  reason VARCHAR(255),
  user_id BIGINT NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS sales_return_items (
  id BIGSERIAL PRIMARY KEY,
  return_id BIGINT NOT NULL REFERENCES sales_returns(id) ON DELETE CASCADE,
  item_id BIGINT NOT NULL REFERENCES items(id),
  quantity NUMERIC(14,3) NOT NULL,
  unit_price NUMERIC(14,2) NOT NULL,
  total NUMERIC(14,2) NOT NULL
);

CREATE TABLE IF NOT EXISTS purchase_returns (
  id BIGSERIAL PRIMARY KEY,
  number VARCHAR(40) NOT NULL UNIQUE,
  invoice_id BIGINT NOT NULL REFERENCES purchase_invoices(id),
  branch_id BIGINT REFERENCES branches(id),
  warehouse_id BIGINT REFERENCES warehouses(id),
  total NUMERIC(14,2) NOT NULL DEFAULT 0,
  reason VARCHAR(255),
  user_id BIGINT NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS purchase_return_items (
  id BIGSERIAL PRIMARY KEY,
  return_id BIGINT NOT NULL REFERENCES purchase_returns(id) ON DELETE CASCADE,
  item_id BIGINT NOT NULL REFERENCES items(id),
  quantity NUMERIC(14,3) NOT NULL,
  unit_cost NUMERIC(14,2) NOT NULL,
  total NUMERIC(14,2) NOT NULL
);

CREATE TABLE IF NOT EXISTS receipt_vouchers (
  id BIGSERIAL PRIMARY KEY,
  number VARCHAR(40) NOT NULL UNIQUE,
  branch_id BIGINT REFERENCES branches(id),
  customer_id BIGINT REFERENCES customers(id),
  amount NUMERIC(14,2) NOT NULL,
  description VARCHAR(255),
  user_id BIGINT NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS payment_vouchers (
  id BIGSERIAL PRIMARY KEY,
  number VARCHAR(40) NOT NULL UNIQUE,
  branch_id BIGINT REFERENCES branches(id),
  supplier_id BIGINT REFERENCES suppliers(id),
  amount NUMERIC(14,2) NOT NULL,
  description VARCHAR(255),
  user_id BIGINT NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS notifications (
  id BIGSERIAL PRIMARY KEY,
  branch_id BIGINT REFERENCES branches(id),
  type VARCHAR(40) NOT NULL,
  title VARCHAR(160) NOT NULL,
  message VARCHAR(500),
  is_read BOOLEAN NOT NULL DEFAULT FALSE,
  user_id BIGINT REFERENCES users(id),
  created_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS login_attempts (
  id BIGSERIAL PRIMARY KEY,
  username VARCHAR(80),
  ip VARCHAR(64),
  success BOOLEAN NOT NULL,
  created_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_login_attempts_username ON login_attempts(username);
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications(is_read, created_at);
