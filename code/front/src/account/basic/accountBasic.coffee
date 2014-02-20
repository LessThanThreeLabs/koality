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
		request = $http.get "/app/users/" + $window.userId
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

	handleNameUpdated = (data) ->
		if $scope.account.oldFirstName is $scope.account.firstName
			$scope.account.oldFirstName = $scope.account.firstName = data.firstName
		if $scope.account.oldLastName is $scope.account.lastName
			$scope.account.oldLastName = $scope.account.lastName = data.lastName

	getUser()

	nameUpdatedEvents = events('users', 'name', false, $window.userId).setCallback(handleNameUpdated).subscribe()
	$scope.$on '$destroy', nameUpdatedEvents.unsubscribe

	$scope.submit = () ->
		return if $scope.makingRequest
		$scope.makingRequest = true

		# In case they change in the UI while waiting for request to come back
		firstName = $scope.account.firstName
		lastName = $scope.account.lastName

		request = $http.post("/app/users/name", $scope.account)
		request.success (data, status, headers, config) =>
			$scope.makingRequest = false
			$scope.account.oldFirstName = firstName
			$scope.account.oldLastName = lastName
			notification.success 'Updated account information'
		request.error (data, status, headers, config) =>
			$scope.makingRequest = false
			notification.error data
]
