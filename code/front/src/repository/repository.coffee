'use strict'

window.Repository = ['$scope', '$location', '$routeParams', 'rpc', 'currentRepository', 'currentChange', 'currentStage', ($scope, $location, $routeParams, rpc, currentRepository, currentChange, currentStage) ->
	$scope.selectedRepository = currentRepository
	$scope.selectedChange = currentChange
	$scope.selectedStage = currentStage

	syncToRouteParams = () ->
		$scope.selectedRepository.setId $routeParams.repositoryId
		$scope.selectedRepository.retrieveInformation $routeParams.repositoryId

		if $routeParams.change?
			$scope.selectedChange.setId $routeParams.repositoryId, $routeParams.change
			$scope.selectedChange.retrieveInformation $routeParams.repositoryId, $routeParams.change
		else
			$scope.selectedChange.clear()

		if $routeParams.change? and $routeParams.stage?
			$scope.selectedStage.setId $routeParams.repositoryId, $routeParams.change, $routeParams.stage
			$scope.selectedStage.retrieveInformation $routeParams.repositoryId, $routeParams.stage
		else
			$scope.selectedStage.clear()

		$scope.selectedStage.setSummary() if not $routeParams.stage?
		$scope.selectedStage.setSkipped() if $routeParams.skipped?
		$scope.selectedStage.setMerge() if $routeParams.merge?
		$scope.selectedStage.setDebug() if $routeParams.debug?
	syncToRouteParams()

	$scope.$watch 'selectedRepository.getInformation().type + selectedRepository.getInformation().uri', () ->
		repositoryInformation = $scope.selectedRepository.getInformation()

		if repositoryInformation?
			$scope.cloneUri = repositoryInformation.type + ' clone ' + repositoryInformation.uri

	$scope.$watch 'selectedChange.getId()', (newValue) ->
		$location.search 'change', newValue ? null

	$scope.$watch 'selectedStage.getId()', (newValue) ->
		$location.search 'stage', newValue ? null

	$scope.$watch 'selectedStage.isSkipped()', (newValue) ->
		$location.search 'skipped', if newValue then true else null

	$scope.$watch 'selectedStage.isMerge()', (newValue) ->
		$location.search 'merge', if newValue then true else null

	$scope.$watch 'selectedStage.isDebug()', (newValue) ->
		$location.search 'debug', if newValue then true else null
]
