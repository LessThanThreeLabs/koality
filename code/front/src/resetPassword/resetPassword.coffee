'use strict'

window.ResetPassword = ['$scope', 'rpc', 'notification', ($scope, rpc, notification) ->
	$scope.account = {}
	$scope.showSuccess = false
	
	$scope.resetPassword = () ->
		rpc 'users', 'update', 'resetPassword', $scope.account, (error) ->
			if error? then notification.error error
			else $scope.showSuccess = true
]
