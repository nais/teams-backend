--
-- PostgreSQL database dump
--

-- Dumped from database version 14.3
-- Dumped by pg_dump version 14.4

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
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
-- Name: api_keys; Type: TABLE; Schema: public; Owner: console
--

CREATE TABLE public.api_keys (
    id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by_id uuid,
    updated_by_id uuid,
    updated_at timestamp with time zone,
    api_key text NOT NULL,
    user_id uuid NOT NULL
);


ALTER TABLE public.api_keys OWNER TO console;

--
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: console
--

CREATE TABLE public.audit_logs (
    id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by_id uuid,
    updated_by_id uuid,
    updated_at timestamp with time zone,
    actor_id uuid,
    correlation_id uuid NOT NULL,
    target_system_id uuid NOT NULL,
    target_team_id uuid,
    target_user_id uuid,
    action text NOT NULL,
    message text NOT NULL
);


ALTER TABLE public.audit_logs OWNER TO console;

--
-- Name: authorizations; Type: TABLE; Schema: public; Owner: console
--

CREATE TABLE public.authorizations (
    id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by_id uuid,
    updated_by_id uuid,
    updated_at timestamp with time zone,
    name text NOT NULL
);


ALTER TABLE public.authorizations OWNER TO console;

--
-- Name: correlations; Type: TABLE; Schema: public; Owner: console
--

CREATE TABLE public.correlations (
    id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by_id uuid,
    updated_by_id uuid,
    updated_at timestamp with time zone
);


ALTER TABLE public.correlations OWNER TO console;

--
-- Name: reconcile_errors; Type: TABLE; Schema: public; Owner: console
--

CREATE TABLE public.reconcile_errors (
    id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by_id uuid,
    updated_by_id uuid,
    updated_at timestamp with time zone,
    correlation_id uuid NOT NULL,
    system_id uuid NOT NULL,
    team_id uuid NOT NULL,
    message text NOT NULL
);


ALTER TABLE public.reconcile_errors OWNER TO console;

--
-- Name: role_authorizations; Type: TABLE; Schema: public; Owner: console
--

CREATE TABLE public.role_authorizations (
    authorization_id uuid NOT NULL,
    role_id uuid NOT NULL
);


ALTER TABLE public.role_authorizations OWNER TO console;

--
-- Name: roles; Type: TABLE; Schema: public; Owner: console
--

CREATE TABLE public.roles (
    id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by_id uuid,
    updated_by_id uuid,
    updated_at timestamp with time zone,
    name text NOT NULL
);


ALTER TABLE public.roles OWNER TO console;

--
-- Name: system_states; Type: TABLE; Schema: public; Owner: console
--

CREATE TABLE public.system_states (
    id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by_id uuid,
    updated_by_id uuid,
    updated_at timestamp with time zone,
    system_id uuid NOT NULL,
    team_id uuid NOT NULL,
    state jsonb DEFAULT '{}'::jsonb NOT NULL
);


ALTER TABLE public.system_states OWNER TO console;

--
-- Name: systems; Type: TABLE; Schema: public; Owner: console
--

CREATE TABLE public.systems (
    id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by_id uuid,
    updated_by_id uuid,
    updated_at timestamp with time zone,
    name text NOT NULL
);


ALTER TABLE public.systems OWNER TO console;

--
-- Name: team_metadata; Type: TABLE; Schema: public; Owner: console
--

CREATE TABLE public.team_metadata (
    id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by_id uuid,
    updated_by_id uuid,
    updated_at timestamp with time zone,
    team_id uuid NOT NULL,
    key text NOT NULL,
    value text
);


ALTER TABLE public.team_metadata OWNER TO console;

--
-- Name: teams; Type: TABLE; Schema: public; Owner: console
--

CREATE TABLE public.teams (
    id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by_id uuid,
    updated_by_id uuid,
    updated_at timestamp with time zone,
    slug text NOT NULL,
    name text NOT NULL,
    purpose text
);


ALTER TABLE public.teams OWNER TO console;

--
-- Name: user_roles; Type: TABLE; Schema: public; Owner: console
--

CREATE TABLE public.user_roles (
    id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by_id uuid,
    updated_by_id uuid,
    updated_at timestamp with time zone,
    role_id uuid NOT NULL,
    user_id uuid NOT NULL,
    target_id uuid
);


ALTER TABLE public.user_roles OWNER TO console;

--
-- Name: user_teams; Type: TABLE; Schema: public; Owner: console
--

CREATE TABLE public.user_teams (
    id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by_id uuid,
    updated_by_id uuid,
    updated_at timestamp with time zone,
    user_id uuid NOT NULL,
    team_id uuid NOT NULL
);


ALTER TABLE public.user_teams OWNER TO console;

--
-- Name: users; Type: TABLE; Schema: public; Owner: console
--

CREATE TABLE public.users (
    id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by_id uuid,
    updated_by_id uuid,
    updated_at timestamp with time zone,
    email text NOT NULL,
    name text NOT NULL
);


ALTER TABLE public.users OWNER TO console;

--
-- Name: api_keys api_keys_api_key_key; Type: CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_api_key_key UNIQUE (api_key);


--
-- Name: api_keys api_keys_pkey; Type: CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_pkey PRIMARY KEY (id);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: authorizations authorizations_name_key; Type: CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.authorizations
    ADD CONSTRAINT authorizations_name_key UNIQUE (name);


--
-- Name: authorizations authorizations_pkey; Type: CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.authorizations
    ADD CONSTRAINT authorizations_pkey PRIMARY KEY (id);


--
-- Name: correlations correlations_pkey; Type: CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.correlations
    ADD CONSTRAINT correlations_pkey PRIMARY KEY (id);


--
-- Name: reconcile_errors reconcile_errors_pkey; Type: CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.reconcile_errors
    ADD CONSTRAINT reconcile_errors_pkey PRIMARY KEY (id);


--
-- Name: role_authorizations role_authorizations_pkey; Type: CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.role_authorizations
    ADD CONSTRAINT role_authorizations_pkey PRIMARY KEY (authorization_id, role_id);


--
-- Name: roles roles_name_key; Type: CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_name_key UNIQUE (name);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: system_states system_states_pkey; Type: CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.system_states
    ADD CONSTRAINT system_states_pkey PRIMARY KEY (id);


--
-- Name: systems systems_pkey; Type: CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.systems
    ADD CONSTRAINT systems_pkey PRIMARY KEY (id);


--
-- Name: team_metadata team_metadata_pkey; Type: CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.team_metadata
    ADD CONSTRAINT team_metadata_pkey PRIMARY KEY (id);


--
-- Name: teams teams_name_key; Type: CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_name_key UNIQUE (name);


--
-- Name: teams teams_pkey; Type: CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_pkey PRIMARY KEY (id);


--
-- Name: teams teams_slug_key; Type: CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_slug_key UNIQUE (slug);


--
-- Name: user_roles user_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY (id);


--
-- Name: user_teams user_teams_pkey; Type: CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.user_teams
    ADD CONSTRAINT user_teams_pkey PRIMARY KEY (id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: correlation_system_team_key; Type: INDEX; Schema: public; Owner: console
--

CREATE UNIQUE INDEX correlation_system_team_key ON public.reconcile_errors USING btree (correlation_id, system_id, team_id);


--
-- Name: idx_api_keys_created_at; Type: INDEX; Schema: public; Owner: console
--

CREATE INDEX idx_api_keys_created_at ON public.api_keys USING btree (created_at);


--
-- Name: idx_audit_logs_action; Type: INDEX; Schema: public; Owner: console
--

CREATE INDEX idx_audit_logs_action ON public.audit_logs USING btree (action);


--
-- Name: idx_audit_logs_created_at; Type: INDEX; Schema: public; Owner: console
--

CREATE INDEX idx_audit_logs_created_at ON public.audit_logs USING btree (created_at);


--
-- Name: idx_authorizations_created_at; Type: INDEX; Schema: public; Owner: console
--

CREATE INDEX idx_authorizations_created_at ON public.authorizations USING btree (created_at);


--
-- Name: idx_correlations_created_at; Type: INDEX; Schema: public; Owner: console
--

CREATE INDEX idx_correlations_created_at ON public.correlations USING btree (created_at);


--
-- Name: idx_reconcile_errors_created_at; Type: INDEX; Schema: public; Owner: console
--

CREATE INDEX idx_reconcile_errors_created_at ON public.reconcile_errors USING btree (created_at);


--
-- Name: idx_roles_created_at; Type: INDEX; Schema: public; Owner: console
--

CREATE INDEX idx_roles_created_at ON public.roles USING btree (created_at);


--
-- Name: idx_system_states_created_at; Type: INDEX; Schema: public; Owner: console
--

CREATE INDEX idx_system_states_created_at ON public.system_states USING btree (created_at);


--
-- Name: idx_system_states_system_id; Type: INDEX; Schema: public; Owner: console
--

CREATE INDEX idx_system_states_system_id ON public.system_states USING btree (system_id);


--
-- Name: idx_system_states_team_id; Type: INDEX; Schema: public; Owner: console
--

CREATE INDEX idx_system_states_team_id ON public.system_states USING btree (team_id);


--
-- Name: idx_systems_created_at; Type: INDEX; Schema: public; Owner: console
--

CREATE INDEX idx_systems_created_at ON public.systems USING btree (created_at);


--
-- Name: idx_systems_name; Type: INDEX; Schema: public; Owner: console
--

CREATE UNIQUE INDEX idx_systems_name ON public.systems USING btree (name);


--
-- Name: idx_team_metadata_created_at; Type: INDEX; Schema: public; Owner: console
--

CREATE INDEX idx_team_metadata_created_at ON public.team_metadata USING btree (created_at);


--
-- Name: idx_teams_created_at; Type: INDEX; Schema: public; Owner: console
--

CREATE INDEX idx_teams_created_at ON public.teams USING btree (created_at);


--
-- Name: idx_user_roles_created_at; Type: INDEX; Schema: public; Owner: console
--

CREATE INDEX idx_user_roles_created_at ON public.user_roles USING btree (created_at);


--
-- Name: idx_user_teams_created_at; Type: INDEX; Schema: public; Owner: console
--

CREATE INDEX idx_user_teams_created_at ON public.user_teams USING btree (created_at);


--
-- Name: idx_users_created_at; Type: INDEX; Schema: public; Owner: console
--

CREATE INDEX idx_users_created_at ON public.users USING btree (created_at);


--
-- Name: system_team_key; Type: INDEX; Schema: public; Owner: console
--

CREATE UNIQUE INDEX system_team_key ON public.system_states USING btree (system_id, team_id);


--
-- Name: team_key; Type: INDEX; Schema: public; Owner: console
--

CREATE UNIQUE INDEX team_key ON public.team_metadata USING btree (team_id, key);


--
-- Name: user_role_target; Type: INDEX; Schema: public; Owner: console
--

CREATE UNIQUE INDEX user_role_target ON public.user_roles USING btree (role_id, user_id, target_id);


--
-- Name: user_teams_index; Type: INDEX; Schema: public; Owner: console
--

CREATE UNIQUE INDEX user_teams_index ON public.user_teams USING btree (user_id, team_id);


--
-- Name: api_keys fk_api_keys_created_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT fk_api_keys_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: api_keys fk_api_keys_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT fk_api_keys_updated_by FOREIGN KEY (updated_by_id) REFERENCES public.users(id);


--
-- Name: api_keys fk_api_keys_user; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT fk_api_keys_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: audit_logs fk_audit_logs_actor; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT fk_audit_logs_actor FOREIGN KEY (actor_id) REFERENCES public.users(id);


--
-- Name: audit_logs fk_audit_logs_correlation; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT fk_audit_logs_correlation FOREIGN KEY (correlation_id) REFERENCES public.correlations(id);


--
-- Name: audit_logs fk_audit_logs_created_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT fk_audit_logs_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: audit_logs fk_audit_logs_target_system; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT fk_audit_logs_target_system FOREIGN KEY (target_system_id) REFERENCES public.systems(id);


--
-- Name: audit_logs fk_audit_logs_target_user; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT fk_audit_logs_target_user FOREIGN KEY (target_user_id) REFERENCES public.users(id);


--
-- Name: audit_logs fk_audit_logs_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT fk_audit_logs_updated_by FOREIGN KEY (updated_by_id) REFERENCES public.users(id);


--
-- Name: authorizations fk_authorizations_created_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.authorizations
    ADD CONSTRAINT fk_authorizations_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: authorizations fk_authorizations_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.authorizations
    ADD CONSTRAINT fk_authorizations_updated_by FOREIGN KEY (updated_by_id) REFERENCES public.users(id);


--
-- Name: correlations fk_correlations_created_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.correlations
    ADD CONSTRAINT fk_correlations_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: correlations fk_correlations_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.correlations
    ADD CONSTRAINT fk_correlations_updated_by FOREIGN KEY (updated_by_id) REFERENCES public.users(id);


--
-- Name: reconcile_errors fk_reconcile_errors_correlation; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.reconcile_errors
    ADD CONSTRAINT fk_reconcile_errors_correlation FOREIGN KEY (correlation_id) REFERENCES public.correlations(id);


--
-- Name: reconcile_errors fk_reconcile_errors_created_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.reconcile_errors
    ADD CONSTRAINT fk_reconcile_errors_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: reconcile_errors fk_reconcile_errors_system; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.reconcile_errors
    ADD CONSTRAINT fk_reconcile_errors_system FOREIGN KEY (system_id) REFERENCES public.systems(id);


--
-- Name: reconcile_errors fk_reconcile_errors_team; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.reconcile_errors
    ADD CONSTRAINT fk_reconcile_errors_team FOREIGN KEY (team_id) REFERENCES public.teams(id);


--
-- Name: reconcile_errors fk_reconcile_errors_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.reconcile_errors
    ADD CONSTRAINT fk_reconcile_errors_updated_by FOREIGN KEY (updated_by_id) REFERENCES public.users(id);


--
-- Name: role_authorizations fk_role_authorizations_authorization; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.role_authorizations
    ADD CONSTRAINT fk_role_authorizations_authorization FOREIGN KEY (authorization_id) REFERENCES public.authorizations(id);


--
-- Name: role_authorizations fk_role_authorizations_role; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.role_authorizations
    ADD CONSTRAINT fk_role_authorizations_role FOREIGN KEY (role_id) REFERENCES public.roles(id);


--
-- Name: roles fk_roles_created_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT fk_roles_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: roles fk_roles_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT fk_roles_updated_by FOREIGN KEY (updated_by_id) REFERENCES public.users(id);


--
-- Name: system_states fk_system_states_created_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.system_states
    ADD CONSTRAINT fk_system_states_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: system_states fk_system_states_system; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.system_states
    ADD CONSTRAINT fk_system_states_system FOREIGN KEY (system_id) REFERENCES public.systems(id);


--
-- Name: system_states fk_system_states_team; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.system_states
    ADD CONSTRAINT fk_system_states_team FOREIGN KEY (team_id) REFERENCES public.teams(id);


--
-- Name: system_states fk_system_states_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.system_states
    ADD CONSTRAINT fk_system_states_updated_by FOREIGN KEY (updated_by_id) REFERENCES public.users(id);


--
-- Name: systems fk_systems_created_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.systems
    ADD CONSTRAINT fk_systems_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: systems fk_systems_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.systems
    ADD CONSTRAINT fk_systems_updated_by FOREIGN KEY (updated_by_id) REFERENCES public.users(id);


--
-- Name: team_metadata fk_team_metadata_created_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.team_metadata
    ADD CONSTRAINT fk_team_metadata_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: team_metadata fk_team_metadata_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.team_metadata
    ADD CONSTRAINT fk_team_metadata_updated_by FOREIGN KEY (updated_by_id) REFERENCES public.users(id);


--
-- Name: audit_logs fk_teams_audit_logs; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT fk_teams_audit_logs FOREIGN KEY (target_team_id) REFERENCES public.teams(id);


--
-- Name: teams fk_teams_created_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT fk_teams_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: team_metadata fk_teams_metadata; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.team_metadata
    ADD CONSTRAINT fk_teams_metadata FOREIGN KEY (team_id) REFERENCES public.teams(id);


--
-- Name: teams fk_teams_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT fk_teams_updated_by FOREIGN KEY (updated_by_id) REFERENCES public.users(id);


--
-- Name: user_roles fk_user_roles_created_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT fk_user_roles_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: user_roles fk_user_roles_role; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT fk_user_roles_role FOREIGN KEY (role_id) REFERENCES public.roles(id);


--
-- Name: user_roles fk_user_roles_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT fk_user_roles_updated_by FOREIGN KEY (updated_by_id) REFERENCES public.users(id);


--
-- Name: user_teams fk_user_teams_created_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.user_teams
    ADD CONSTRAINT fk_user_teams_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: user_teams fk_user_teams_team; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.user_teams
    ADD CONSTRAINT fk_user_teams_team FOREIGN KEY (team_id) REFERENCES public.teams(id);


--
-- Name: user_teams fk_user_teams_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.user_teams
    ADD CONSTRAINT fk_user_teams_updated_by FOREIGN KEY (updated_by_id) REFERENCES public.users(id);


--
-- Name: user_teams fk_user_teams_user; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.user_teams
    ADD CONSTRAINT fk_user_teams_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: users fk_users_created_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT fk_users_created_by FOREIGN KEY (created_by_id) REFERENCES public.users(id);


--
-- Name: user_roles fk_users_role_bindings; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT fk_users_role_bindings FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: users fk_users_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: console
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT fk_users_updated_by FOREIGN KEY (updated_by_id) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--

