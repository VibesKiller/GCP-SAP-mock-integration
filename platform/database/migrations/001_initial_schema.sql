CREATE TABLE IF NOT EXISTS customers (
  customer_id TEXT PRIMARY KEY,
  customer_number TEXT NOT NULL,
  full_name TEXT NOT NULL,
  email TEXT NOT NULL,
  phone TEXT,
  country_code TEXT NOT NULL,
  city TEXT,
  postal_code TEXT,
  segment TEXT,
  status TEXT NOT NULL,
  last_event_id TEXT NOT NULL,
  last_correlation_id TEXT NOT NULL,
  source_updated_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS orders (
  order_id TEXT PRIMARY KEY,
  customer_id TEXT NOT NULL REFERENCES customers(customer_id) ON DELETE RESTRICT,
  sales_org TEXT NOT NULL,
  distribution_channel TEXT NOT NULL,
  division TEXT NOT NULL,
  currency TEXT NOT NULL,
  status TEXT NOT NULL,
  requested_delivery_date DATE NOT NULL,
  document_date DATE NOT NULL,
  net_amount NUMERIC(18,2) NOT NULL,
  tax_amount NUMERIC(18,2) NOT NULL,
  total_amount NUMERIC(18,2) NOT NULL,
  last_event_id TEXT NOT NULL,
  last_correlation_id TEXT NOT NULL,
  source_updated_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
  order_id TEXT NOT NULL REFERENCES orders(order_id) ON DELETE CASCADE,
  line_number INTEGER NOT NULL,
  sku TEXT NOT NULL,
  description TEXT NOT NULL,
  quantity NUMERIC(18,3) NOT NULL,
  unit TEXT NOT NULL,
  unit_price NUMERIC(18,2) NOT NULL,
  net_amount NUMERIC(18,2) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (order_id, line_number)
);

CREATE TABLE IF NOT EXISTS invoices (
  invoice_id TEXT PRIMARY KEY,
  order_id TEXT NOT NULL REFERENCES orders(order_id) ON DELETE RESTRICT,
  customer_id TEXT NOT NULL REFERENCES customers(customer_id) ON DELETE RESTRICT,
  currency TEXT NOT NULL,
  status TEXT NOT NULL,
  issue_date TIMESTAMPTZ NOT NULL,
  due_date TIMESTAMPTZ NOT NULL,
  net_amount NUMERIC(18,2) NOT NULL,
  tax_amount NUMERIC(18,2) NOT NULL,
  total_amount NUMERIC(18,2) NOT NULL,
  last_event_id TEXT NOT NULL,
  last_correlation_id TEXT NOT NULL,
  source_updated_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS processed_events (
  event_id TEXT PRIMARY KEY,
  event_type TEXT NOT NULL,
  version TEXT NOT NULL,
  source TEXT NOT NULL,
  occurred_at TIMESTAMPTZ NOT NULL,
  correlation_id TEXT NOT NULL,
  kafka_topic TEXT NOT NULL,
  kafka_partition INTEGER NOT NULL,
  kafka_offset BIGINT NOT NULL,
  kafka_key TEXT,
  kafka_headers JSONB NOT NULL DEFAULT '{}'::jsonb,
  payload JSONB NOT NULL,
  status TEXT NOT NULL,
  processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_invoices_customer_id ON invoices(customer_id);
CREATE INDEX IF NOT EXISTS idx_invoices_order_id ON invoices(order_id);
CREATE INDEX IF NOT EXISTS idx_processed_events_type_occurred_at ON processed_events(event_type, occurred_at DESC);
