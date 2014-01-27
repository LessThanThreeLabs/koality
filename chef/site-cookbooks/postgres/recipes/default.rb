include_recipe "apt"

apt_repository 'apt.postgresql.org' do
  uri 'http://apt.postgresql.org/pub/repos/apt'
  distribution node["lsb"]["codename"] + '-pgdg'
  components ['main', node["postgres"]["version"]]
  key 'http://apt.postgresql.org/pub/repos/apt/ACCC4CF8.asc'
  action :add
end

package 'postgresql' do
	action :remove
	notifies :run, 'execute[apt-get autoremove]', :immediately
	not_if "psql --version | grep -q #{node['postgres']['version']}"
end

package 'postgresql-' + node["postgres"]["version"] do
	action		:install
end

cookbook_file "/etc/postgresql/#{node['postgres']['version']}/main/postgresql.conf" do
	source		"postgresql.conf"
	action	 	:create
end

service "postgresql" do
	action 		[:enable, :start]
	supports 	:status=>true, :restart=>true, :start => true, :stop => true, :reload=>true
end

execute "create-role" do
	user "postgres"
	command "psql -c \"CREATE USER #{node[:postgres][:username]} PASSWORD '#{node[:postgres][:password]}' SUPERUSER\""
	not_if "psql -c \"SELECT 1 FROM pg_user WHERE usename='#{node[:postgres][:username]}'\" | grep -q 1", :user => "postgres"
end

execute "create-database" do
	user "postgres"
	command "createdb #{node[:postgres][:database_name]} --template template0 --locale #{node[:postgres][:locale]} --encoding #{node[:postgres][:character_encoding]}"
	not_if "psql -c \"SELECT 1 FROM pg_database WHERE datname='#{node[:postgres][:database_name]}'\" | grep -q 1", :user => "postgres"
end
