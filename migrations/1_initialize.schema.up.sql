

CREATE TABLE if not exists Users (
   id SERIAL PRIMARY KEY,
   subject        TEXT    NOT NULL UNIQUE,
   email          TEXT    NOT NULL UNIQUE,
   provider          TEXT    NOT NULL UNIQUE,
   last_login  timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE if not exists Races (
   id SERIAL PRIMARY KEY NOT NULL,
   user_id integer NOT NULL,
   race_name          VARCHAR(256)    NOT NULL,
   start_time timestamp without time zone,
   status VARCHAR(32) NOT NULL
);

CREATE TABLE if not exists Race_Registrations  (
   user_id integer NOT NULL,
   race_id int not null REFERENCES Races(id),
   registered_on timestamp without time zone,
   placed int not null
) 
