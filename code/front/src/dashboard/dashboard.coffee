'use strict'

window.Dashboard = ['$scope', '$http', 'ChangesManager', 'localStorage', ($scope, $http, ChangesManager, localStorage) ->
	repositoryCache = {}

	$scope.search =
		mode: localStorage.dashboardSearchMode ? 'all'
		query: ''

	getRepositories = () ->
		request = $http.get('/app/repositories/')
		request.success (data, status, headers, config) ->
			console.log data
			$scope.repositories = data
			createRepositoryCache()
			getChanges()
		request.error (data, status, headers, config) ->
			notification.error data

	createRepositoryCache = () ->
		for repository in $scope.repositories
			repositoryCache[repository.id] = repository

	updateChangesWithRepositoryInformation = () ->
		assert.ok repositoryCache?

		return if not $scope.changesManager?

		for change in $scope.changesManager.getChanges()
			repository = repositoryCache[change.repository.id]
			change.repository = repository if repository?

	getChanges = () ->
		repositoryIds = $scope.repositories.map (repository) -> return repository.id
		$scope.changesManager = ChangesManager.create repositoryIds, $scope.search
		
		$scope.changesManager.listenToEvents()
		$scope.$on '$destroy', $scope.changesManager.stopListeningToEvents

		$scope.changesManager.retrieveInitialChanges()

	getRepositories()

	$scope.$watch 'changesManager.getChanges()', (() ->
		updateChangesWithRepositoryInformation()
	), true

	$scope.$watch 'search', ((newValue, oldValue) ->
		return if newValue is oldValue
		$scope.changesManager.retrieveInitialChanges() if $scope.changesManager?
		localStorage.dashboardSearchMode = $scope.search.mode
	), true

]