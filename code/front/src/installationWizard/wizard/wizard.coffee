'use strict'

window.Wizard = ['$scope', '$location', '$http', 'rpc', 'notification', ($scope, $location, $http, rpc, notification) ->
	$scope.stage = 'licenseKey'

	$scope.license = {}
	$scope.admin = {}
	$scope.website = domainName: $location.host()
	$scope.aws = {}

	$scope.waitingOnRequest = false

	$scope.completeLicenseKey = () ->
		return if $scope.waitingOnRequest
		$scope.waitingOnRequest = true

		rpc 'systemSettings', 'update', 'validateLicenseKey', $scope.license, (error, isValid) ->
			$scope.waitingOnRequest = false
			if error? then notification.error 'Error while trying to validate license key'
			else if not isValid then notification.error 'Invalid license key. Either the key does not exist or is already in use'
			else $scope.stage = 'admin'

	$scope.completeAdminInformation = () ->
		return if $scope.waitingOnRequest
		$scope.waitingOnRequest = true

		rpc 'users', 'create', 'validateInitialAdminUser', $scope.admin, (error) ->
			$scope.waitingOnRequest = false
			if error? then notification.error error
			else $scope.stage = 'verifyAdmin'

	$scope.completeVerifyAdmin = () ->
		return if $scope.waitingOnRequest
		$scope.waitingOnRequest = true

		rpc 'users', 'create', 'validateInitialAdminToken', $scope.admin, (error) ->
			$scope.waitingOnRequest = false
			if error? then notification.error error
			else $scope.stage = 'website'

	$scope.completeWebsiteInformation = () ->
		return if $scope.waitingOnRequest
		$scope.waitingOnRequest = true

		rpc 'systemSettings', 'update', 'validateInitialWebsiteSettings', $scope.website, (error) ->
			$scope.waitingOnRequest = false
			if error? then notification.error error
			else $scope.stage = 'aws'

	$scope.completeAwsInformation = () ->
		return if $scope.waitingOnRequest
		$scope.waitingOnRequest = true

		requestParams =
			license: $scope.license
			admin: $scope.admin
			website: $scope.website
			aws: $scope.aws

		rpc 'systemSettings', 'update', 'setDeploymentInitialized', requestParams, (error) ->
			$scope.waitingOnRequest = false
			if error? then notification.error error
			else
				$http.post('/turnOffInstallationWizard').error (data, status, headers, config) ->
					notification.error 'Fatal: unable to start Koality service!'
				$scope.stage = 'complete'

	$scope.goToCreateRepository = () ->
		window.location.href = '/admin?view=repositories'

	$scope.goToKoality = () ->
		window.location.href = '/'
]
