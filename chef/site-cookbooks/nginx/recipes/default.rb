include_recipe "apt"

apt_repository "nginx" do
	uri          "http://nginx.org/packages/ubuntu/"
	distribution node["lsb"]["codename"]
	components   ["nginx"]
	key          "http://nginx.org/keys/nginx_signing.key"
	deb_src      true
end

apt_package "nginx" do
	version		node["nginx"]["version"]
	action		:install
end

file "/etc/nginx/nginx.conf" do
	action		:delete
end

link "/etc/nginx/nginx.conf" do
	to			node["nginx"]["conf_path"]
	owner		"nginx"
	group		"nginx"
	action		:create
	notifies	:restart, "service[nginx]", :delayed
end

cookbook_file "/etc/init/nginx-run.conf" do
	source		"upstart/nginx-run.conf"
	action	 	:create_if_missing
end

service "nginx" do
	action [:enable, :start]
	supports :status=>true, :restart=>true, :start => true, :stop => true, :reload=>true
end
