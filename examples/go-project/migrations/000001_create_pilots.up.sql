CREATE TABLE pilots (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    callsgn TEXT UNIQUE
);
