include_recipe "apt"

apt_repository "nginx" do
	uri          "http://nginx.org/packages/ubuntu/"
	distribution node["lsb"]["codename"]
	components   ["nginx"]
	key          "http://nginx.org/keys/nginx_signing.key"
	deb_src      true
end

apt_package "nginx" do
	version	node["nginx"]["version"]
	action	:install
end

file "/etc/nginx/nginx.conf" do
	action	:delete
end

link "/etc/nginx/nginx.conf" do
	to		node["nginx"]["conf_path"]
	action	:create
end

service "nginx" do
	action	:restart
end