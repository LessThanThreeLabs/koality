'use strict'

window.AccountBasic = ['$scope', '$window', '$http', 'events', 'notification', ($scope, $window, $http, events, notification) ->
	$scope.makingRequest = false
	$scope.account =
		email: null
		firstName: null
		oldFirstName: null
		lastName: null
		oldLastName: null
		infoChanged: false

	getUser = () ->
		request = $http.get("/app/users/" + $window.userId)
		request.success (data, status, headers, config) =>
			processUserInformation data
		request.error (data, status, headers, config) =>
			notification.error data

	processUserInformation = (userInformation) ->
		$scope.account.email = userInformation.email
		$scope.account.oldFirstName = userInformation.firstName
		$scope.account.oldLastName = userInformation.lastName

		if not $scope.account.firstName?
			$scope.account.firstName = userInformation.firstName
			
		if not $scope.account.lastName?
			$scope.account.lastName = userInformation.lastName

	# handleNameUpdated = (data) ->
	# 	return if data.resourceId isnt initialState.user.id
	# 	processName data

	getUser()

	# nameUpdatedEvents = events('users', 'user name updated', initialState.user.id).setCallback(handleNameUpdated).subscribe()
	# $scope.$on '$destroy', nameUpdatedEvents.unsubscribe

	$scope.submit = () ->
		return if $scope.makingRequest
		$scope.makingRequest = true

		# in case they change in the UI while waiting for request to come back
		firstName = $scope.account.firstName
		lastName = $scope.account.lastName

		request = $http.post("/app/users/name", $scope.account)
		request.success (data, status, headers, config) =>
			$scope.makingRequest = false
			notification.success 'Updated account information'
			$scope.account.oldFirstName = firstName
			$scope.account.oldLastName = lastName
		request.error (data, status, headers, config) =>
			$scope.makingRequest = false
			notification.error data
]
