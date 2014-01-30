'use strict'

window.AdminApi = ['$scope', 'rpc', 'events', 'notification', ($scope, rpc, events, notification) ->
	$scope.mustConfirmRegenerateKey = false
	$scope.makingRequest = false

	getApiKey = () ->
		rpc 'systemSettings', 'read', 'getAdminApiKey', null, (error, apiKey) ->
			$scope.apiKey = apiKey

	getDomainName = () ->
		rpc 'systemSettings', 'read', 'getWebsiteSettings', null, (error, websiteSettings) ->
			$scope.domainName = websiteSettings.domainName

	handleApiKeyUpdated = (data) ->
		$scope.apiKey = data

	getApiKey()
	getDomainName()

	apiKeyUpdatedEvents = events('systemSettings', 'admin api key updated', null).setCallback(handleApiKeyUpdated).subscribe()
	$scope.$on '$destroy', apiKeyUpdatedEvents.unsubscribe

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
