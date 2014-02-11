'use strict'

window.AdminNotifications = ['$scope', '$http', 'events', 'notification', ($scope, $http, events, notification) ->
	$scope.hipChat = {}
	$scope.makingRequest = false

	getHipChatSettings = () ->
		request = $http.get "/app/settings/hipChat"
		request.success (data, status, headers, config) =>
			console.log data
			$scope.hipChat.authenticationToken = data.authenticationToken
			$scope.hipChat.rooms = data.rooms.join ' '
			$scope.hipChat.notifyOn = data.notifyOn
			$scope.hipChat.enabled = if data? and Object.keys(data).length isnt 0 then 'yes' else 'no'
		request.error (data, status, headers, config) =>
			notification.error data
			$scope.hipChat.enabled = 'no'

	# handleSettigsUpdated = (data) ->
	# 	$scope.settings = data
	# 	$scope.settings?.hipChat?.rooms = $scope.settings?.hipChat?.rooms?.join ' '
	# 	updateHipChatEnabledRadio()

	getHipChatSettings()

	# settingsUpdatedEvents = events('systemSettings', 'notification settings updated', null).setCallback(handleSettigsUpdated).subscribe()
	# $scope.$on '$destroy', settingsUpdatedEvents.unsubscribe

	$scope.submit = () ->
		getHipChatRooms = () ->
			return [] if not $scope.hipChat?.rooms?

			hipChatRooms = []
			if $scope.hipChat.rooms isnt ''
				hipChatRooms = $scope.hipChat.rooms.split(/[,; ]/)
				hipChatRooms = hipChatRooms.filter (room) -> return room isnt ''
			return hipChatRooms

		return if $scope.makingRequest
		$scope.makingRequest = true

		request = null
		if $scope.hipChat.enabled is 'yes'
			requestParams = 
				authenticationToken: $scope.hipChat?.authenticationToken
				rooms: getHipChatRooms()
				notifyOn: $scope.hipChat?.notifyOn
			request = $http.put "/app/settings/hipChat", requestParams
		else
			request = $http.delete "/app/settings/hipChat"

		request.success (data, status, headers, config) =>
			$scope.makingRequest = false
			notification.success 'Successfully updated notification settings'
		request.error (data, status, headers, config) =>
			$scope.makingRequest = false
			notification.error data
]
