# -*- mode: ruby -*-
# vi: set ft=ruby :

VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
	config.vm.box = "precise64"
	config.vm.box_url = "http://files.vagrantup.com/precise64.box"

	config.vm.network "public_network"

	config.vm.provision :chef_solo do |chef|
		chefRoot = "chef"
		chef.cookbooks_path = ["#{chefRoot}/cookbooks", "#{chefRoot}/site-cookbooks"]
		chef.data_bags_path = "#{chefRoot}/databags"

		chef.add_recipe "apt"
		chef.add_recipe "build-essential"
		chef.add_recipe "vim"
		chef.add_recipe "git"
		chef.add_recipe "golang"
		chef.add_recipe "erlang"
		chef.add_recipe "rabbitmq"
		chef.add_recipe "varnish"

		chef.json = {
			:go => {
				:version => "1.1.2"
			}
		}
	end
end
