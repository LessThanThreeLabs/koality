'use strict'

window.Dashboard = ['$scope', '$http', 'BuildsManager', 'localStorage', ($scope, $http, BuildsManager, localStorage) ->
	repositoryCache = {}

	$scope.search =
		mode: localStorage.dashboardSearchMode ? 'all'
		query: ''

	getRepositories = () ->
		request = $http.get('/app/repositories/')
		request.success (data, status, headers, config) ->
			$scope.repositories = data
			createRepositoryCache()
			getBuilds()
		request.error (data, status, headers, config) ->
			notification.error data

	createRepositoryCache = () ->
		for repository in $scope.repositories
			repositoryCache[repository.id] = repository

	updateBuildsWithRepositoryInformation = () ->
		assert.ok repositoryCache?

		return if not $scope.buildsManager?

		for build in $scope.buildsManager.getBuilds()
			repository = repositoryCache[build.repositoryId]
			build.repository = repository if repository?

	getBuilds = () ->
		repositoryIds = $scope.repositories.map (repository) -> return repository.id
		$scope.buildsManager = BuildsManager.create repositoryIds, $scope.search
		
		$scope.buildsManager.listenToEvents()
		$scope.$on '$destroy', $scope.buildsManager.stopListeningToEvents

		$scope.buildsManager.retrieveInitialBuilds()

	getRepositories()

	$scope.$watch 'buildsManager.getBuilds()', (() ->
		updateBuildsWithRepositoryInformation()
	), true

	$scope.$watch 'search', ((newValue, oldValue) ->
		return if newValue is oldValue
		$scope.buildsManager.retrieveInitialBuilds() if $scope.buildsManager?
		localStorage.dashboardSearchMode = $scope.search.mode
	), true

]