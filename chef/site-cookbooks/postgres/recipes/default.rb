include_recipe "apt"

# apt_repository "postgres" do
# 	uri          "http://apt.postgresql.org/pub/repos/apt/"
# 	distribution node["lsb"]["codename"]
# 	components   ["postgres"]
# 	key          "https://www.postgresql.org/media/keys/ACCC4CF8.asc"
# 	deb_src      true
# end

apt_package "postgresql" do
	version	node["postgres"]["version"]
	action	:install
end

# file "/etc/postgres/postgres.conf" do
# 	action	:delete
# end

# link "/etc/postgres/postgres.conf" do
# 	to			node["postgres"]["conf_path"]
# 	action	:create
# end

# service "postgres" do
# 	action	:restart
# end