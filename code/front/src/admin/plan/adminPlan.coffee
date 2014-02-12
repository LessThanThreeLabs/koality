'use strict'

window.AdminPlan = ['$scope', '$http', 'notification', ($scope, $http, notification) ->
	$scope.license = {}

	getLicense = () ->
		request = $http.get "/app/settings/license"
		request.success (data, status, headers, config) =>
			$scope.license = data
		request.error (data, status, headers, config) =>
			notification.error data

	getLicense()
]
