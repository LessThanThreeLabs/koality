'use strict'

window.AdminPools = ['$scope', '$http', 'events', 'notification', ($scope, $http, events, notification) ->
	console.log 'hello'
	$scope.allowedAmiIds = ["id1", "id2", "id3"]
	$scope.allowedInstanceSizes = [2, 4, 8, 16]
	$scope.allowedSecurityGroups = [
		{
			id: "id1",
			displayName: "insecure"
		},
		{
			id: "id2",
			displayName: "kindofsecure",
		},
		{
			id: "id3",
			displayName: "verysecure"
		}]
	$scope.pool = {
		id: 5,
		name: "thepoolname",
		awsKeys: {
			accessKey: "accessKeyyyyyyyyyy",
			secretKey: "secretKeyyyyyyyyyy"
		},
		verifierPoolSettings: {
			userData: "#!/bin/sh\necho ohai > /home/lt3/ohai"
			amiId: "id",
			minReady: 0,
			maxRunning: 16,
			username: "lt3",
			instanceSize: 4,
			rootDriveSize: 15
		},
		instanceSettings: {
			securityGroupId: "id1",
			subnetId: "subnetId?"
		}
	}
]
