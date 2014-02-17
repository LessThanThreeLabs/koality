CREATE TABLE IF NOT EXISTS licenses (
	id 				serial PRIMARY KEY,
	key 			varchar(16) NOT NULL,
	max_executors 	integer NOT NULL,
	is_active	 	boolean NOT NULL DEFAULT true,
	server_id	 	varchar(64),
	last_ping 		timestamp with time zone
);
