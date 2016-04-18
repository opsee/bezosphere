SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

CREATE FUNCTION update_time() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
      BEGIN
      NEW.updated_at := CURRENT_TIMESTAMP;
      RETURN NEW;
      END;
      $$;


SET default_tablespace = '';
SET default_with_oids = false;

create table resources (
  id character varying(21) not null,
  customer_id UUID not null,
  request_type character varying(64) not null,
  request_data jsonb not null,
  response_type character varying(64) not null,
  response_data jsonb not null,
  created_at timestamp with time zone DEFAULT now() NOT NULL,
  updated_at timestamp with time zone DEFAULT now() NOT NULL,
  primary key (customer_id, id)
);

create trigger update_resources before update on resources for each row execute procedure update_time();
