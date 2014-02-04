'use strict'

window.AccountBasic = ['$scope', '$http', 'initialState', 'events', 'notification', ($scope, $http, initialState, events, notification) ->
	# $scope.makingRequest = false
	# $scope.account =
	# 	email: initialState.user.email
	# 	firstName: null
	# 	oldFirstName: null
	# 	lastName: null
	# 	oldLastName: null
	# 	infoChanged: false

	# getName = () ->
	# 	rpc 'users', 'read', 'getBasicInformation', null, (error, basicInformation) ->
	# 		if error? then notification.error error
	# 		else
	# 			$scope.account.email = basicInformation.email
	# 			processName basicInformation

	# processName = (nameInformation) ->
	# 	$scope.account.oldFirstName = nameInformation.firstName
	# 	$scope.account.oldLastName = nameInformation.lastName

	# 	if not $scope.account.firstName?
	# 		$scope.account.firstName = nameInformation.firstName
			
	# 	if not $scope.account.lastName?
	# 		$scope.account.lastName = nameInformation.lastName

	# handleNameUpdated = (data) ->
	# 	return if data.resourceId isnt initialState.user.id
	# 	processName data

	# getName()

	# nameUpdatedEvents = events('users', 'user name updated', initialState.user.id).setCallback(handleNameUpdated).subscribe()
	# $scope.$on '$destroy', nameUpdatedEvents.unsubscribe

	# $scope.submit = () ->
	# 	return if $scope.makingRequest
	# 	$scope.makingRequest = true

	# 	# in case they change in the UI while waiting for request to come back
	# 	firstName = $scope.account.firstName
	# 	lastName = $scope.account.lastName

	# 	rpc 'users', 'update', 'changeBasicInformation', $scope.account, (error) ->
	# 		$scope.makingRequest = false
	# 		if error? then notification.error 'Unable to update account information'
	# 		else
	# 			notification.success 'Updated account information'
	# 			$scope.account.oldFirstName = firstName
	# 			$scope.account.oldLastName = lastName
]
