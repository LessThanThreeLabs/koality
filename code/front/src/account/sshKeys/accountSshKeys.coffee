'use strict'

window.AccountSshKeys = ['$scope', '$window', '$location', '$routeParams', '$http', '$timeout', 'events', 'notification', ($scope, $window, $location, $routeParams, $http, $timeout, events, notification) ->
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

	handleAddedKey = (data) ->
		getKeys()

	handleRemovedKey = (data) ->
		getKeys()

	addKeyEvents = events('users', 'sshKeyAdded', false, $window.userId).setCallback(handleAddedKey).subscribe()
	removeKeyEvents = events('users', 'sshKeyRemoved', false, $window.userId).setCallback(handleRemovedKey).subscribe()
	$scope.$on '$destroy', addKeyEvents.unsubscribe
	$scope.$on '$destroy', removeKeyEvents.unsubscribe

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

	$scope.importFromGitHub = () ->
		return if $scope.waitingOnGitHubImportRequest
		$scope.waitingOnGitHubImportRequest = true

		request = $http.post "/app/users/addKeysFromGitHub"
		request.success (data, status, headers, config) =>
			if data.numKeysAdded?
				if data.numKeysAdded > 0
					notification.success "Added #{data.numKeysAdded} SSH Keys from GitHub"
				else
					notification.success "No new SSH Keys added from GitHub"
			$scope.waitingOnGitHubImportRequest = false
		request.error (data, status, headers, config) =>
			$scope.waitingOnGitHubImportRequest = false
			if data.redirectUri?
				window.location.href = data.redirectUri
			else
				notification.error data
]
