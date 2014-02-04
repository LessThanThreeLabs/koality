'use strict'

window.Repository = ['$scope', '$location', '$routeParams', 'currentRepository', 'currentBuild', 'currentStage', ($scope, $location, $routeParams, currentRepository, currentBuild, currentStage) ->
	$scope.selectedRepository = currentRepository
	$scope.selectedBuild = currentBuild
	$scope.selectedStage = currentStage

	syncToRouteParams = () ->
		$scope.selectedRepository.setId $routeParams.repositoryId
		$scope.selectedRepository.retrieveInformation $routeParams.repositoryId

		if $routeParams.build?
			$scope.selectedBuild.setId $routeParams.repositoryId, $routeParams.build
			$scope.selectedBuild.retrieveInformation $routeParams.repositoryId, $routeParams.build
		else
			$scope.selectedBuild.clear()

		# if $routeParams.build? and $routeParams.stage?
		# 	$scope.selectedStage.setId $routeParams.repositoryId, $routeParams.build, $routeParams.stage
		# 	$scope.selectedStage.retrieveInformation $routeParams.repositoryId, $routeParams.stage
		# else
		# 	$scope.selectedStage.clear()

		# $scope.selectedStage.setSummary() if not $routeParams.stage?
		# $scope.selectedStage.setSkipped() if $routeParams.skipped?
		# $scope.selectedStage.setMerge() if $routeParams.merge?
		# $scope.selectedStage.setDebug() if $routeParams.debug?
	syncToRouteParams()

	$scope.$watch 'selectedRepository.getInformation().type + selectedRepository.getInformation().uri', () ->
		repositoryInformation = $scope.selectedRepository.getInformation()

		if repositoryInformation?
			$scope.cloneUri = repositoryInformation.vcsType + ' clone ' + repositoryInformation.uri

	$scope.$watch 'selectedBuild.getId()', (newValue) ->
		$location.search 'build', newValue ? null

	# $scope.$watch 'selectedStage.getId()', (newValue) ->
	# 	$location.search 'stage', newValue ? null

	# $scope.$watch 'selectedStage.isSkipped()', (newValue) ->
	# 	$location.search 'skipped', if newValue then true else null

	# $scope.$watch 'selectedStage.isMerge()', (newValue) ->
	# 	$location.search 'merge', if newValue then true else null

	# $scope.$watch 'selectedStage.isDebug()', (newValue) ->
	# 	$location.search 'debug', if newValue then true else null
]
