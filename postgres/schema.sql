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
	password_hash 		bytea NOT NULL,
	password_salt 		bytea NOT NULL,
	github_oauth 		varchar(40),
	is_admin			boolean NOT NULL DEFAULT false,
	created 			timestamp with time zone NOT NULL DEFAULT current_timestamp,
	deleted 			integer NOT NULL DEFAULT 0,  -- set to id when deleted

	UNIQUE (email, deleted),
	CHECK (deleted = 0 OR deleted = id)
);

CREATE TABLE IF NOT EXISTS ssh_keys (
	id 					serial PRIMARY KEY,
	user_id 			integer NOT NULL references users(id) ON DELETE CASCADE,
	name				varchar(256) NOT NULL,
	public_key			varchar(1024) NOT NULL,
	created 			timestamp with time zone NOT NULL DEFAULT current_timestamp,

	UNIQUE (user_id, name),
	UNIQUE (public_key)
);

CREATE TABLE IF NOT EXISTS repositories (
	id 					serial PRIMARY KEY,
	name 				varchar(256) NOT NULL,
	status 				varchar(32) NOT NULL,
	vcs_type			varchar(32) NOT NULL,
	remote_uri			varchar(1024) NOT NULL,
	created 			timestamp with time zone NOT NULL DEFAULT current_timestamp,
	deleted 			integer NOT NULL DEFAULT 0,  -- set to id when deleted

	UNIQUE (name, deleted),
	UNIQUE (remote_uri, deleted),
	CHECK (deleted = 0 OR deleted = id)
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

CREATE TABLE IF NOT EXISTS ec2_pools (
	id 					serial PRIMARY KEY,
	name 				varchar(256) NOT NULL,
	access_key 			varchar(20) NOT NULL,
	secret_key 			varchar(40) NOT NULL,
	username 			varchar(256) NOT NULL,
	base_ami_id 		varchar(12),
	security_group_id 	varchar(11),
	vpc_subnet_id 		varchar(15),
	instance_type 		varchar(64) NOT NULL,
	num_ready_instances integer NOT NULL,
	num_max_instances 	integer NOT NULL,
	root_drive_size 	integer NOT NULL,  -- in GB
	user_data 			varchar(1000000),  -- 1MB
	created 			timestamp with time zone NOT NULL DEFAULT current_timestamp,
	deleted 			integer NOT NULL DEFAULT 0,  -- set to id when deleted

	UNIQUE (name, deleted),
	CHECK (deleted = 0 OR deleted = id),
	CHECK (num_ready_instances <= num_max_instances)
);

CREATE TABLE IF NOT EXISTS snapshots (
	id 					serial PRIMARY KEY,
	pool_id				integer NOT NULL references ec2_pools(id) ON DELETE CASCADE,
	image_id			varchar(12),
	image_type			varchar(20) NOT NULL,
	status 				varchar(32) NOT NULL,
	created 			timestamp with time zone NOT NULL DEFAULT current_timestamp,
	started				timestamp with time zone,
	ended				timestamp with time zone,
	deleted 			integer NOT NULL DEFAULT 0,  -- set to id when deleted

	UNIQUE (image_id),
	CHECK (deleted = 0 OR deleted = id),
	CHECK (started IS NULL OR created <= started),
	CHECK (started IS NULL OR ended IS NULL OR started <= ended)
);

CREATE TABLE IF NOT EXISTS verifications (
	id 					serial PRIMARY KEY,
	repository_id		integer NOT NULL references repositories(id) ON DELETE CASCADE,
	snapshot_id			integer references snapshots(id) ON DELETE CASCADE,
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

CREATE TABLE IF NOT EXISTS console_lines (
	id 					serial PRIMARY KEY,
	run_id				integer NOT NULL references stage_runs(id) ON DELETE CASCADE,
	number				integer NOT NULL,
	text 				text NOT NULL
);

-- DO
-- $create_console_lines_run_id_idx$
-- BEGIN
-- 	IF NOT EXISTS (SELECT 1 FROM pg_class where relname = 'console_lines_run_id_idx') THEN
-- 		CREATE INDEX console_lines_run_id_idx ON console_lines(run_id);
-- 	END IF;
-- END
-- $create_console_lines_run_id_idx$;

-- DO
-- $create_console_lines_number_idx$
-- BEGIN
-- 	IF NOT EXISTS (SELECT 1 FROM pg_class where relname = 'console_lines_number_idx') THEN
-- 		CREATE INDEX console_lines_number_idx ON console_lines(number);
-- 	END IF;
-- END
-- $create_console_lines_number_idx$;

CREATE TABLE IF NOT EXISTS xunit_results (
	id 					serial PRIMARY KEY,
	run_id				integer NOT NULL references stage_runs(id) ON DELETE CASCADE,
	name 				text NOT NULL,
	path 				text NOT NULL,
	sysout 				text,
	syserr 				text,
	failure_text 		text,
	error_text 			text,
	seconds				double precision NOT NULL
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
	bucket				varchar(1024) NOT NULL,
	path				varchar(1024) NOT NULL,
	key 				varchar(1024) NOT NULL,

	UNIQUE (run_id, path),
	UNIQUE (run_id, key)
);

CREATE TABLE IF NOT EXISTS settings (
	id 					serial PRIMARY KEY,
	resource			varchar(256) NOT NULL,
	key 				varchar(256) NOT NULL,
	value 				bytea NOT NULL,

	UNIQUE (resource, key)
);

CREATE TABLE IF NOT EXISTS version (
	id 					serial PRIMARY KEY,
	version 			integer NOT NULL
);

DO
$insert_version$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM version) THEN
		INSERT INTO version (version) VALUES (0);  -- change this whenever we change the schema
	END IF;
END
$insert_version$
