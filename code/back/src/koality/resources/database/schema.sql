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
	password_hash 		varchar(100) NOT NULL,
	password_salt 		varchar(64) NOT NULL,
	github_oauth 		varchar(40),
	admin 				boolean DEFAULT false,
	created 			timestamp with time zone NOT NULL,
	deleted 			timestamp with time zone,

	UNIQUE (email, deleted)
);

CREATE TABLE IF NOT EXISTS ssh_keys (
	id 					serial PRIMARY KEY,
	user_id 			integer NOT NULL references users(id),
	alias				varchar(256) NOT NULL,
	public_key			varchar(1024) NOT NULL,
	created				timestamp with time zone NOT NULL,

	UNIQUE (user_id, alias),
	UNIQUE (public_key)
);

CREATE TABLE IF NOT EXISTS repositories (
	id 					serial PRIMARY KEY,
	type				varchar(32) NOT NULL,
	name 				varchar(256) NOT NULL,
	local_uri			varchar(1024) NOT NULL,
	remote_uri			varchar(1024) NOT NULL,
	created 			timestamp with time zone NOT NULL,
	deleted 			timestamp with time zone,

	UNIQUE (name, deleted),
	UNIQUE (local_uri, deleted),
	UNIQUE (remote_uri, deleted)
);

CREATE TABLE IF NOT EXISTS repository_github_metadatas (
	id 					serial PRIMARY KEY,
	repository_id		integer NOT NULL references repositories(id),
	owner_name			varchar(256) NOT NULL,
	repository_name		varchar(256) NOT NULL,  -- name on GitHub
	hook_id				integer,
	hook_secret			varchar(32),
	hook_types			varchar(64)[],

	UNIQUE (repository_id)
);

CREATE TABLE IF NOT EXISTS changesets (
	id 					serial PRIMARY KEY,
	user_id 			integer NOT NULL references users(id),
	repository_id		integer NOT NULL references repositories(id),
	head_sha			varchar(40) NOT NULL,
	base_sha			varchar(40),
	head_message		varchar(1000000), -- 1MB
	head_username		varchar(128),
	head_email			varchar(256),

	UNIQUE (head_sha, base_sha)
);

CREATE TABLE IF NOT EXISTS changes (
	id 					serial PRIMARY KEY,
	commit_id			integer NOT NULL references changesets(id),
	repository_id		integer NOT NULL references repositories(id),
	merge_target		varchar(1024),
	verification_status varchar(32),
	merge_status		varchar(32),
	created				timestamp with time zone NOT NULL,
	started				timestamp with time zone,
	ended				timestamp with time zone
);

CREATE TABLE IF NOT EXISTS stages (
	id 					serial PRIMARY KEY,
	change_id			integer NOT NULL references changes(id),
	name 				varchar(1024) NOT NULL,
	type 				varchar(32) NOT NULL,
	order_number 		integer NOT NULL,

	UNIQUE (change_id, name, type)
);

CREATE TABLE IF NOT EXISTS runs (
	id 					serial PRIMARY KEY,
	stage_id			integer NOT NULL references stages(id),
	return_code			integer,
	created				timestamp with time zone NOT NULL,
	started				timestamp with time zone,
	ended				timestamp with time zone
);

CREATE TABLE IF NOT EXISTS console_texts (
	id 					serial PRIMARY KEY,
	run_id				integer NOT NULL references runs(id),
	line_number			integer NOT NULL,
	line 				text NOT NULL,

	UNIQUE (run_id, line_number)
);

CREATE TABLE IF NOT EXISTS xunits (
	id 					serial PRIMARY KEY,
	run_id				integer NOT NULL references runs(id),
	path				varchar(1024) NOT NULL,
	contents			text NOT NULL,

	UNIQUE (run_id, path)
);

CREATE TABLE IF NOT EXISTS exports (
	id 					serial PRIMARY KEY,
	run_id				integer NOT NULL references runs(id),
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

DO
$create_super_admins$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM users WHERE id = 1) THEN
		INSERT INTO users (id, email, first_name, last_name, password_hash, password_salt, admin, created) 
		VALUES (1, 'admin-koala@koalitycode.com', 'Admin', 'Koala', 
			'mooonIJXsb0zgz2V0LXvN/N4N4zbZE9FadrFl/YBJvzh3Z8O3VT/FH1q6OzWplbrX99D++PO6mpez7QdoIUQ6A==',
			'GMZhGiZU4/JYE3NlmCZgGA==', true, current_timestamp);
	END IF;

	IF NOT EXISTS (SELECT 1 FROM users WHERE id = 2) THEN
		INSERT INTO users (id, email, first_name, last_name, password_hash, password_salt, admin, created)
		VALUES (2, 'api`-koala@koalitycode.com', 'Api', 'Koala',
			'mooonIJXsb0zgz2V0LXvN/N4N4zbZE9FadrFl/YBJvzh3Z8O3VT/FH1q6OzWplbrX99D++PO6mpez7QdoIUQ6A==',
			'GMZhGiZU4/JYE3NlmCZgGA==', true, current_timestamp);
	END IF;

	IF NOT EXISTS (SELECT 1 FROM users WHERE id = 3) THEN
		INSERT INTO users (id, email, first_name, last_name, password_hash, password_salt, admin, created)
		VALUES (3, 'verifier-koala@koalitycode.com', 'Verifier', 'Koala',
			'mooonIJXsb0zgz2V0LXvN/N4N4zbZE9FadrFl/YBJvzh3Z8O3VT/FH1q6OzWplbrX99D++PO6mpez7QdoIUQ6A==',
			'GMZhGiZU4/JYE3NlmCZgGA==', true, current_timestamp);
	END IF;
END
$create_super_admins$;
