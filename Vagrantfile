# -*- mode: ruby -*-
# vi: set ft=ruby :

VAGRANTFILE_API_VERSION = "2"
KOALITY_HOME_DIRECTORY = "/home/koality"

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

	config.vm.synced_folder "code/", "#{KOALITY_HOME_DIRECTORY}/code", nfs: true
	config.vm.synced_folder "nginx/", "#{KOALITY_HOME_DIRECTORY}/nginx", nfs: true
	config.vm.synced_folder "postgres/", "#{KOALITY_HOME_DIRECTORY}/postgres", nfs: true

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
		chef.add_recipe "nginx"
		chef.add_recipe "postgres"

		chef.json = {
			:oh_my_zsh => {
				:users => [
					{
					:login => "koality",
					:theme => {:name => "bbland", :source => "https://gist.github.com/BrianBland/7884348/raw/934802429044760bc5a2b90c773e71b13d261563/bbland.zsh-theme" },
					:plugins => ["git", "golang"]
					}
				]
			},
			:go => {
				:version => "1.2",
				:gopath => "#{KOALITY_HOME_DIRECTORY}/code/back",
				:gobin => "#{KOALITY_HOME_DIRECTORY}/code/back/bin"
			},
			:nginx => {
				:conf_path => "#{KOALITY_HOME_DIRECTORY}/nginx/nginx.conf"
			}
		}
	end
end
