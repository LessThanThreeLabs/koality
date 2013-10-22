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
		v.customize ["modifyvm", :id, "--cpuexecutioncap", "100"]
		v.customize ["modifyvm", :id, "--memory", "1024"]
	end

	config.vm.synced_folder "code/", "#{VAGRANT_HOME_DIRECTORY}/code"
	config.vm.synced_folder "nginx/", "#{VAGRANT_HOME_DIRECTORY}/nginx"

	config.vm.network :forwarded_port, guest: 80,    host: 1080  # Nginx
	config.vm.network :forwarded_port, guest: 443,   host: 10443 # Nginx
	config.vm.network :forwarded_port, guest: 8080,  host: 8080  # Webserver
	config.vm.network :forwarded_port, guest: 5672,  host: 5672  # RabbitMQ
	config.vm.network :forwarded_port, guest: 15672, host: 15672 # RabbitMQ Management

	config.vm.provision :chef_solo do |chef|
		chefRoot = "chef"
		chef.cookbooks_path = ["#{chefRoot}/cookbooks", "#{chefRoot}/site-cookbooks"]
		chef.data_bags_path = "#{chefRoot}/databags"

		chef.add_recipe "apt"
		chef.add_recipe "build-essential"
		chef.add_recipe "vim"
		chef.add_recipe "git"
		chef.add_recipe "oh-my-zsh"
		chef.add_recipe "golang"
		chef.add_recipe "nodejs"
		chef.add_recipe "erlang"
		chef.add_recipe "rabbitmq"
		chef.add_recipe "nginx"

		chef.json = {
			:oh_my_zsh => {
				:users => [
					{
					:login => "vagrant",
					:theme => "minimal",
					:plugins => ["git", "golang"]
					}
				]
			},
			:go => {
				:version => "1.1.2",
				:gopath => "#{VAGRANT_HOME_DIRECTORY}/code/back",
				:gobin => "#{VAGRANT_HOME_DIRECTORY}/code/back/bin"
			},
			:nodejs => {
				:version => "0.10.21",
			},
			:rabbitmq => {
				:version => "3.1.5",
				:enabled_plugins => ["rabbitmq_management"]
			},
			:nginx => {
				:conf_path => "#{VAGRANT_HOME_DIRECTORY}/nginx/nginx.conf"
			}
		}
	end
end
