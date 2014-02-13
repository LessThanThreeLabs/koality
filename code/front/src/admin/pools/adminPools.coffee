'use strict'

window.AdminPools = ['$scope', '$http', 'events', 'notification', ($scope, $http, events, notification) ->
	console.log 'hello'
	$scope.allowedAmiIds = ["id1", "id2", "id3"]
	$scope.allowedInstanceSizes = [2, 4, 8, 16]
	$scope.allowedSecurityGroups = ["insecure", "kindofsecure", "verysecure"]
	$scope.pool = {
		awsKeys: {
			accessKey: "accessKeyyyyyyyyyy",
			secretKey: "secretKeyyyyyyyyyy"
		},
		verifierPoolSettings: {
			userData: "",
			amiId: "id",
			minReady: 6,
			maxReady: 10
		}
	}
]
