-- AxisRelay initial RBAC permission catalog. Idempotent by permission code.
INSERT INTO permissions (code, name, description) VALUES
  ('dashboard.read', 'View dashboard', 'Read commercial operations dashboard'),
  ('users.read', 'View users', 'Read user and profile data'),
  ('users.write', 'Manage users', 'Create, update, suspend and restore users'),
  ('wallet.read', 'View wallets', 'Read balances and wallet ledger'),
  ('wallet.write', 'Manage wallets', 'Perform authorized wallet adjustments'),
  ('orders.read', 'View orders', 'Read orders and payments'),
  ('orders.write', 'Manage orders', 'Operate orders and payment reconciliation'),
  ('plans.read', 'View plans', 'Read plans and subscriptions'),
  ('plans.write', 'Manage plans', 'Create and update commercial plans'),
  ('gateway.read', 'View gateway', 'Read providers, models and gateway state'),
  ('gateway.write', 'Manage gateway', 'Manage providers, models and gateway configuration'),
  ('apikeys.read', 'View API keys', 'Read user API key metadata'),
  ('apikeys.write', 'Manage API keys', 'Create, rotate, disable and revoke API keys'),
  ('usage.read', 'View usage', 'Read usage and billing records'),
  ('risk.read', 'View risk', 'Read audit and risk records'),
  ('risk.write', 'Manage risk', 'Manage risk policies and review workflow'),
  ('system.read', 'View system settings', 'Read system settings'),
  ('system.write', 'Manage system settings', 'Update system settings')
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  description = VALUES(description);

INSERT INTO roles (code, name, description, built_in) VALUES
  ('super_admin', 'Super Admin', 'Built-in full-access administrator role', 1),
  ('operator', 'Operator', 'Built-in operations role', 1),
  ('auditor', 'Auditor', 'Built-in read-only audit role', 1)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  description = VALUES(description),
  built_in = VALUES(built_in);

INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p WHERE r.code = 'super_admin';

INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p ON p.code IN (
  'dashboard.read','users.read','wallet.read','orders.read','orders.write',
  'plans.read','gateway.read','gateway.write','apikeys.read','apikeys.write',
  'usage.read','risk.read','system.read'
) WHERE r.code = 'operator';

INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p ON p.code IN (
  'dashboard.read','users.read','wallet.read','orders.read','plans.read',
  'gateway.read','apikeys.read','usage.read','risk.read','system.read'
) WHERE r.code = 'auditor';
