CREATE TABLE IF NOT EXISTS profiles (
    user_id varchar(255) PRIMARY KEY,
    profile json NOT NULL
);

ALTER TABLE profiles ADD COLUMN IF NOT EXISTS created_at  timestamp DEFAULT now();
ALTER TABLE profiles ADD COLUMN IF NOT EXISTS updated_at  timestamp DEFAULT now();
ALTER TABLE profiles ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1;


CREATE OR REPLACE FUNCTION update_metadata()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  NEW.version := OLD.version + 1;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS set_metadata ON profiles;
CREATE TRIGGER set_metadata
  BEFORE UPDATE ON profiles
  FOR EACH ROW
EXECUTE PROCEDURE update_metadata();
