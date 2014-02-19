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
	factory('events', ['$window', '$http', 'integerConverter', 'notification', ($window, $http, integerConverter, notification) ->
		connectionId = Math.floor Math.random() * 1000000
		subscriptionCallbacks = {}

		try
			websocket = new WebSocket "wss://#{$window.location.host}/websockets/connect?csrfToken=#{$window.csrfToken}&connectionId=#{connectionId}"
			websocket.onopen = () ->
				console.log 'Websocket connected'
			websocket.onmessage = (event) ->
				message = JSON.parse(event.data)
				callback = subscriptionCallbacks[message.subscriptionId]
				if callback? then callback message.data
				else console.warn 'Unexpected event', message
			websocket.onerror = (error) ->
				console.error error
			websocket.onclose = () ->
				console.warn 'Websocket closed'
				notification.warning 'Websocket has been closed. No events will be received', 0
		catch exception
			console.error exception
			notification.warning 'Failed to open websocket. No events will be received', 0

		class EventListener
			constructor: (@connectionId, @resourceType, @eventName, @allResources, @resourceId, @subscriptionCallbackRegistry) ->
				assert.ok typeof @connectionId is 'number'
				assert.ok typeof @resourceType is 'string'
				assert.ok typeof @eventName is 'string'
				assert.ok typeof @allResources is 'boolean'
				assert.ok typeof @resourceId is 'number'
				assert.ok typeof subscriptionCallbackRegistry is 'object'
				@callback = null

			setCallback: (callback) =>
				assert.ok callback?
				@callback = callback
				return @

			subscribe: () =>
				assert.ok @callback?
				assert.ok not @subscriptionId?
				request = $http.post "/app/events/#{@resourceType}/#{@eventName}/subscribe",
					connectionId: @connectionId
					allResources: @allResources
					resourceId: @resourceId
				request.success (data, status, headers, config) =>
					@subscriptionId = integerConverter.toInteger data
					@subscriptionCallbackRegistry[@subscriptionId] = @callback
				request.error (data, status, headers, config) =>
					notification.error data
				return @

			unsubscribe: () =>
				@callback = null
				request = $http.delete "/app/events/#{@resourceType}/#{@eventName}/#{subscriptionId}"
				request.success (data, status, headers, config) =>
					@subscriptionId = null
					delete @subscriptionCallbackRegistry[@subscriptionId]
				request.error (data, status, headers, config) =>
					notification.error data
				return @

		return (resourceType, eventName, allResources=false, resourceId=0) ->
			return new EventListener connectionId, resourceType, eventName, allResources, resourceId, subscriptionCallbacks
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
