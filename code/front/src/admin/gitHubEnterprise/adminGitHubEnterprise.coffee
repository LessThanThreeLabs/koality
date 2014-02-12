'use strict'

window.AdminGitHubEnterprise = ['$scope', '$http', 'events', 'notification', ($scope, $http, events, notification) ->
	$scope.settings = {}
	$scope.makingRequest = false

	getGitHubEnterpriseSettings = () ->
		request = $http.get "/app/settings/gitHubEnterprise"
		request.success (data, status, headers, config) =>
			$scope.settings.baseUri = data.baseUri
			$scope.settings.oAuthClientId = data.oAuthClientId
			$scope.settings.oAuthClientSecret = data.oAuthClientSecret
			$scope.settings.enabled = if data? and Object.keys(data).length isnt 0 then 'yes' else 'no'
		request.error (data, status, headers, config) =>
			notification.error data
			$scope.settings.enabled = 'no'

	# handleSettigsUpdated = (data) ->
	# 	$scope.settings = data
	# 	updateEnabledRadio()

	getGitHubEnterpriseSettings()

	# settingsUpdatedEvents = events('systemSettings', 'github enterprise settings updated', null).setCallback(handleSettigsUpdated).subscribe()
	# $scope.$on '$destroy', settingsUpdatedEvents.unsubscribe

	$scope.submit = () ->
		return if $scope.makingRequest
		$scope.makingRequest = true

		request = null
		if $scope.settings.enabled is 'yes'
			requestParams = 
				baseUri: $scope.settings?.baseUri
				oAuthClientId: $scope.settings?.oAuthClientId
				oAuthClientSecret: $scope.settings?.oAuthClientSecret
			request = $http.put "/app/settings/gitHubEnterprise", requestParams
		else
			request = $http.delete "/app/settings/gitHubEnterprise"

		request.success (data, status, headers, config) =>
			$scope.makingRequest = false
			notification.success 'Successfully updated GitHub Enterprise settings'
		request.error (data, status, headers, config) =>
			$scope.makingRequest = false
			notification.error data
]
