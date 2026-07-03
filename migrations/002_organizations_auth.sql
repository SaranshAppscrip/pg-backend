-- Multi-organization auth migration

CREATE TABLE organizations (
  id         UUID PRIMARY KEY,
  name       TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO organizations (id, name, created_at, updated_at)
SELECT id, pg_name, created_at, updated_at FROM business_settings;

DROP TABLE business_settings;

-- staff
ALTER TABLE staff ADD COLUMN organization_id UUID REFERENCES organizations(id);
UPDATE staff SET organization_id = (SELECT id FROM organizations ORDER BY created_at LIMIT 1);
ALTER TABLE staff ALTER COLUMN organization_id SET NOT NULL;
ALTER TABLE staff DROP CONSTRAINT IF EXISTS staff_email_key;
CREATE UNIQUE INDEX staff_org_email_unique ON staff (organization_id, email);

-- rooms
ALTER TABLE rooms ADD COLUMN organization_id UUID REFERENCES organizations(id);
UPDATE rooms SET organization_id = (SELECT id FROM organizations ORDER BY created_at LIMIT 1);
ALTER TABLE rooms ALTER COLUMN organization_id SET NOT NULL;
ALTER TABLE rooms DROP CONSTRAINT IF EXISTS rooms_room_number_key;
CREATE UNIQUE INDEX rooms_org_room_number_unique ON rooms (organization_id, room_number);

-- tenants
ALTER TABLE tenants ADD COLUMN organization_id UUID REFERENCES organizations(id);
UPDATE tenants SET organization_id = (SELECT id FROM organizations ORDER BY created_at LIMIT 1);
ALTER TABLE tenants ALTER COLUMN organization_id SET NOT NULL;

ALTER TABLE tenants ADD COLUMN email TEXT;
ALTER TABLE tenants ADD COLUMN password_hash TEXT;
ALTER TABLE tenants ALTER COLUMN phone DROP NOT NULL;
ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_phone_key;
CREATE UNIQUE INDEX tenants_org_email_unique ON tenants (organization_id, email) WHERE email IS NOT NULL;

-- expenses
ALTER TABLE expenses ADD COLUMN organization_id UUID REFERENCES organizations(id);
UPDATE expenses SET organization_id = (SELECT id FROM organizations ORDER BY created_at LIMIT 1);
ALTER TABLE expenses ALTER COLUMN organization_id SET NOT NULL;

-- kitchen_items
ALTER TABLE kitchen_items ADD COLUMN organization_id UUID REFERENCES organizations(id);
UPDATE kitchen_items SET organization_id = (SELECT id FROM organizations ORDER BY created_at LIMIT 1);
ALTER TABLE kitchen_items ALTER COLUMN organization_id SET NOT NULL;

-- payments
ALTER TABLE payments ADD COLUMN organization_id UUID REFERENCES organizations(id);
UPDATE payments SET organization_id = (
  SELECT t.organization_id FROM tenants t WHERE t.id = payments.tenant_id
);
ALTER TABLE payments ALTER COLUMN organization_id SET NOT NULL;

-- kitchen_log
ALTER TABLE kitchen_log ADD COLUMN organization_id UUID REFERENCES organizations(id);
UPDATE kitchen_log SET organization_id = (
  SELECT ki.organization_id FROM kitchen_items ki WHERE ki.id = kitchen_log.item_id
);
UPDATE kitchen_log SET organization_id = (SELECT id FROM organizations ORDER BY created_at LIMIT 1)
WHERE organization_id IS NULL;
ALTER TABLE kitchen_log ALTER COLUMN organization_id SET NOT NULL;

-- Seed owner staff for invite-only onboarding (password: admin123)
INSERT INTO staff (id, organization_id, email, password_hash, full_name, is_owner)
SELECT
  '00000000-0000-0000-0000-000000000002',
  (SELECT id FROM organizations ORDER BY created_at LIMIT 1),
  'owner@nivas.local',
  '$2b$12$RXZR8g39XZFhg2CZmW0ml.99x74mYj43iQdd6Aytd/TsyX4aPK5Qy',
  'Owner',
  true
WHERE NOT EXISTS (SELECT 1 FROM staff);
