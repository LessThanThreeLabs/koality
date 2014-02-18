'use strict'

angular.module('koality.service', []).
	factory('localStorage', ['$window', ($window) ->
		return window.localStorage
	]).
	factory('integerConverter', [() ->
		return toInteger: (integerAsString) ->
			if typeof integerAsString is 'number'
				if integerAsString isnt Math.floor(integerAsString) then return null
				else return integerAsString

			return null if typeof integerAsString isnt 'string'
			return null if integerAsString.indexOf('.') isnt -1

			integer = parseInt integerAsString
			return null if isNaN integer
			return integer
	]).
	factory('ansiParser', ['$window', ($window) ->
		return parse: (text) ->
			return '<span class="ansi">' + $window.ansiParse(text) + '</span>'
	]).
	factory('xUnitParser', ['$window', ($window) ->
		return getTestCases: (xunitOutputs) ->
			testCases = []
			for xunitOutput in xunitOutputs
				testCases = testCases.concat $window.xUnitParse xunitOutput
			return testCases
	]).
	factory('stringHasher', [() ->
		return hash: (text) =>
			return null if typeof text isnt 'string'

			hash = 0
			hash += text.charCodeAt index for index in [0...text.length]
			return hash
	]).
	factory('events', ['$window', '$http', 'notification', ($window, $http, notification) ->
		connectionId = Math.floor Math.random() * 1000000

		try
			websocket = new WebSocket "wss://#{$window.location.host}/websockets/connect?csrfToken=#{$window.csrfToken}&connectionId=#{connectionId}"
			websocket.onopen = () ->
				console.log 'Websocket connected'

				requestParams = 
					connectionId: connectionId
					allResources: true
					resourceId: 0
				request = $http.post '/app/events/users/created/subscribe', requestParams
				request.success (data, status, headers, config) =>
					console.log data
				request.error (data, status, headers, config) =>
					notification.error data

			websocket.onmessage = (event) ->
				console.log event
				console.log event.data
			websocket.onerror = (error) ->
				console.error error
			websocket.onclose = () ->
				console.log 'Websocket closed'
				notification.warning 'Websocket has been closed. No events will be received', 0
		catch exception
			console.error exception
			notification.warning 'Failed to open websocket. No events will be received', 0
	]).
	factory('notification', ['$compile', '$rootScope', '$document', '$timeout', ($compile, $rootScope, $document, $timeout) ->
		container = $document.find '#notificationsContainer'
		
		add = (type, text, durationInSeconds) ->
			assert.ok typeof durationInSeconds is 'number' and durationInSeconds >= 0

			if typeof text is 'object'
				text = (value for key, value of text).join ', '

			notification = "<notification type='#{type}' duration-in-seconds=#{durationInSeconds} unselectable>#{text}</notification>"

			scope = $rootScope.$new(true)
			notification = $compile(notification)(scope)
			$timeout (() -> scope.$apply () -> container.append notification)

		toReturn =
			success: (text, durationInSeconds=8) -> add 'success', text, durationInSeconds
			warning: (text, durationInSeconds=8) -> add 'warning', text, durationInSeconds
			error: (text, durationInSeconds=8) -> add 'error', text, durationInSeconds
		return toReturn
	])
