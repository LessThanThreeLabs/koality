'use strict'

window.Login = ['$scope', '$location', '$routeParams', '$timeout', 'initialState', 'rpc', 'cookieExtender', 'notification', ($scope, $location, $routeParams, $timeout, initialState, rpc, cookieExtender, notification) ->
	$scope.loginConfig = 
		type: initialState.userConnectionType
		defaultType: initialState.userConnectionType
	$scope.account = {}
	$scope.makingRequest = false

	if $routeParams.googleLoginError
		googleLoginError = $routeParams.googleLoginError
		$location.search 'googleLoginError', null
		$timeout (() -> notification.error googleLoginError), 100

	redirectToHome = () ->
		# this will force a refresh, rather than do html5 pushstate
		window.location.href = '/'

	$scope.login = () ->
		return if $scope.makingRequest
		$scope.makingRequest = true

		rpc 'users', 'update', 'login', $scope.account, (error, result) ->
			$scope.makingRequest = false
			
			if error?
				$scope.account.password = ''
				notification.error error
			else
				if $scope.account.rememberMe is 'yes'
					cookieExtender.extendCookie (error) ->
						console.error error if error?
						redirectToHome()
				else
					redirectToHome()

	$scope.googleLogin = () ->
		return if $scope.makingRequest
		$scope.makingRequest = true

		rpc 'users', 'read', 'getGoogleLoginRedirect', null, (error, redirectUri) ->
			$scope.makingRequest = false
			
			if error? then notification.error error
			else
				if $scope.account.rememberMe is 'yes'
					cookieExtender.extendOAuthCookie (error) ->
						console.error error if error?
						window.location.href = redirectUri
				else
					window.location.href = redirectUri
]
