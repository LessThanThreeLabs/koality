'use strict'

window.CreateAccount = ['$scope', '$window', '$location', '$routeParams', '$http', '$timeout', 'notification', ($scope, $window, $location, $routeParams, $http, $timeout, notification) ->
	$scope.createAccountType = if $window.googleAccountsAllowed then 'google' else 'default'
	$scope.account = {}
	$scope.makingRequest = false
	$scope.showVerifyEmailSent = false

	redirectToHome = () ->
		# this will force a refresh, rather than do html5 pushstate
		window.location.href = '/'

	$scope.createAccount = () ->
		return if $scope.makingRequest
		$scope.makingRequest = true

		request = $http.post "/app/accounts/create", $scope.account
		request.success (data, status, headers, config) =>
			$scope.makingRequest = false
			$scope.showVerifyEmailSent = true
		request.error (data, status, headers, config) =>
			$scope.makingRequest = false
			notification.error data

	$scope.googleCreateAccount = () ->
		return if $scope.makingRequest
		$scope.makingRequest = true

		request = $http.get "/app/accounts/googleCreateAccountRedirect"
		request.success (data, status, headers, config) =>
			window.location.href = data
		request.error (data, status, headers, config) =>
			$scope.makingRequest = false
			notification.error data

	confirmAccount = (token) ->
		return if $scope.makingRequest
		$scope.makingRequest = true

		request = $http.post "/app/accounts/confirm", token
		request.success (data, status, headers, config) =>
			notification.success "asdf"
			redirectToHome()
		request.error (data, status, headers, config) =>
			$scope.makingRequest = false
			notification.error data

	if $routeParams.googleCreateAccountError
		googleCreateAccountError = $routeParams.googleCreateAccountError
		$location.search 'googleCreateAccountError', null
		$timeout (() -> notification.error googleCreateAccountError), 100
	else if $routeParams.token
		token = $routeParams.token
		$location.search 'token', null
		confirmAccount token
]
