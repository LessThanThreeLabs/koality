'use strict'

window.ResetPassword = ['$scope', '$http', 'notification', ($scope, $http, notification) ->
	$scope.account = {}
	$scope.showSuccess = false
	
	$scope.resetPassword = () ->
		return if $scope.makingRequest
		$scope.makingRequest = true

		request = $http.post "/app/accounts/resetPassword", $scope.account.email
		request.success (data, status, headers, config) =>
			$scope.makingRequest = false
			$scope.showSuccess = true
		request.error (data, status, headers, config) =>
			$scope.makingRequest = false
			notification.error data
]
