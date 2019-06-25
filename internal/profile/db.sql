CREATE TABLE IF NOT EXISTS profiles (
    user_id varchar(255) PRIMARY KEY,
    profile json NOT NULL
);

ALTER TABLE profiles ADD COLUMN created_at  timestamp DEFAULT now();
ALTER TABLE profiles ADD COLUMN updated_at  timestamp DEFAULT now();
ALTER TABLE profiles ADD COLUMN version INTEGER DEFAULT 1;


CREATE OR REPLACE FUNCTION update_metadata()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  NEW.version := OLD.version + 1;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_version
  BEFORE UPDATE ON profiles
  FOR EACH ROW
EXECUTE PROCEDURE update_metadata();

commit