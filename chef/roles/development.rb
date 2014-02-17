name "development"
description "Configures a Koality development environment"
run_list "role[koality]", "role[license_server]", "recipe[vim]", "recipe[oh-my-zsh]"
default_attributes :koality => {
		:location => "/home/koality/koality"
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
