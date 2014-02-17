# -*- mode: ruby -*-
# vi: set ft=ruby :

VAGRANTFILE_API_VERSION = "2"
KOALITY_HOME_DIRECTORY = "/home/koality"
KOALITY_REPOSITORY_DIRECTORY = "#{KOALITY_HOME_DIRECTORY}/koality"

Vagrant.require_version ">= 1.4.0"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
	config.vm.box = "koality-v0"
	config.vm.box_url = "https://s3-us-west-2.amazonaws.com/koality-boxes/v0.box"

	config.ssh.username = "koality"

	config.vm.provider "virtualbox" do |v|
		v.customize ["modifyvm", :id, "--name", "koality"]
		v.customize ["modifyvm", :id, "--cpus", "2"]
		v.customize ["modifyvm", :id, "--cpuexecutioncap", "100"]
		v.customize ["modifyvm", :id, "--memory", "2048"]
	end

	config.vm.network "private_network", ip: "10.10.10.10"

	config.vm.network :forwarded_port, guest: 80,    host: 1080  # Nginx
	config.vm.network :forwarded_port, guest: 443,   host: 10443 # Nginx
	config.vm.network :forwarded_port, guest: 8080,  host: 8080  # Webserver

	config.vm.synced_folder ".", KOALITY_REPOSITORY_DIRECTORY, nfs: true

	config.vm.provision :chef_solo do |chef|
		chefRoot = "chef"
		chef.cookbooks_path = ["#{chefRoot}/cookbooks", "#{chefRoot}/site-cookbooks"]
		chef.data_bags_path = "#{chefRoot}/databags"
		chef.roles_path = "#{chefRoot}/roles"

		chef.add_role "development"
	end
end
