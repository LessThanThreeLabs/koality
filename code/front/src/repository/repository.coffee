'use strict'

window.Repository = ['$scope', '$location', '$routeParams', 'currentRepository', 'currentBuild', 'currentStage', 'currentStageRun', ($scope, $location, $routeParams, currentRepository, currentBuild, currentStage, currentStageRun) ->
	$scope.selectedRepository = currentRepository
	$scope.selectedBuild = currentBuild
	$scope.selectedStage = currentStage
	$scope.selectedStageRun = currentStageRun

	syncToRouteParams = () ->
		$scope.selectedRepository.setId $routeParams.repositoryId
		$scope.selectedRepository.retrieveInformation()

		if $routeParams.build?
			$scope.selectedBuild.setId $routeParams.repositoryId, $routeParams.build
			$scope.selectedBuild.retrieveInformation()
		else
			$scope.selectedBuild.clear()

		if $routeParams.build? and $routeParams.stage?
			$scope.selectedStage.setId $routeParams.repositoryId, $routeParams.build, $routeParams.stage
			$scope.selectedStage.retrieveInformation()
		else
			$scope.selectedStage.clear()
		$scope.selectedStage.setSummary() if not $routeParams.stage?

		if $routeParams.build? and $routeParams.stage? and $routeParams.run?
			$scope.selectedStageRun.setId $routeParams.repositoryId, $routeParams.build, $routeParams.stage, $routeParams.run
			$scope.selectedStageRun.retrieveInformation()
		else
			$scope.selectedStageRun.clear()
	syncToRouteParams()

	$scope.$watch 'selectedRepository.getInformation().type + selectedRepository.getInformation().uri', () ->
		repositoryInformation = $scope.selectedRepository.getInformation()

		if repositoryInformation?
			$scope.cloneUri = repositoryInformation.vcsType + ' clone ' + repositoryInformation.uri

	$scope.$watch 'selectedBuild.getId()', (newValue) ->
		$location.search 'build', newValue ? null

	$scope.$watch 'selectedStage.getId()', (newValue) ->
		$location.search 'stage', newValue ? null

	$scope.$watch 'selectedStageRun.getId()', (newValue) ->
		$location.search 'run', newValue ? null
]
