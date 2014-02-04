'use strict'

window.RepositoryBuilds = ['$scope', 'BuildsManager', 'currentRepository', 'currentBuild', 'localStorage', ($scope, BuildsManager, currentRepository, currentBuild, localStorage) ->
	$scope.selectedRepository = currentRepository
	$scope.selectedBuild = currentBuild

	$scope.search =
		mode: localStorage.repositoryBuildsSearchMode ? 'all'
		query: ''

	$scope.buildsManager = BuildsManager.create [$scope.selectedRepository.getId()], $scope.search
	$scope.buildsManager.listenToEvents()
	$scope.$on '$destroy', $scope.buildsManager.stopListeningToEvents
	$scope.buildsManager.retrieveInitialBuilds()

	$scope.selectBuild = (build) ->
		$scope.selectedBuild.setId $scope.selectedRepository.getId(), build.id
		$scope.selectedBuild.setInformation build

	$scope.$watch 'buildsManager.getBuilds()', (() ->
		if not $scope.selectedBuild.getId()? and $scope.buildsManager.getBuilds().length > 0
			firstBuild = $scope.buildsManager.getBuilds()[0]
			$scope.selectedBuild.setId $scope.selectedRepository.getId(), firstBuild.id
			$scope.selectedBuild.setInformation firstBuild
	), true

	$scope.$watch 'search', ((newValue, oldValue) ->
		return if newValue is oldValue
		$scope.buildsManager.retrieveInitialBuilds()
		localStorage.repositoryBuildsSearchMode = $scope.search.mode
	), true
]
