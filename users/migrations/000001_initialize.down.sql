-- Drop the index
DROP INDEX IF EXISTS idx_users_email;

-- Drop the users table
DROP TABLE IF EXISTS users;

-- Drop the uuid-ossp extension
DROP EXTENSION IF EXISTS "uuid-ossp";
