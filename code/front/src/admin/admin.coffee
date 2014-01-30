'use strict'

window.Admin = ['$scope', '$location', '$routeParams', 'rpc', ($scope, $location, $routeParams, rpc) ->
	$scope.view = $routeParams.view ? 'users'
	$scope.gitHubEnterpriseAllowed = true

	getCloudProvider = () ->
		rpc 'systemSettings', 'read', 'getCloudProvider', null, (error, cloudProvider) ->
			$scope.cloudProvider = cloudProvider

	getCloudProvider()

	$scope.selectView = (view) ->
		$scope.view = view

	$scope.$watch 'view', () ->
		$location.search 'view', $scope.view
]
