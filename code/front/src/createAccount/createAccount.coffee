'use strict'

window.CreateAccount = ['$scope', '$location', '$routeParams', '$timeout', 'initialState', 'rpc', 'cookieExtender', 'notification', ($scope, $location, $routeParams, $timeout, initialState, rpc, cookieExtender, notification) ->
	$scope.createAccountType = initialState.userConnectionType
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

		rpc 'users', 'create', 'createUser', $scope.account, (error, result) ->
			$scope.makingRequest = false
			if error then notification.error error
			else
				$scope.showVerifyEmailSent = true

	$scope.googleCreateAccount = () ->
		return if $scope.makingRequest
		$scope.makingRequest = true

		rpc 'users', 'read', 'getGoogleCreateAccountRedirect', null, (error, redirectUri) ->
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
