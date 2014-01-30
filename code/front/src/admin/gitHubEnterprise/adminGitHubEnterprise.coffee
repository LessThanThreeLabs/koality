'use strict'

window.AdminGitHubEnterprise = ['$scope', 'rpc', 'events', 'notification', ($scope, rpc, events, notification) ->
	$scope.makingRequest = false

	updateEnabledRadio = () ->
		$scope.settings.enabled = if $scope.settings.uri isnt '' then 'yes' else 'no'

	getGitHubEnterpriseSettings = () ->
		rpc 'systemSettings', 'read', 'getGitHubEnterpriseSettings', null, (error, gitHubEnterpriseSettings) ->
			$scope.settings = gitHubEnterpriseSettings
			updateEnabledRadio()

	handleSettigsUpdated = (data) ->
		$scope.settings = data
		updateEnabledRadio()

	getGitHubEnterpriseSettings()

	settingsUpdatedEvents = events('systemSettings', 'github enterprise settings updated', null).setCallback(handleSettigsUpdated).subscribe()
	$scope.$on '$destroy', settingsUpdatedEvents.unsubscribe

	$scope.submit = () ->
		return if $scope.makingRequest
		$scope.makingRequest = true

		requestParams = 
			uri: if $scope.settings.enabled is 'yes' then $scope.settings.uri else ''
			clientId: if $scope.settings.enabled is 'yes' then $scope.settings.clientId else ''
			clientSecret: if $scope.settings.enabled is 'yes' then $scope.settings.clientSecret else ''

		rpc 'systemSettings', 'update', 'setGitHubEnterpriseSettings', requestParams, (error) =>
			$scope.makingRequest = false
			if error? then notification.error error
			else notification.success 'Updated GitHub Enterprise settings'
]
