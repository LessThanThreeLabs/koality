'use strict'

window.AccountSshKeys = ['$scope', '$location', '$routeParams', '$http', '$timeout', 'events', 'notification', ($scope, $location, $routeParams, $http, $timeout, events, notification) ->
	$scope.orderByPredicate = 'alias'
	$scope.orderByReverse = false

	$scope.currentlyOpenDrawer = null
	$scope.waitingOnGitHubImportRequest = false

	$scope.addKey =
		makingRequest: false
		drawerOpen: false

	if $routeParams.importGitHubKeys
		$location.search 'importGitHubKeys', null
		$timeout (() -> $scope.importFromGitHub()), 100

	getKeys = () ->
		request = $http.get "/app/users/keys"
		request.success (data, status, headers, config) =>
			$scope.keys = data
		request.error (data, status, headers, config) =>
			notification.error data

	# handleAddedKey = (data) ->
	# 	return if data.resourceId isnt initialState.user.id
	# 	$scope.keys.push data

	# handleRemovedKey = (data) ->
	# 	return if data.resourceId isnt initialState.user.id
	# 	keyToRemoveIndex = (index for key, index in $scope.keys when key.id is data.id)[0]
	# 	$scope.keys.splice keyToRemoveIndex, 1 if keyToRemoveIndex?

	# addKeyEvents = events('users', 'ssh pubkey added', initialState.user.id).setCallback(handleAddedKey).subscribe()
	# removeKeyEvents = events('users', 'ssh pubkey removed', initialState.user.id).setCallback(handleRemovedKey).subscribe()
	# $scope.$on '$destroy', addKeyEvents.unsubscribe
	# $scope.$on '$destroy', removeKeyEvents.unsubscribe

	getKeys()

	$scope.toggleDrawer = (drawerName) ->
		if $scope.currentlyOpenDrawer is drawerName
			$scope.currentlyOpenDrawer = null
		else
			$scope.currentlyOpenDrawer = drawerName

	$scope.removeKey = (key) ->
		request = $http.post "/app/users/removeKey", key
		request.success (data, status, headers, config) =>
			notification.success 'SSH Key has been removed'
		request.error (data, status, headers, config) =>
			notification.error data

	$scope.submitKey = () ->
		return if $scope.addKey.makingRequest
		$scope.addKey.makingRequest = true

		request = $http.post "/app/users/addKey", $scope.addKey
		request.success (data, status, headers, config) =>
			$scope.addKey.makingRequest = false
			notification.success 'Added SSH key: ' + $scope.addKey.name
			$scope.clearAddKey()
		request.error (data, status, headers, config) =>
			$scope.addKey.makingRequest = false
			notification.error data

	$scope.clearAddKey = () ->
		$scope.addKey.name = ''
		$scope.addKey.publicKey = ''
		$scope.currentlyOpenDrawer = null

	# $scope.importFromGitHub = () ->
	# 	return if $scope.waitingOnGitHubImportRequest
	# 	$scope.waitingOnGitHubImportRequest = true

	# 	rpc 'users', 'update', 'addGitHubSshKeys', null, (error) ->
	# 		$scope.waitingOnGitHubImportRequest = false

	# 		if error?
	# 			if error.redirect? then window.location.href = error.redirect
	# 			else notification.error error
	# 		else
	# 			notification.success 'Added GitHub SSH Keys'
]
