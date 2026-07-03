-- Nivas PG Management schema

CREATE TYPE payment_mode AS ENUM ('Cash', 'UPI', 'Bank Transfer', 'Other');
CREATE TYPE expense_category AS ENUM (
  'Kitchen Supplies', 'Maintenance', 'Electricity', 'Water',
  'Staff Salary', 'Rent', 'Other'
);
CREATE TYPE kitchen_unit AS ENUM ('kg', 'litre', 'packet', 'piece', 'dozen');
CREATE TYPE kitchen_log_type AS ENUM ('in', 'out');

CREATE TABLE business_settings (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  pg_name TEXT NOT NULL DEFAULT 'Nivas PG',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO business_settings (id, pg_name) VALUES ('00000000-0000-0000-0000-000000000001', 'Nivas PG');

CREATE TABLE staff (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  full_name TEXT,
  is_owner BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE rooms (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_number TEXT NOT NULL UNIQUE,
  capacity INT NOT NULL CHECK (capacity > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE tenants (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  phone TEXT NOT NULL UNIQUE,
  room_id UUID REFERENCES rooms(id) ON DELETE SET NULL,
  monthly_fee NUMERIC(12, 2) NOT NULL CHECK (monthly_fee >= 0),
  join_date DATE NOT NULL DEFAULT CURRENT_DATE,
  active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE payments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  amount NUMERIC(12, 2) NOT NULL CHECK (amount > 0),
  date DATE NOT NULL DEFAULT CURRENT_DATE,
  for_month TEXT NOT NULL,
  mode payment_mode NOT NULL DEFAULT 'Cash',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE expenses (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  category expense_category NOT NULL,
  amount NUMERIC(12, 2) NOT NULL CHECK (amount > 0),
  date DATE NOT NULL DEFAULT CURRENT_DATE,
  note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE kitchen_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  qty NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (qty >= 0),
  unit kitchen_unit NOT NULL DEFAULT 'kg',
  reorder_threshold NUMERIC(12, 2) NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE kitchen_log (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  item_id UUID NOT NULL REFERENCES kitchen_items(id) ON DELETE CASCADE,
  type kitchen_log_type NOT NULL,
  qty NUMERIC(12, 2) NOT NULL CHECK (qty > 0),
  date DATE NOT NULL DEFAULT CURRENT_DATE,
  note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenants_room ON tenants(room_id) WHERE active = true;
CREATE INDEX idx_tenants_phone ON tenants(phone);
CREATE INDEX idx_payments_tenant ON payments(tenant_id);
CREATE INDEX idx_payments_for_month ON payments(for_month);
CREATE INDEX idx_kitchen_log_item ON kitchen_log(item_id);
