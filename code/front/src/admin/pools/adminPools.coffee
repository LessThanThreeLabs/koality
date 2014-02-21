'use strict'

window.AdminPools = ['$scope', '$http', '$location', 'events', 'notification', ($scope, $http, $location, events, notification) ->
	$scope.needAWSCredentials = "You must enter your AWS credentials before filling out this field."
	$scope.returning = false

	getPoolFromServerResponse = (resp) ->
		{
			id: resp.id,
			name: resp.name,
			awsKeys: {
				accessKey: resp.accessKey,
				secretKey: resp.secretKey
			},
			username: resp.username,
			baseAmi: resp.baseAmiId,
			securityGroupId: resp.securityGroupId,
			vpcSubnetId: resp.vpcSubnetId,
			instanceType: resp.instanceType,
			numReadyInstances: resp.numReadyInstances,
			maxRunningInstances: resp.numMaxInstances,
			rootDriveSize: resp.rootDriveSize,
			userData: resp.userData
		}

	$scope.resetPoolWithId = (id) ->
		$scope.pool = {
			id: id,
			name: null,
			awsKeys: {
				accessKey: "",
				secretKey: ""
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
		request = $http.get "/app/pools/#{parseInt($location.search().poolId)}"
		request.success (data, status, headers, config) ->
			$scope.pool = getPoolFromServerResponse data
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
			baseAmi = baseAmi.substr 0, afterIdIndex
		params = {
			name: $scope.pool.name,
			accessKey: $scope.pool.awsKeys.accessKey,
			secretKey: $scope.pool.awsKeys.secretKey,
			username: $scope.pool.username,
			baseAmiId: baseAmi,
			securityGroupId: $scope.pool.securityGroupId,
			vpcSubnetId: $scope.pool.vpcSubnetId,
			instanceType: $scope.pool.instanceType,
			numReadyInstances: $scope.pool.numReadyInstances,
			maxRunningInstances: $scope.pool.maxRunningInstances,
			rootDriveSize: $scope.pool.rootDriveSize,
			userData: $scope.pool.userData
		}
		switch $scope.action()
			when "update"
				request = $http.put "/app/pools/#{$scope.pool.id}", params
				request.success (data, status, headers, config) ->
					notification.success "Successfully updated pool"
				request.error (data, status, headers, config) ->
					notification.error data
				request.finally () ->
					$scope.makingRequest = false
			when "create"
				request = $http.post "/app/pools/", params
				request.success (data, status, headers, config) ->
					notification.success "Successfully created pool"
					$scope.navigateList()
				request.error (data, status, headers, config) ->
					notification.error data
				request.finally () ->
					$scope.makingRequest = false

	$scope.updateAwsCredentialsIfEnter = (event) ->
		if event.keyCode is 13 # enter
			$scope.updateAwsCredentials()
			event.preventDefault()

	$scope.deletePool = (pool) ->
		pool.deleting = true
		if pool.deleteName isnt pool.name
			return
		request = $http.delete "/app/pools/#{pool.id}"
		request.success (data, status, headers, config) ->
			notification.success "Successfully deleted pool"
			$scope.loadPools()
		request.error (data, status, headers, config) ->
			notification.error data
		request.finally () ->
			pool.deleting = false
	$scope.action = () ->
		action = $location.search().action
		if action in ["update", "create", "list"] then action else "list"

	$scope.navigateNew = () ->
		$location.url "/admin?view=pools&action=create"
		$scope.loadNew()
	$scope.loadNew = () ->
		$scope.awsCredentialState = "Unset"
		$scope.allowedAmis = null
		$scope.allowedInstanceSizes = null
		$scope.allowedSecurityGroups = null
		$scope.resetPoolWithId null
		$scope.oldAwsKeys = angular.copy $scope.pool.awsKeys

	$scope.returnToList = (pristine) ->
		return $scope.navigateList() if pristine
		$scope.returning = true
	$scope.stopReturning = () ->
		$scope.returning = false
	$scope.navigateList = () ->
		$scope.returning = false
		$location.url "/admin?view=pools&action=list"
		$scope.loadPools()
	$scope.loadPools = () ->
		$scope.pool = {}
		$scope.makingRequest = true
		request = $http.get "/app/pools/"
		request.success (data, status, headers, config) ->
			$scope.pools = data.map getPoolFromServerResponse
		request.error (data, status, headers, config) ->
			notification.error data
		request.finally () ->
			$scope.makingRequest = false

	$scope.navigateEdit = (pool) ->
		$location.url "/admin?view=pools&action=update&poolId=#{pool.id}"
		$scope.loadEdit()
	$scope.loadEdit = () ->
		$scope.awsCredentialState = "Unset"
		$scope.allowedAmis = null
		$scope.allowedInstanceSizes = null
		$scope.allowedSecurityGroups = null
		$scope.resetPoolWithId parseInt($location.search().poolId)
		$scope.oldAwsKeys = angular.copy $scope.pool.awsKeys
		$scope.loadPool()

	switch $scope.action()
		when "update"
			$scope.loadEdit()
		when "create"
			$scope.loadNew()
		when "list"
			$scope.loadPools()
]
