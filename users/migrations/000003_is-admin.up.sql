ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT FALSE;
UPDATE users SET is_admin = TRUE WHERE email = 'matlopes1999@gmail.com';
