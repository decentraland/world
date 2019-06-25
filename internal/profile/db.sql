CREATE TABLE IF NOT EXISTS profiles (
    user_id varchar(255) PRIMARY KEY,
    profile json NOT NULL
);

ALTER TABLE profiles ADD COLUMN version BIGINT default(extract(epoch from now()) * 1000);

CREATE OR REPLACE FUNCTION update_version()
RETURNS TRIGGER AS $$
BEGIN
  NEW.version = extract(epoch FROM NOW()) * 1000;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_version
  BEFORE UPDATE ON profiles
  FOR EACH ROW
EXECUTE PROCEDURE update_version();

commit