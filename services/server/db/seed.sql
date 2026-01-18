INSERT INTO interfaces (id, name, address, listen_port, mtu, endpoint)
VALUES
  ('wg0', 'Development WG0', '10.0.0.1/24', 51820, 1420, 'vpn.dev.local:51820'),
  ('wg1', 'Development WG1', '10.1.0.1/24', 51821, 1420, 'vpn.dev.local:51821')
ON CONFLICT (id) DO NOTHING;

INSERT INTO allowed_emails (interface_id, email)
VALUES
  ('wg0', 'dev1@example.com'),
  ('wg0', 'dev2@example.com'),
  ('wg1', 'dev2@example.com')
ON CONFLICT DO NOTHING;
