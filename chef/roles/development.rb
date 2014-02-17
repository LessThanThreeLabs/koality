name "development"
description "Configures a Koality development environment"
run_list "role[koality]", "role[license_server]", "recipe[vim]", "recipe[oh-my-zsh]"
default_attributes :koality => {
		:location => "/home/koality/koality"
	}, :go => {
		:version => "1.2",
		:gopath => "/etc/koality/current/code/back",
		:gobin => "/etc/koality/current/code/back/bin"
	}, :nodejs => {
		:version => "0.10.15",
		:install_method => "package"
	}, :postgres => {
		:database_names => [
			"koality",
			"koalityLicense"
		]
	}, :oh_my_zsh => {
		:users => [
			{
				:login => "koality",
				:theme => {:name => "bbland", :source => "https://gist.github.com/BrianBland/7884348/raw/934802429044760bc5a2b90c773e71b13d261563/bbland.zsh-theme" },
				:plugins => ["git", "golang"]
			}
		]
	}
