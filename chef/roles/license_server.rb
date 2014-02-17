name "license_server"
description "Configures a machine to run the Koality license server"
run_list "recipe[koality]", "recipe[apt]", "recipe[build-essential]", "recipe[gdb]",
	"recipe[git]", "recipe[hg]", "recipe[golang]", "recipe[nginx]", "recipe[postgres]",
	"recipe[nodejs]", "recipe[npm]", "recipe[icedcoffeescript]", "recipe[grunt]"
default_attributes :go => {
		:version => "1.2",
		:gopath => "/etc/koality/current/code/back",
		:gobin => "/etc/koality/current/code/back/bin"
	}, :nodejs => {
		:version => "0.10.15",
		:install_method => "package"
	}, :postgres => {
		:database_names => [
			"koalityLicense"
		]
	}
