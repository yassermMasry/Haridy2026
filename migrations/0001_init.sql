CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  username VARCHAR(80) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  role VARCHAR(20) NOT NULL,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

CREATE TABLE IF NOT EXISTS item_categories (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(120) NOT NULL UNIQUE,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS items (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(160) NOT NULL,
  code VARCHAR(80) NOT NULL UNIQUE,
  barcode VARCHAR(120),
  purchase_price NUMERIC(14,2) NOT NULL DEFAULT 0,
  sale_price NUMERIC(14,2) NOT NULL DEFAULT 0,
  quantity NUMERIC(14,3) NOT NULL DEFAULT 0,
  minimum_stock NUMERIC(14,3) NOT NULL DEFAULT 0,
  category_id BIGINT REFERENCES item_categories(id),
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_items_name ON items(name);
CREATE INDEX IF NOT EXISTS idx_items_barcode ON items(barcode);
CREATE INDEX IF NOT EXISTS idx_items_quantity ON items(quantity);
CREATE INDEX IF NOT EXISTS idx_items_category_id ON items(category_id);
CREATE INDEX IF NOT EXISTS idx_items_deleted_at ON items(deleted_at);

CREATE TABLE IF NOT EXISTS stock_movements (
  id BIGSERIAL PRIMARY KEY,
  item_id BIGINT NOT NULL REFERENCES items(id),
  type VARCHAR(20) NOT NULL,
  quantity NUMERIC(14,3) NOT NULL,
  reference VARCHAR(120),
  notes VARCHAR(255),
  performed_by BIGINT REFERENCES users(id),
  created_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_stock_movements_item_id ON stock_movements(item_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_type ON stock_movements(type);
CREATE INDEX IF NOT EXISTS idx_stock_movements_reference ON stock_movements(reference);
CREATE INDEX IF NOT EXISTS idx_stock_movements_created_at ON stock_movements(created_at);

CREATE TABLE IF NOT EXISTS sales_invoices (
  id BIGSERIAL PRIMARY KEY,
  number VARCHAR(40) NOT NULL UNIQUE,
  user_id BIGINT NOT NULL REFERENCES users(id),
  subtotal NUMERIC(14,2) NOT NULL DEFAULT 0,
  discount NUMERIC(14,2) NOT NULL DEFAULT 0,
  tax NUMERIC(14,2) NOT NULL DEFAULT 0,
  total NUMERIC(14,2) NOT NULL DEFAULT 0,
  paid_cash NUMERIC(14,2) NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_sales_invoices_user_id ON sales_invoices(user_id);
CREATE INDEX IF NOT EXISTS idx_sales_invoices_created_at ON sales_invoices(created_at);

CREATE TABLE IF NOT EXISTS sales_invoice_items (
  id BIGSERIAL PRIMARY KEY,
  invoice_id BIGINT NOT NULL REFERENCES sales_invoices(id) ON DELETE CASCADE,
  item_id BIGINT NOT NULL REFERENCES items(id),
  quantity NUMERIC(14,3) NOT NULL,
  unit_price NUMERIC(14,2) NOT NULL,
  total NUMERIC(14,2) NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sales_invoice_items_invoice_id ON sales_invoice_items(invoice_id);
CREATE INDEX IF NOT EXISTS idx_sales_invoice_items_item_id ON sales_invoice_items(item_id);

CREATE TABLE IF NOT EXISTS treasuries (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(120) NOT NULL UNIQUE,
  balance NUMERIC(14,2) NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS treasury_transactions (
  id BIGSERIAL PRIMARY KEY,
  treasury_id BIGINT NOT NULL REFERENCES treasuries(id),
  type VARCHAR(20) NOT NULL,
  amount NUMERIC(14,2) NOT NULL,
  reference VARCHAR(120),
  description VARCHAR(255),
  user_id BIGINT REFERENCES users(id),
  created_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_treasury_transactions_treasury_id ON treasury_transactions(treasury_id);
CREATE INDEX IF NOT EXISTS idx_treasury_transactions_type ON treasury_transactions(type);
CREATE INDEX IF NOT EXISTS idx_treasury_transactions_reference ON treasury_transactions(reference);
CREATE INDEX IF NOT EXISTS idx_treasury_transactions_created_at ON treasury_transactions(created_at);
