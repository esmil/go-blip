/* In database "blip or something such" */

BEGIN;

CREATE TABLE blip (
       tstamp	       TIMESTAMP(6)
);

CREATE USER blip WITH PASSWORD '%%%PASSWORD%%%';
GRANT INSERT ON blip TO blip;


