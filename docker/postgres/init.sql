-- Create extensions if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create custom types if needed
CREATE TYPE post_status AS ENUM ('scheduled', 'posted', 'failed');
CREATE TYPE influencer_status AS ENUM ('active', 'inactive', 'suspended');

-- Set up any additional database configurations
ALTER DATABASE socialforge SET timezone TO 'UTC'; 