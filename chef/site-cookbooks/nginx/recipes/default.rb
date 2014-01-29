include_recipe "apt"

apt_repository "nginx" do
	uri          "http://nginx.org/packages/ubuntu/"
	distribution node["lsb"]["codename"]
	components   ["nginx"]
	key          "http://nginx.org/keys/nginx_signing.key"
	deb_src      true
end

package 'nginx' do
	action :remove
	notifies :run, 'execute[apt-get autoremove]', :immediately
	not_if "nginx -v 2>&1 | grep -q #{node['nginx']['version'][/^[0-9\\.]+/]}"
end

apt_package "nginx" do
	version		node["nginx"]["version"]
	action		:install
end

cookbook_file "/etc/nginx/nginx.conf" do
	owner 		"nginx"
	group 		"nginx"
	source		"nginx.conf"
	action	 	:create
end

service "nginx" do
	action [:enable, :restart]
	supports :status=>true, :restart=>true, :start=>true, :stop=>true, :reload=>true
end
