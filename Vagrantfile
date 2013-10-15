# -*- mode: ruby -*-
# vi: set ft=ruby :

VAGRANTFILE_API_VERSION = "2"
VAGRANT_HOME_DIRECTORY = "/home/vagrant"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
	config.vm.box = "precise64"
	config.vm.box_url = "http://files.vagrantup.com/precise64.box"

	config.vm.provider "virtualbox" do |v|
		v.customize ["modifyvm", :id, "--name", "koality"]
		v.customize ["modifyvm", :id, "--cpus", "2"]
		v.customize ["modifyvm", :id, "--memory", "1024"]
	end

	config.vm.synced_folder "code/", "#{VAGRANT_HOME_DIRECTORY}/code"
	config.vm.synced_folder "varnish/", "#{VAGRANT_HOME_DIRECTORY}/varnish"

	config.vm.network :forwarded_port, guest: 1080,  host: 1080  # Webserver
	config.vm.network :forwarded_port, guest: 5672,  host: 5672  # RabbitMQ
	config.vm.network :forwarded_port, guest: 15672, host: 15672 # RabbitMQ Management
	config.vm.network :forwarded_port, guest: 6081,  host: 6081  # Varnish

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
				:version => "1.1.2",
				:gopath => "#{VAGRANT_HOME_DIRECTORY}/code",
				:gobin => "#{VAGRANT_HOME_DIRECTORY}/code/bin"
			},
			:rabbitmq => {
				:version => "3.1.5",
				:enabled_plugins => ["rabbitmq_management"]
			},
			:varnish => {
				:dir => "#{VAGRANT_HOME_DIRECTORY}/varnish",
				:storage => "malloc",
				:storage_size => "128M"
			}
		}
	end
end
