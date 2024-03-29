'use strict'

window.Login = ['$scope', '$window', '$location', '$routeParams', '$http', '$timeout', 'notification', ($scope, $window, $location, $routeParams, $http, $timeout, notification) ->
	$scope.loginConfig =
		type: if $window.googleAccountsAllowed then 'google' else 'default'
		defaultType: if $window.googleAccountsAllowed then 'google' else 'default'
	$scope.account = {}
	$scope.makingRequest = false

	if $routeParams.googleLoginError
		googleLoginError = $routeParams.googleLoginError
		$location.search 'googleLoginError', null
		$timeout (() -> notification.error googleLoginError), 100

	if $routeParams.googleCreateAccountError
		googleCreateAccountError = $routeParams.googleCreateAccountError
		$location.search 'googleCreateAccountError', null
		$timeout (() -> notification.error googleCreateAccountError), 100

	$scope.login = () ->
		return if $scope.makingRequest
		$scope.makingRequest = true

		request = $http.post "/app/accounts/login", $scope.account
		request.success (data, status, headers, config) =>
			# this will force a refresh, rather than do html5 pushstate
			window.location.href = '/'
		request.error (data, status, headers, config) =>
			$scope.account.password = ''
			notification.error data
			$scope.makingRequest = false

	$scope.googleLogin = () ->
		return if $scope.makingRequest
		$scope.makingRequest = true

		request = $http.get "/app/accounts/googleLoginRedirect?rememberMe=#{$scope.account.rememberMe}"
		request.success (data, status, headers, config) =>
			window.location.href = data
		request.error (data, status, headers, config) =>
			notification.error data
			$scope.makingRequest = false

	$scope.googleCreateAccount = () ->
		return if $scope.makingRequest
		$scope.makingRequest = true

		request = $http.get "/app/accounts/googleCreateAccountRedirect?rememberMe=#{$scope.account.rememberMe}"
		request.success (data, status, headers, config) =>
			window.location.href = data
		request.error (data, status, headers, config) =>
			notification.error data
			$scope.makingRequest = false
]
