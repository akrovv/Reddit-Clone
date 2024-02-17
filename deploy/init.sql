CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    username varchar(255) NOT NULL,
    password varchar(255) NOT NULL,
    is_active boolean NOT NULL
);