DO
$create_user_id_seq$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_class where relname = 'user_id_seq') THEN
		CREATE SEQUENCE user_id_seq START WITH 1000;
	END IF;
END
$create_user_id_seq$;

CREATE TABLE IF NOT EXISTS users (
	id 					integer PRIMARY KEY DEFAULT nextval('user_id_seq'),
	email 				varchar(256) NOT NULL,
	first_name 			varchar(64) NOT NULL,
	last_name 			varchar(64) NOT NULL,
	password_hash 		varchar(100) NOT NULL,  -- base64 encoding
	password_salt 		varchar(64) NOT NULL,   -- base64 encoding
	github_oauth 		varchar(40),
	is_admin			boolean NOT NULL DEFAULT false,
	created 			timestamp with time zone NOT NULL DEFAULT current_timestamp,
	deleted 			integer NOT NULL DEFAULT 0,  -- set to id when deleted

	UNIQUE (email, deleted)
);

CREATE TABLE IF NOT EXISTS ssh_keys (
	id 					serial PRIMARY KEY,
	user_id 			integer NOT NULL references users(id) ON DELETE CASCADE,
	alias				varchar(256) NOT NULL,
	public_key			varchar(1024) NOT NULL,
	created 			timestamp with time zone NOT NULL DEFAULT current_timestamp,

	UNIQUE (user_id, alias),
	UNIQUE (public_key)
);

CREATE TABLE IF NOT EXISTS repositories (
	id 					serial PRIMARY KEY,
	name 				varchar(256) NOT NULL,
	vcs_type			varchar(32) NOT NULL,
	local_uri			varchar(1024) NOT NULL,
	remote_uri			varchar(1024) NOT NULL,
	created 			timestamp with time zone DEFAULT current_timestamp,
	deleted 			integer NOT NULL DEFAULT 0,  -- set to id when deleted

	UNIQUE (name, deleted),
	UNIQUE (local_uri, deleted),
	UNIQUE (remote_uri, deleted)
);

CREATE TABLE IF NOT EXISTS repository_github_metadatas (
	id 					serial PRIMARY KEY,
	repository_id		integer NOT NULL references repositories(id) ON DELETE CASCADE,
	owner				varchar(256) NOT NULL,
	name				varchar(256) NOT NULL,
	hook_id				integer,
	hook_secret			varchar(32),
	hook_types			varchar(128),  -- e.g. "push,pull_request"

	UNIQUE (repository_id)
);

CREATE TABLE IF NOT EXISTS changesets (
	id 					serial PRIMARY KEY,
	repository_id		integer NOT NULL references repositories(id) ON DELETE CASCADE,
	head_sha			varchar(40) NOT NULL,
	base_sha			varchar(40) NOT NULL,
	head_message		varchar(1000000) NOT NULL,  -- 1MB
	head_username		varchar(128) NOT NULL,
	head_email			varchar(256) NOT NULL,
	created 			timestamp with time zone NOT NULL DEFAULT current_timestamp,

	UNIQUE (head_sha, base_sha)
);

CREATE TABLE IF NOT EXISTS verifications (
	id 					serial PRIMARY KEY,
	repository_id		integer NOT NULL references repositories(id) ON DELETE CASCADE,
	changeset_id		integer NOT NULL references changesets(id) ON DELETE CASCADE,
	merge_target		varchar(1024),
	email_to_notify		varchar(256),
	status 				varchar(32) NOT NULL,
	merge_status		varchar(32),
	created 			timestamp with time zone NOT NULL DEFAULT current_timestamp,
	started				timestamp with time zone,
	ended				timestamp with time zone,

	CHECK (started IS NULL OR created <= started),
	CHECK (started IS NULL OR ended IS NULL OR started <= ended)
);

CREATE TABLE IF NOT EXISTS stages (
	id 					serial PRIMARY KEY,
	verification_id		integer NOT NULL references verifications(id) ON DELETE CASCADE,
	section_number		integer NOT NULL,
	name 				varchar(1024) NOT NULL,
	order_number 		integer NOT NULL,

	UNIQUE (verification_id, section_number, name)
);

CREATE TABLE IF NOT EXISTS stage_runs (
	id 					serial PRIMARY KEY,
	stage_id			integer NOT NULL references stages(id) ON DELETE CASCADE,
	return_code			integer DEFAULT -1,
	created 			timestamp with time zone NOT NULL DEFAULT current_timestamp,
	started				timestamp with time zone,
	ended				timestamp with time zone,

	CHECK (started IS NULL OR created <= started),
	CHECK (started IS NULL OR ended IS NULL OR started <= ended)
);

CREATE TABLE IF NOT EXISTS console_texts (
	id 					serial PRIMARY KEY,
	run_id				integer NOT NULL references stage_runs(id) ON DELETE CASCADE,
	number				integer NOT NULL,
	text 				text NOT NULL
);

-- DO
-- $create_console_texts_run_id_idx$
-- BEGIN
-- 	IF NOT EXISTS (SELECT 1 FROM pg_class where relname = 'console_texts_run_id_idx') THEN
-- 		CREATE INDEX console_texts_run_id_idx ON console_texts(run_id);
-- 	END IF;
-- END
-- $create_console_texts_run_id_idx$;

-- DO
-- $create_console_texts_number_idx$
-- BEGIN
-- 	IF NOT EXISTS (SELECT 1 FROM pg_class where relname = 'console_texts_number_idx') THEN
-- 		CREATE INDEX console_texts_number_idx ON console_texts(number);
-- 	END IF;
-- END
-- $create_console_texts_number_idx$;

CREATE TABLE IF NOT EXISTS xunit_results (
	id 					serial PRIMARY KEY,
	run_id				integer NOT NULL references stage_runs(id) ON DELETE CASCADE,
	name 				text NOT NULL,
	path 				text NOT NULL,
	sysout 				text,
	syserr 				text,
	failure_text 		text,
	error_text 			text,
	started 			timestamp NOT NULL,
	seconds				real NOT NULL
);

-- DO
-- $create_xunit_results_run_id_idx$
-- BEGIN
-- 	IF NOT EXISTS (SELECT 1 FROM pg_class where relname = 'xunit_results_run_id_idx') THEN
-- 		CREATE INDEX xunit_results_run_id_idx ON xunit_results(run_id);
-- 	END IF;
-- END
-- $create_xunit_results_run_id_idx$;

CREATE TABLE IF NOT EXISTS exports (
	id 					serial PRIMARY KEY,
	run_id				integer NOT NULL references stage_runs(id) ON DELETE CASCADE,
	path				varchar(1024) NOT NULL,
	uri 				varchar(1024) NOT NULL,

	UNIQUE (run_id, path),
	UNIQUE (run_id, uri)
);

CREATE TABLE IF NOT EXISTS settings (
	id 					serial PRIMARY KEY,
	resource			varchar(256) NOT NULL,
	key 				varchar(256) NOT NULL,
	value 				text NOT NULL,

	UNIQUE (resource, key)
);
