name "license_server"
description "Configures a machine to run the Koality license server"
run_list "recipe[koality]", "recipe[apt]", "recipe[build-essential]", "recipe[git]", "recipe[golang]", "recipe[postgres]"
default_attributes :go => {
		:version => "1.2",
		:gopath => "/etc/koality/current/code/back",
		:gobin => "/etc/koality/current/code/back/bin"
	}, :postgres => {
		:database_names => [
			"koalityLicense"
		]
	}
