include_recipe "apt"

apt_repository 'apt.postgresql.org' do
  uri 'http://apt.postgresql.org/pub/repos/apt'
  distribution node["lsb"]["codename"] + '-pgdg'
  components ['main', node["postgres"]["version"]]
  key 'http://apt.postgresql.org/pub/repos/apt/ACCC4CF8.asc'
  action :add
end

package 'postgresql-' + node["postgres"]["version"] do
	action		:install
end

file "/etc/postgresql/#{node['postgres']['version']}/main/postgresql.conf" do
	action		:delete
end

link "/etc/postgresql/#{node['postgres']['version']}/main/postgresql.conf" do
	to			node["postgres"]["conf_path"]
	owner		"postgres"
	group		"postgres"
	action		:create
	notifies 	:restart, "service[postgresql]", :delayed
end

cookbook_file "/etc/init/postgresql-run.conf" do
	source		"upstart/postgresql-run.conf"
	action	 	:create_if_missing
end

service "postgresql" do
	action 		[:enable, :start]
	supports 	:status=>true, :restart=>true, :start => true, :stop => true, :reload=>true
end

execute "create-role" do
	user "postgres"
	command "psql -c \"SELECT 1 FROM pg_user WHERE usename='#{node[:postgres][:username]}'\" | grep -q 1 || psql -c \"CREATE USER #{node[:postgres][:username]} PASSWORD '#{node[:postgres][:password]}' SUPERUSER\""
end

execute "create-database" do
	user "postgres"
	command "psql -c \"SELECT 1 FROM pg_database WHERE datname='#{node[:postgres][:database_name]}'\" | grep -q 1 || createdb #{node[:postgres][:database_name]} --template template0 --locale #{node[:postgres][:locale]} --encoding #{node[:postgres][:character_encoding]}"
end
