
drop table race_registrations;
drop table races;
drop table users;
CREATE TABLE Users (
   id SERIAL PRIMARY KEY NOT NULL,
   email          TEXT    NOT NULL UNIQUE,
   provider  VARCHAR(32),
   password_digest BYTEA,
   last_login  timestamp without time zone
);

CREATE TABLE Races (
   id SERIAL PRIMARY KEY NOT NULL,
   name          VARCHAR(256)    NOT NULL,
   start_time timestamp without time zone,
   status VARCHAR(32) NOT NULL
);

CREATE TABLE Race_Registrations  (
   id SERIAL PRIMARY KEY NOT NULL,
   user_id int not null REFERENCES Users(id),
   race_id int not null REFERENCES Races(id),
   placed int not null
) 
