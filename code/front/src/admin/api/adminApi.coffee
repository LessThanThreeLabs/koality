'use strict'

window.AdminApi = ['$scope', '$http', 'events', 'notification', ($scope, $http, events, notification) ->
	$scope.mustConfirmRegenerateKey = false
	$scope.makingRequest = false

	getApiKey = () ->
		request = $http.get "/app/settings/apiKey"
		request.success (data, status, headers, config) =>
			$scope.apiKey = data
		request.error (data, status, headers, config) =>
			notification.error data

	getDomainName = () ->
		request = $http.get "/app/settings/domainName"
		request.success (data, status, headers, config) =>
			$scope.domainName = data
		request.error (data, status, headers, config) =>
			notification.error data

	# handleApiKeyUpdated = (data) ->
	# 	$scope.apiKey = data

	getApiKey()
	getDomainName()

	# apiKeyUpdatedEvents = events('systemSettings', 'admin api key updated', null).setCallback(handleApiKeyUpdated).subscribe()
	# $scope.$on '$destroy', apiKeyUpdatedEvents.unsubscribe

	$scope.regenerateKey = () ->
		$scope.mustConfirmRegenerateKey = true

	$scope.confirmRegenerateKey = () ->
		return if $scope.makingRequest
		$scope.makingRequest = true

		rpc 'systemSettings', 'update', 'regenerateApiKey', null, (error, apiKey) ->
			$scope.makingRequest = false
			$scope.apiKey = apiKey
			$scope.mustConfirmRegenerateKey = false
			notification.success 'Successfully updated API key'

	$scope.cancelRegenerateKey = () ->
		$scope.mustConfirmRegenerateKey = false
]
