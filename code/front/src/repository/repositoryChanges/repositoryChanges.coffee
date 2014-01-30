'use strict'

window.RepositoryChanges = ['$scope', 'ChangesManager', 'currentRepository', 'currentChange', 'localStorage', ($scope, ChangesManager, currentRepository, currentChange, localStorage) ->
	$scope.selectedRepository = currentRepository
	$scope.selectedChange = currentChange

	$scope.search =
		mode: localStorage.repositoryChangesSearchMode ? 'all'
		query: ''

	$scope.changesManager = ChangesManager.create [$scope.selectedRepository.getId()], $scope.search
	$scope.changesManager.listenToEvents()
	$scope.$on '$destroy', $scope.changesManager.stopListeningToEvents
	$scope.changesManager.retrieveInitialChanges()

	$scope.selectChange = (change) ->
		$scope.selectedChange.setId $scope.selectedRepository.getId(), change.id
		$scope.selectedChange.setInformation change

	$scope.$watch 'changesManager.getChanges()', (() ->
		if not $scope.selectedChange.getId()? and $scope.changesManager.getChanges().length > 0
			firstChange = $scope.changesManager.getChanges()[0]
			$scope.selectedChange.setId $scope.selectedRepository.getId(), firstChange.id
			$scope.selectedChange.setInformation firstChange
	), true

	$scope.$watch 'search', ((newValue, oldValue) ->
		return if newValue is oldValue
		$scope.changesManager.retrieveInitialChanges()
		localStorage.repositoryChangesSearchMode = $scope.search.mode
	), true
]
