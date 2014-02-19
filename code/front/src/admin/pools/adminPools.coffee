'use strict'

window.AdminPools = ['$scope', '$http', 'events', 'notification', ($scope, $http, events, notification) ->
	$scope.allowedAmis = null
	$scope.allowedInstanceSizes = null
	$scope.allowedSecurityGroups = null
	$scope.pool = {
		id: null,
		name: null,
		awsKeys: {
			accessKey: null,
			secretKey: null
		},
		verifierPoolSettings: {
			userData: null,
			amiId: null,
			minReady: null,
			maxRunning: null,
			username: null,
			instanceSize: null,
			rootDriveSize: null
		},
		instanceSettings: {
			securityGroupId: null,
			subnetId: null
		}
	}
	$scope.oldAwsKeys = angular.copy $scope.pool.awsKeys
	$scope.needAWSCredentials = "You must enter your AWS credentials before filling out this field."
	$scope.awsCredentialState = "Unset"
	$scope.updateAwsCredentials = () ->
		console.log "updateAwsCredentials() called"
		if angular.equals($scope.pool.awsKeys, $scope.oldAwsKeys) || not $scope.pool.awsKeys.accessKey || not $scope.pool.awsKeys.secretKey
			return
		newAwsKeys = $scope.pool.awsKeys
		$scope.awsCredentialState = "Updating"
		$scope.makingRequest = true
		request = $http.get "/app/pools/getAwsSettings", {
			params: {
				secretKey: newAwsKeys.secretKey,
				accessKey: newAwsKeys.accessKey,
			}
		}
		request.success (data, status, headers, config) ->
			if newAwsKeys isnt $scope.pool.awsKeys
				return
			console.log "success"
			$scope.allowedAmiIds = data.amis
			$scope.allowedInstanceSizes = data.instanceTypes
			$scope.allowedSecurityGroups = data.securityGroups
			$scope.awsCredentialState = "Set"
		request.error (data, status, headers, config) ->
			if newAwsKeys isnt $scope.pool.awsKeys
				return
			console.log "error"
			notification.error data
			$scope.awsCredentialState = "Unset"
		request.finally () ->
			$scope.makingRequest = false
		$scope.oldAwsKeys = angular.copy $scope.pool.awsKeys
	$scope.awsCredentialsSet = () ->
		$scope.awsCredentialState == "Set"
	$scope.awsCredentialsNotSet = () ->
		$scope.awsCredentialState != "Set"
]
