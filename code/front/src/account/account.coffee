'use strict'

window.Account = ['$scope', '$location', '$routeParams', ($scope, $location, $routeParams) ->
	$scope.view = $routeParams.view ? 'basic'

	$scope.selectView = (view) ->
		$scope.view = view

	$scope.$watch 'view', () ->
		$location.search 'view', $scope.view
]
