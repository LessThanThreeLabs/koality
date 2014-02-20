'use strict'

window.AccountPassword = ['$scope', '$http', 'notification', ($scope, $http, notification) ->
	$scope.makingRequest = false
	$scope.password = {}

	$scope.submit = () ->
		if $scope.password.newPassword isnt $scope.password.confirmPassword
			notification.error 'Invalid password confirmation. Please check that you correctly confirmed your new password'
		else if $scope.password.oldPassword is $scope.password.newPassword
			notification.error 'New password must be different from old password'
		else
			return if $scope.makingRequest
			$scope.makingRequest = true

			request = $http.post "/app/users/password", $scope.password
			request.success (data, status, headers, config) =>
				$scope.makingRequest = false
				$scope.password = {}
				notification.success 'Updated account password'
			request.error (data, status, headers, config) =>
				$scope.makingRequest = false
				notification.error data
]
