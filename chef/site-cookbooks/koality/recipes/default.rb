directory "/etc/koality/repositories" do
	owner 		"koality"
	group 		"koality"
	recursive	true
	mode 		0755
	action		:create
end

directory "/etc/koality/conf" do
	owner 		"koality"
	group 		"koality"
	recursive	true
	mode 		0755
	action		:create
end

link "/etc/koality/current" do
	owner 		"koality"
	group		"koality"
	to 			node["koality"]["location"]
	link_type 	:symbolic
	action 		:create
end
