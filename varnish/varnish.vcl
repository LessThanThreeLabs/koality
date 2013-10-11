backend node0 {
	.host = "127.0.0.1";
	.port = "1080";
	.probe = {
		.url = "/ping";
		.interval = 5s;
		.timeout = 1s;
		.window = 5;
		.threshold = 3;
	}
}

backend node1 {
	.host = "127.0.0.1";
	.port = "1081";
	.probe = {
		.url = "/ping";
		.interval = 5s;
		.timeout = 1s;
		.window = 5;
		.threshold = 3;
	}
}

director default_director round-robin {
	{ .backend = node0; }
	{ .backend = node1; }
}

sub vcl_recv {
	if (req.http.X-Forwarded-Proto && req.http.X-Forwarded-Proto !~ "(?i)https") {
		error 750 "https://koalitycode.com";
	}

	set req.backend = default_director;
	if (req.request == "GET" && req.url ~ "^/(html|js|css|img|font)") {
		unset req.http.cookie;
		unset req.http.Authorization;
		return(lookup);
	}
}

sub vcl_fetch {
	if (req.request == "GET" && req.url ~ "^/(html|js|css|img|font)") {
		unset beresp.http.Set-Cookie;
		set beresp.ttl = 30m;
		return(deliver);
	}
}

sub vcl_error {
	if (obj.status == 750) {
		set obj.http.Location = obj.response;
		set obj.status = 301;
		return(deliver);
	}
}
