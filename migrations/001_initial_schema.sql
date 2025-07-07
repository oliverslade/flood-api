--
-- Migration 001: Initial schema
-- PostgreSQL database dump
--

-- Dumped from database version 15.13 (Postgres.app)
-- Dumped by pg_dump version 17.0

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: rainfalls; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.rainfalls (
    stationid text,
    "timestamp" text,
    level real
);


--
-- Name: riverlevels; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.riverlevels (
    "timestamp" text,
    level real
);


--
-- Name: stationnames; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.stationnames (
    id text,
    name text
);


--
-- PostgreSQL database dump complete
--
