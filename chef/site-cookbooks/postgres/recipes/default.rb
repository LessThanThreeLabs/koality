include_recipe "apt"

apt_repository 'apt.postgresql.org' do
  uri 'http://apt.postgresql.org/pub/repos/apt'
  distribution node["lsb"]["codename"] + '-pgdg'
  components ['main', node["postgres"]["version"]]
  key 'http://apt.postgresql.org/pub/repos/apt/ACCC4CF8.asc'
  action :add
end

package 'postgresql-' + node["postgres"]["version"] do
	action	:install
end

file "/etc/postgresql/#{node['postgres']['version']}/main/postgresql.conf" do
	action	:delete
end

link "/etc/postgresql/9.3/main/postgresql.conf" do
	to		node["postgres"]["conf_path"]
	action	:create
end

service "postgresql" do
	action	:restart
end
