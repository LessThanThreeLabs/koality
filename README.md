![Koala](http://i.imgur.com/dquB6fL.png)

#Getting started
1. Download Vagrant [v1.3.5](http://downloads.vagrantup.com/tags/v1.3.5)
2. Download VirtualBox [v4.2.18](https://www.virtualbox.org/wiki/Download_Old_Builds_4_2)
3. Run `git clone --recursive git@github.com:LessThanThreeLabs/koality.git`


#Commands to know

###Git
* `git submodule add git@domain:user/repo.git path/to/location`
	* Add a new submodule to the codebase
* `git submodule update --init --recursive`
	* This will pull down updates to submodules
	* Use this if you're missing a library
* `git submodule sync`
	* This will update the remote urls for all your submodules
	* Use this if updating submodules fails

###Vagrant

* `vagrant up`
	* This will start the virtual machine with Koality code inside
	* Services such as Nginx, RabbitMQ, etc. will be running
* `vagrant provision`
	* Re-provisions the codebase to include the latest libraries, etc.
* `vagrant suspend`
	* Will put the virtual machine to sleep
	* Be sure to do this before putting your computer to sleep
* `vagrant reload`
	* Will restart the virtual machine
	* Use this if the virtual machine is ever giving you trouble
* `sudo /Library/StartupItems/VirtualBox/VirtualBox restart`
	* When you see "Error while adding new interface..."
