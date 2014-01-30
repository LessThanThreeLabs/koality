'use strict'

window.AdminNotifications = ['$scope', 'rpc', 'events', 'notification', ($scope, rpc, events, notification) ->
	$scope.makingRequest = false

	updateHipChatEnabledRadio = () ->
		$scope.settings?.hipChat?.enabled = if $scope.settings?.hipChat?.token isnt '' then 'yes' else 'no'

	getNotificationSettings = () ->
		rpc 'systemSettings', 'read', 'getNotificationSettings', null, (error, notificationSettings) ->
			$scope.settings = notificationSettings
			$scope.settings?.hipChat?.rooms = $scope.settings?.hipChat?.rooms?.join ' '
			updateHipChatEnabledRadio()

	handleSettigsUpdated = (data) ->
		$scope.settings = data
		$scope.settings?.hipChat?.rooms = $scope.settings?.hipChat?.rooms?.join ' '
		updateHipChatEnabledRadio()

	getNotificationSettings()

	settingsUpdatedEvents = events('systemSettings', 'notification settings updated', null).setCallback(handleSettigsUpdated).subscribe()
	$scope.$on '$destroy', settingsUpdatedEvents.unsubscribe

	$scope.submit = () ->
		getHipChatRooms = () ->
			return [] if not $scope.settings?.hipChat?.rooms?

			hipChatRooms = []
			if $scope.settings.hipChat.rooms isnt ''
				hipChatRooms = $scope.settings.hipChat.rooms.split(/[,; ]/)
				hipChatRooms = hipChatRooms.filter (room) -> return room isnt ''
			return hipChatRooms

		return if $scope.makingRequest
		$scope.makingRequest = true

		requestParams = 
			hipChat:
				type: if $scope.settings?.hipChat?.enabled is 'yes' then $scope.settings?.hipChat?.type else '' 
				token: if $scope.settings?.hipChat?.enabled is 'yes' then $scope.settings?.hipChat?.token else ''
				rooms: if $scope.settings?.hipChat?.enabled is 'yes' then getHipChatRooms() else []
		rpc 'systemSettings', 'update', 'setNotificationSettings', requestParams, (error) =>
			$scope.makingRequest = false
			if error? then notification.error error
			else notification.success 'Successfully updated notification settings'
]
