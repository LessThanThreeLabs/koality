'use strict'

window.CreateAccount = ['$scope', '$window', '$location', '$routeParams', '$timeout', 'notification', ($scope, $window, $location, $routeParams, $timeout, notification) ->
	$scope.createAccountType = if $window.googleLoginAllowed then 'google' else 'default'
	$scope.account = {}
	$scope.makingRequest = false
	$scope.showVerifyEmailSent = false

	if $routeParams.googleCreateAccountError
		googleCreateAccountError = $routeParams.googleCreateAccountError
		$location.search 'googleCreateAccountError', null
		$timeout (() -> notification.error googleCreateAccountError), 100

	redirectToHome = () ->
		# this will force a refresh, rather than do html5 pushstate
		window.location.href = '/'
	
	$scope.createAccount = () ->
		return if $scope.makingRequest
		$scope.makingRequest = true

		request = $http.post("/app/accounts/create", $scope.account)
		request.success (data, status, headers, config) =>
			$scope.makingRequest = false
			$scope.showVerifyEmailSent = true
		request.error (data, status, headers, config) =>
			$scope.makingRequest = false
			notification.error data

	$scope.googleCreateAccount = () ->
		return if $scope.makingRequest
		$scope.makingRequest = true

		request = $http.post("/app/accounts/gitHub/create", $scope.account)
		request.success (data, status, headers, config) =>
			console.log 'data'
			# window.location.href = redirectUri
			$scope.makingRequest = false
		request.error (data, status, headers, config) =>
			console.error data
			$scope.makingRequest = false
			notification.error data
]
