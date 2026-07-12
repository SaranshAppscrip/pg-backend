-- Nivas PG Management schema (multi-organization)

CREATE TYPE payment_mode AS ENUM ('Cash', 'UPI', 'Bank Transfer', 'Other');
CREATE TYPE expense_category AS ENUM (
  'Kitchen Supplies', 'Maintenance', 'Electricity', 'Water',
  'Staff Salary', 'Rent', 'Other'
);
CREATE TYPE kitchen_unit AS ENUM ('kg', 'litre', 'packet', 'piece', 'dozen');
CREATE TYPE kitchen_log_type AS ENUM ('in', 'out');

CREATE TABLE organizations (
  id         UUID PRIMARY KEY,
  name       TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO organizations (id, name) VALUES ('00000000-0000-0000-0000-000000000001', 'Nivas PG');

CREATE TABLE staff (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id),
  email           TEXT NOT NULL,
  password_hash   TEXT NOT NULL,
  full_name       TEXT,
  is_owner        BOOLEAN NOT NULL DEFAULT false,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT staff_org_email_unique UNIQUE (organization_id, email)
);

CREATE TABLE rooms (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id),
  room_number     TEXT NOT NULL,
  capacity        INT NOT NULL CHECK (capacity > 0),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT rooms_org_room_number_unique UNIQUE (organization_id, room_number)
);

CREATE TABLE tenants (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id),
  name            TEXT NOT NULL,
  email           TEXT NOT NULL,
  password_hash   TEXT NOT NULL,
  phone           TEXT,
  room_id         UUID REFERENCES rooms(id) ON DELETE SET NULL,
  monthly_fee     NUMERIC(12, 2) NOT NULL CHECK (monthly_fee >= 0),
  join_date       DATE NOT NULL DEFAULT CURRENT_DATE,
  active          BOOLEAN NOT NULL DEFAULT true,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT tenants_org_email_unique UNIQUE (organization_id, email)
);

CREATE UNIQUE INDEX tenants_org_phone_unique ON tenants (organization_id, phone) WHERE phone IS NOT NULL;
CREATE UNIQUE INDEX tenants_org_name_unique ON tenants (organization_id, lower(trim(name))) WHERE active = true;

CREATE TABLE payments (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id),
  tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  amount          NUMERIC(12, 2) NOT NULL CHECK (amount > 0),
  date            DATE NOT NULL DEFAULT CURRENT_DATE,
  for_month       TEXT NOT NULL,
  mode            payment_mode NOT NULL DEFAULT 'Cash',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE expenses (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id),
  category        expense_category NOT NULL,
  amount          NUMERIC(12, 2) NOT NULL CHECK (amount > 0),
  date            DATE NOT NULL DEFAULT CURRENT_DATE,
  note            TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE kitchen_items (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id   UUID NOT NULL REFERENCES organizations(id),
  name              TEXT NOT NULL,
  qty               NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (qty >= 0),
  unit              kitchen_unit NOT NULL DEFAULT 'kg',
  reorder_threshold NUMERIC(12, 2) NOT NULL DEFAULT 0,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX kitchen_items_org_name_unique ON kitchen_items (organization_id, lower(trim(name)));

CREATE TABLE kitchen_log (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id),
  item_id         UUID NOT NULL REFERENCES kitchen_items(id) ON DELETE CASCADE,
  type            kitchen_log_type NOT NULL,
  qty             NUMERIC(12, 2) NOT NULL CHECK (qty > 0),
  date            DATE NOT NULL DEFAULT CURRENT_DATE,
  note            TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenants_room ON tenants(room_id) WHERE active = true;
CREATE INDEX idx_tenants_org_email ON tenants(organization_id, email);
CREATE INDEX idx_payments_tenant ON payments(tenant_id);
CREATE INDEX idx_payments_org ON payments(organization_id);
CREATE INDEX idx_payments_for_month ON payments(for_month);
CREATE INDEX idx_kitchen_log_item ON kitchen_log(item_id);
CREATE INDEX idx_kitchen_log_org ON kitchen_log(organization_id);

CREATE TABLE password_reset_tokens (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  staff_id   UUID NOT NULL REFERENCES staff(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at    TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX password_reset_tokens_staff_id_idx ON password_reset_tokens (staff_id);
CREATE INDEX password_reset_tokens_token_hash_idx ON password_reset_tokens (token_hash);

CREATE TABLE tenant_password_reset_tokens (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id  UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at    TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX tenant_password_reset_tokens_tenant_id_idx ON tenant_password_reset_tokens (tenant_id);
CREATE INDEX tenant_password_reset_tokens_token_hash_idx ON tenant_password_reset_tokens (token_hash);

-- Refresh tokens (from access/refresh token auth)
CREATE TABLE IF NOT EXISTS refresh_tokens (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_type       TEXT NOT NULL CHECK (user_type IN ('staff', 'tenant')),
  user_id         UUID NOT NULL,
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  token_hash      TEXT NOT NULL,
  expires_at      TIMESTAMPTZ NOT NULL,
  revoked_at      TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS refresh_tokens_token_hash_idx ON refresh_tokens (token_hash);
CREATE INDEX IF NOT EXISTS refresh_tokens_user_idx ON refresh_tokens (user_type, user_id);
CREATE INDEX IF NOT EXISTS refresh_tokens_expires_at_idx ON refresh_tokens (expires_at) WHERE revoked_at IS NULL;

-- Soft deletes for financial records
ALTER TABLE payments ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_payments_active ON payments (organization_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_expenses_active ON expenses (organization_id) WHERE deleted_at IS NULL;

-- Staff audit trail
CREATE TYPE audit_entity_type AS ENUM ('payment', 'expense', 'tenant');
CREATE TYPE audit_action AS ENUM ('create', 'delete', 'move_out');

CREATE TABLE staff_audit_log (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  staff_id        UUID REFERENCES staff(id) ON DELETE SET NULL,
  entity_type     audit_entity_type NOT NULL,
  entity_id       UUID NOT NULL,
  action          audit_action NOT NULL,
  metadata        JSONB NOT NULL DEFAULT '{}',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_staff_audit_log_org ON staff_audit_log (organization_id, created_at DESC);
CREATE INDEX idx_staff_audit_log_entity ON staff_audit_log (entity_type, entity_id);

-- Multi-property support
CREATE TABLE properties (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  name            TEXT NOT NULL,
  address         TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT properties_org_name_unique UNIQUE (organization_id, name)
);

-- Default property per org for existing data
INSERT INTO properties (id, organization_id, name)
SELECT gen_random_uuid(), o.id, o.name
FROM organizations o
WHERE NOT EXISTS (SELECT 1 FROM properties p WHERE p.organization_id = o.id);

ALTER TABLE rooms ADD COLUMN IF NOT EXISTS property_id UUID REFERENCES properties(id);

UPDATE rooms r
SET property_id = p.id
FROM properties p
WHERE r.property_id IS NULL AND p.organization_id = r.organization_id;

ALTER TABLE rooms ALTER COLUMN property_id SET NOT NULL;

ALTER TABLE rooms DROP CONSTRAINT IF EXISTS rooms_org_room_number_unique;
ALTER TABLE rooms ADD CONSTRAINT rooms_property_room_number_unique UNIQUE (property_id, room_number);
CREATE INDEX IF NOT EXISTS idx_rooms_property ON rooms(property_id);

-- Rent reminder deduplication
CREATE TABLE rent_reminder_log (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id     UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  for_month     TEXT NOT NULL,
  reminder_type TEXT NOT NULL CHECK (reminder_type IN ('due', 'overdue')),
  sent_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT rent_reminder_unique UNIQUE (tenant_id, for_month, reminder_type)
);

CREATE INDEX idx_rent_reminder_tenant ON rent_reminder_log(tenant_id);

-- Document storage (tenant ID proof, lease; owner PG registration / permissions)
CREATE TYPE tenant_document_type AS ENUM (
  'id_proof', 'lease_agreement', 'police_verification', 'photo', 'other'
);

CREATE TYPE organization_document_type AS ENUM (
  'pg_registration', 'fire_safety_noc', 'police_permission',
  'trade_license', 'property_tax', 'other'
);

CREATE TABLE tenant_documents (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id   UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  tenant_id         UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  document_type     tenant_document_type NOT NULL,
  title             TEXT,
  storage_key       TEXT NOT NULL,
  original_filename TEXT NOT NULL,
  content_type      TEXT NOT NULL,
  size_bytes        BIGINT NOT NULL CHECK (size_bytes > 0),
  uploaded_by       UUID REFERENCES staff(id) ON DELETE SET NULL,
  expires_at        DATE,
  deleted_at        TIMESTAMPTZ,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE organization_documents (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id   UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  property_id       UUID REFERENCES properties(id) ON DELETE SET NULL,
  document_type     organization_document_type NOT NULL,
  title             TEXT,
  storage_key       TEXT NOT NULL,
  original_filename TEXT NOT NULL,
  content_type      TEXT NOT NULL,
  size_bytes        BIGINT NOT NULL CHECK (size_bytes > 0),
  uploaded_by       UUID REFERENCES staff(id) ON DELETE SET NULL,
  expires_at        DATE,
  deleted_at        TIMESTAMPTZ,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenant_documents_tenant ON tenant_documents(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenant_documents_org ON tenant_documents(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_organization_documents_org ON organization_documents(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_organization_documents_property ON organization_documents(property_id) WHERE deleted_at IS NULL;

-- Announcements / notice board
CREATE TYPE announcement_category AS ENUM ('maintenance', 'holiday', 'rules', 'general');

CREATE TABLE announcements (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  property_id     UUID REFERENCES properties(id) ON DELETE SET NULL,
  title           TEXT NOT NULL,
  body            TEXT NOT NULL,
  category        announcement_category NOT NULL DEFAULT 'general',
  pinned          BOOLEAN NOT NULL DEFAULT false,
  published       BOOLEAN NOT NULL DEFAULT true,
  expires_at      TIMESTAMPTZ,
  created_by      UUID REFERENCES staff(id) ON DELETE SET NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_announcements_org ON announcements(organization_id, published, created_at DESC);

-- Maintenance / complaint requests
CREATE TYPE maintenance_category AS ENUM ('electrical', 'plumbing', 'wifi', 'cleaning', 'other');
CREATE TYPE maintenance_status AS ENUM ('open', 'in_progress', 'resolved', 'closed');
CREATE TYPE maintenance_priority AS ENUM ('low', 'medium', 'high', 'urgent');

CREATE TABLE maintenance_requests (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  category        maintenance_category NOT NULL,
  title           TEXT NOT NULL,
  description     TEXT NOT NULL,
  status          maintenance_status NOT NULL DEFAULT 'open',
  priority        maintenance_priority NOT NULL DEFAULT 'medium',
  assigned_to     UUID REFERENCES staff(id) ON DELETE SET NULL,
  staff_note      TEXT,
  resolved_at     TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_maintenance_org_status ON maintenance_requests(organization_id, status, created_at DESC);
CREATE INDEX idx_maintenance_tenant ON maintenance_requests(tenant_id);
CREATE INDEX idx_maintenance_assigned ON maintenance_requests(assigned_to) WHERE assigned_to IS NOT NULL;

-- Visitor log
CREATE TABLE visitor_log (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  property_id     UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
  tenant_id       UUID REFERENCES tenants(id) ON DELETE SET NULL,
  visitor_name    TEXT NOT NULL,
  visitor_phone   TEXT,
  purpose             TEXT,
  id_type             TEXT,
  id_number_encrypted TEXT,
  id_number_last4     TEXT,
  entry_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  exit_at         TIMESTAMPTZ,
  logged_by       UUID REFERENCES staff(id) ON DELETE SET NULL,
  notes           TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_visitor_log_org ON visitor_log(organization_id, entry_at DESC);
CREATE INDEX idx_visitor_log_property ON visitor_log(property_id, entry_at DESC);

