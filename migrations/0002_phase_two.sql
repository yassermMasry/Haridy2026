CREATE TABLE IF NOT EXISTS customers (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(160) NOT NULL,
  phone VARCHAR(40),
  address VARCHAR(255),
  notes VARCHAR(500),
  credit_limit NUMERIC(14,2) NOT NULL DEFAULT 0,
  balance NUMERIC(14,2) NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_customers_name ON customers(name);
CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone);
CREATE INDEX IF NOT EXISTS idx_customers_balance ON customers(balance);

CREATE TABLE IF NOT EXISTS suppliers (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(160) NOT NULL,
  phone VARCHAR(40),
  address VARCHAR(255),
  notes VARCHAR(500),
  balance NUMERIC(14,2) NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_suppliers_name ON suppliers(name);
CREATE INDEX IF NOT EXISTS idx_suppliers_phone ON suppliers(phone);
CREATE INDEX IF NOT EXISTS idx_suppliers_balance ON suppliers(balance);

CREATE TABLE IF NOT EXISTS customer_transactions (
  id BIGSERIAL PRIMARY KEY,
  customer_id BIGINT NOT NULL REFERENCES customers(id),
  type VARCHAR(20) NOT NULL,
  debit NUMERIC(14,2) NOT NULL DEFAULT 0,
  credit NUMERIC(14,2) NOT NULL DEFAULT 0,
  reference VARCHAR(120),
  description VARCHAR(255),
  user_id BIGINT REFERENCES users(id),
  created_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_customer_transactions_customer_id ON customer_transactions(customer_id);
CREATE INDEX IF NOT EXISTS idx_customer_transactions_created_at ON customer_transactions(created_at);

CREATE TABLE IF NOT EXISTS supplier_transactions (
  id BIGSERIAL PRIMARY KEY,
  supplier_id BIGINT NOT NULL REFERENCES suppliers(id),
  type VARCHAR(20) NOT NULL,
  debit NUMERIC(14,2) NOT NULL DEFAULT 0,
  credit NUMERIC(14,2) NOT NULL DEFAULT 0,
  reference VARCHAR(120),
  description VARCHAR(255),
  user_id BIGINT REFERENCES users(id),
  created_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_supplier_transactions_supplier_id ON supplier_transactions(supplier_id);
CREATE INDEX IF NOT EXISTS idx_supplier_transactions_created_at ON supplier_transactions(created_at);

ALTER TABLE sales_invoices ADD COLUMN IF NOT EXISTS customer_id BIGINT REFERENCES customers(id);
ALTER TABLE sales_invoices ADD COLUMN IF NOT EXISTS payment_type VARCHAR(20) NOT NULL DEFAULT 'cash';
CREATE INDEX IF NOT EXISTS idx_sales_invoices_customer_id ON sales_invoices(customer_id);
CREATE INDEX IF NOT EXISTS idx_sales_invoices_payment_type ON sales_invoices(payment_type);

CREATE TABLE IF NOT EXISTS purchase_invoices (
  id BIGSERIAL PRIMARY KEY,
  number VARCHAR(40) NOT NULL UNIQUE,
  supplier_id BIGINT NOT NULL REFERENCES suppliers(id),
  user_id BIGINT NOT NULL REFERENCES users(id),
  payment_type VARCHAR(20) NOT NULL DEFAULT 'cash',
  subtotal NUMERIC(14,2) NOT NULL DEFAULT 0,
  discount NUMERIC(14,2) NOT NULL DEFAULT 0,
  tax NUMERIC(14,2) NOT NULL DEFAULT 0,
  total NUMERIC(14,2) NOT NULL DEFAULT 0,
  paid_cash NUMERIC(14,2) NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_purchase_invoices_supplier_id ON purchase_invoices(supplier_id);
CREATE INDEX IF NOT EXISTS idx_purchase_invoices_created_at ON purchase_invoices(created_at);

CREATE TABLE IF NOT EXISTS purchase_invoice_items (
  id BIGSERIAL PRIMARY KEY,
  invoice_id BIGINT NOT NULL REFERENCES purchase_invoices(id) ON DELETE CASCADE,
  item_id BIGINT NOT NULL REFERENCES items(id),
  quantity NUMERIC(14,3) NOT NULL,
  unit_cost NUMERIC(14,2) NOT NULL,
  total NUMERIC(14,2) NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_purchase_invoice_items_invoice_id ON purchase_invoice_items(invoice_id);
CREATE INDEX IF NOT EXISTS idx_purchase_invoice_items_item_id ON purchase_invoice_items(item_id);

CREATE TABLE IF NOT EXISTS chart_of_accounts (
  id BIGSERIAL PRIMARY KEY,
  code VARCHAR(30) NOT NULL UNIQUE,
  name VARCHAR(120) NOT NULL,
  type VARCHAR(20) NOT NULL,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS journal_entries (
  id BIGSERIAL PRIMARY KEY,
  number VARCHAR(40) NOT NULL UNIQUE,
  reference VARCHAR(120),
  description VARCHAR(255),
  created_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS journal_entry_lines (
  id BIGSERIAL PRIMARY KEY,
  entry_id BIGINT NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
  account_id BIGINT NOT NULL REFERENCES chart_of_accounts(id),
  debit NUMERIC(14,2) NOT NULL DEFAULT 0,
  credit NUMERIC(14,2) NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS audit_logs (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT REFERENCES users(id),
  action VARCHAR(80) NOT NULL,
  entity VARCHAR(80) NOT NULL,
  entity_id BIGINT,
  details VARCHAR(1000),
  created_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
