'use strict'

window.AdminNotifications = ['$scope', '$http', 'events', 'notification', ($scope, $http, events, notification) ->
	$scope.smtp = {}
	$scope.hipChat = {}
	$scope.makingRequest = false

	getSmtpSettings = () ->
		request = $http.get "/app/settings/smtp"
		request.success (data, status, headers, config) =>
			$scope.smtp.hostname = data.hostname
			$scope.smtp.port = data.port
			$scope.smtp.authenticationType = data.authenticationType
			$scope.smtp.username = data.username

			if !data.authenticationType?
				$scope.smtp.authenticationType = 'plain'
				$scope.smtp.hostname = 'smtp.gmail.com'
				$scope.smtp.port = 587

		request.error (data, status, headers, config) =>
			notification.error data

	getHipChatSettings = () ->
		request = $http.get "/app/settings/hipChat"
		request.success (data, status, headers, config) =>
			$scope.hipChat.authenticationToken = data.authenticationToken
			$scope.hipChat.rooms = if data.rooms? then data.rooms.join ', ' else ''
			$scope.hipChat.notifyOn = data.notifyOn
			$scope.hipChat.enabled = if data? and Object.keys(data).length isnt 0 then 'yes' else 'no'
		request.error (data, status, headers, config) =>
			notification.error data
			$scope.hipChat.enabled = 'no'

	# handleSettigsUpdated = (data) ->
	# 	$scope.settings = data
	# 	$scope.settings?.hipChat?.rooms = $scope.settings?.hipChat?.rooms?.join ' '
	# 	updateHipChatEnabledRadio()

	getSmtpSettings()
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

		updateSmtpSettings = (callback) ->
			request = $http.put "/app/settings/smtp", $scope.smtp
			request.success (data, status, headers, config) =>
				callback null
			request.error (data, status, headers, config) =>
				callback data

		updateHipChatSettings = (callback) ->
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
				callback null
			request.error (data, status, headers, config) =>
				callback data

		await
			updateSmtpSettings defer smtpError
			updateHipChatSettings defer hipChatError

		$scope.makingRequest = false
		if smtpError?
			notification.error smtpError
		else if hipChatError?
			notification.error hipChatError
		else
			notification.success 'Successfully updated notification settings'
]
