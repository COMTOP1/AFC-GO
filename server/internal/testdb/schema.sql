CREATE TABLE teams (
    id           SERIAL PRIMARY KEY,
    name         TEXT    NOT NULL,
    description  TEXT,
    league       TEXT,
    division     TEXT,
    league_table TEXT,
    fixtures     TEXT,
    coach        TEXT,
    physio       TEXT,
    file_name    TEXT,
    active       BOOLEAN NOT NULL DEFAULT FALSE,
    youth        BOOLEAN NOT NULL DEFAULT FALSE,
    ages         INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE users (
    id             SERIAL PRIMARY KEY,
    name           TEXT    NOT NULL,
    email          TEXT    NOT NULL UNIQUE,
    phone          TEXT,
    team_id        INTEGER NOT NULL DEFAULT 0,
    role           TEXT    NOT NULL,
    file_name      TEXT,
    reset_password BOOLEAN NOT NULL DEFAULT FALSE,
    password       TEXT,
    hash           TEXT,
    salt           TEXT
);

CREATE TABLE players (
    id            SERIAL PRIMARY KEY,
    name          TEXT    NOT NULL,
    file_name     TEXT,
    date_of_birth DATE,
    position      TEXT,
    captain       BOOLEAN NOT NULL DEFAULT FALSE,
    team_id       INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE sponsors (
    id        SERIAL PRIMARY KEY,
    name      TEXT NOT NULL,
    website   TEXT,
    file_name TEXT,
    purpose   TEXT,
    team_id   TEXT NOT NULL DEFAULT ''
);

CREATE TABLE affiliations (
    id        SERIAL PRIMARY KEY,
    name      TEXT NOT NULL,
    website   TEXT,
    file_name TEXT
);

CREATE TABLE documents (
    id        SERIAL PRIMARY KEY,
    name      TEXT NOT NULL,
    file_name TEXT NOT NULL
);

CREATE TABLE images (
    id        SERIAL PRIMARY KEY,
    file_name TEXT NOT NULL,
    caption   TEXT
);

CREATE TABLE news (
    id        SERIAL PRIMARY KEY,
    title     TEXT        NOT NULL,
    file_name TEXT,
    content   TEXT,
    date      TIMESTAMPTZ NOT NULL
);

CREATE TABLE whatson (
    id            SERIAL PRIMARY KEY,
    title         TEXT        NOT NULL,
    file_name     TEXT,
    content       TEXT,
    date          TIMESTAMPTZ NOT NULL,
    date_of_event DATE        NOT NULL
);

CREATE TABLE programme_seasons (
    id     SERIAL PRIMARY KEY,
    season TEXT NOT NULL
);

CREATE TABLE programmes (
    id                  SERIAL PRIMARY KEY,
    name                TEXT    NOT NULL,
    file_name           TEXT    NOT NULL,
    date_of_programme   DATE    NOT NULL,
    programme_season_id INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE settings (
    id           TEXT PRIMARY KEY,
    setting_text TEXT NOT NULL
);
