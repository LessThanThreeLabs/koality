'use strict'

window.AdminPools = ['$scope', '$http', 'events', 'notification', ($scope, $http, events, notification) ->
	$scope.allowedAmis = null
	$scope.allowedInstanceSizes = null
	$scope.allowedSecurityGroups = null
	$scope.pool = {
		id: 1, # TODO(dhuang) make this not hardcoded
		name: null,
		awsKeys: {
			accessKey: null,
			secretKey: null
		},
		username: null,
		baseAmi: null,
		securityGroupId: null,
		vpcSubnetId: null
		instanceType: null,
		numReadyInstances: null,
		maxRunningInstances: null,
		rootDriveSize: null,
		userData: null
	}
	$scope.oldAwsKeys = angular.copy $scope.pool.awsKeys
	$scope.needAWSCredentials = "You must enter your AWS credentials before filling out this field."
	$scope.awsCredentialState = "Unset"
	$scope.updateAwsCredentials = () ->
		if angular.equals($scope.pool.awsKeys, $scope.oldAwsKeys) || $scope.pool.awsKeys.accessKey.length is 0 || $scope.pool.awsKeys.secretKey.length is 0
			return
		newAwsKeys = $scope.pool.awsKeys
		$scope.awsCredentialState = "Updating"
		$scope.makingRequest = true
		request = $http.get "/app/pools/getAwsSettings", {
			params: newAwsKeys
		}
		request.success (data, status, headers, config) ->
			if newAwsKeys isnt $scope.pool.awsKeys
				return
			$scope.allowedAmis = data.amis
			$scope.allowedInstanceSizes = data.instanceTypes
			$scope.allowedSecurityGroups = data.securityGroups
			$scope.awsCredentialState = "Set"
		request.error (data, status, headers, config) ->
			if newAwsKeys isnt $scope.pool.awsKeys
				return
			notification.error data
			$scope.awsCredentialState = "Unset"
		request.finally () ->
			$scope.makingRequest = false
		$scope.oldAwsKeys = angular.copy $scope.pool.awsKeys
	$scope.awsCredentialsSet = () ->
		$scope.awsCredentialState is "Set"
	$scope.awsCredentialsNotSet = () ->
		$scope.awsCredentialState isnt "Set"
	$scope.loadPool = () ->
		$scope.makingRequest = true
		request = $http.get "/app/pools/#{$scope.pool.id}"
		request.success (data, status, headers, config) ->
			$scope.pool.name = data.name
			$scope.pool.awsKeys = {
				accessKey: data.accessKey,
				secretKey: data.secretKey
			}
			$scope.pool.username = data.username
			$scope.pool.baseAmi = data.baseAmiId
			$scope.pool.securityGroupId = data.securityGroupId
			$scope.pool.vpcSubnetId = data.vpcSubnetId
			$scope.pool.instanceType = data.instanceType
			$scope.pool.numReadyInstances = data.numReadyInstances
			$scope.pool.maxRunningInstances = data.numMaxInstances
			$scope.pool.rootDriveSize = data.rootDriveSize
			$scope.pool.userData = data.userData
		request.error (data, status, headers, config) ->
			notification.error data
		request.finally () ->
			$scope.makingRequest = false
			$scope.updateAwsCredentials()
	$scope.submit = () ->
		$scope.makingRequest = true
		baseAmi = $scope.pool.baseAmi
		afterIdIndex = baseAmi.indexOf " ("
		if afterIdIndex >= 0
			baseAmiId = baseAmi.substr 0, afterIdIndex
		request = $http.put "/app/pools/#{$scope.pool.id}", {
			name: $scope.pool.name,
			accessKey: $scope.pool.awsKeys.accessKey,
			secretKey: $scope.pool.awsKeys.secretKey,
			username: $scope.pool.username,
			baseAmiId: baseAmiId,
			securityGroupId: $scope.pool.secretKey,
			vpcSubnetId: $scope.pool.secretKey,
			instanceType: $scope.pool.instanceType,
			numReadyInstances: $scope.pool.numReadyInstances,
			maxRunningInstances: $scope.pool.maxRunningInstances,
			rootDriveSize: $scope.pool.rootDriveSize,
			userData: $scope.pool.userData
		}
		request.success (data, status, headers, config) ->
			notification.success "Successfully updated pool"
		request.error (data, status, headers, config) ->
			notification.error data
		request.finally () ->
			$scope.makingRequest = false
	$scope.updateAwsCredentialsIfEnter = (event) ->
		if event.keyCode is 13 # enter
			$scope.updateAwsCredentials()
			event.preventDefault()
	$scope.loadPool()
]
