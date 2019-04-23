
CREATE TABLE profiles (
    user_id varchar(255) PRIMARY KEY,
    schema_version int NOT NULL,
    profile json NOT NULL
);
