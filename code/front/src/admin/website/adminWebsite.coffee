'use strict'

window.AdminWebsite = ['$scope', '$http', 'events', 'notification', ($scope, $http, events, notification) ->
	$scope.domain = {}

	getDomainName = () ->
		request = $http.get "/app/settings/domainName"
		request.success (data, status, headers, config) =>
			$scope.domain.domainName = data
		request.error (data, status, headers, config) =>
			notification.error data

	# handleDomainNameUpdated = (data) ->
	# 	$scope.domain.domainName = data

	getDomainName()

	# domainNameUpdatedEvents = events('systemSettings', 'domain name updated', null).setCallback(handleDomainNameUpdated).subscribe()
	# $scope.$on '$destroy', domainNameUpdatedEvents.unsubscribe

	$scope.submit = () ->
		request = $http.put "/app/settings/domainName", $scope.domain.domainName
		request.success (data, status, headers, config) =>
			notification.success 'Updated website domain'
		request.error (data, status, headers, config) =>
			notification.error data
]
