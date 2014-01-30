'use strict'

window.Analytics = ['$scope', 'rpc', 'events', ($scope, rpc, events) ->
	allRepository = {id: -1, name: 'All Repositories'}

	$scope.allowedGraphs = [
		{value: 'all', title: 'Passed / Failed Changes'}
		{value: 'passed', title: 'Passed Rate'}
		{value: 'failed', title: 'Failed Rate'}
	]

	$scope.allowedDurations = [
		{value: 'today', title: 'Today'}
		{value: 'yesterday', title: 'Yesterday'}
		{value: 'last7', title: 'Last 7 Days'}
		{value: 'last30', title: 'Last 30 Days'}
		{value: 'last365', title: 'Last 12 Months'}
	]

	allowedIntervals =
		hour: {value: 'hour', title: 'Hours'}
		day: {value: 'day', title: 'Days'}
		week: {value: 'week', title: 'Weeks'}
		month: {value: 'month', title: 'Months'}

	$scope.options = 
		graph: $scope.allowedGraphs[0]
		repository: allRepository
		repositories: [allRepository]
		duration: $scope.allowedDurations[1]
		interval: allowedIntervals.hour

	$scope.graphOptions =
		graphType: $scope.allowedGraphs[0].value
		changes: []
		start: new Date()
		end: new Date()
		interval: allowedIntervals.hour.value

	setGraphBounds = () ->
		getMidnightOfDateDelta = (dateDelta) ->
			midnightToday = new Date((new Date()).setHours(0, 0, 0, 0))
			return new Date(midnightToday.setDate(midnightToday.getDate() + dateDelta))

		if $scope.options.duration.value is 'today'
			$scope.graphOptions.start = getMidnightOfDateDelta 0
			$scope.graphOptions.end = getMidnightOfDateDelta 1

		if $scope.options.duration.value is 'yesterday'
			$scope.graphOptions.start = getMidnightOfDateDelta -1
			$scope.graphOptions.end = getMidnightOfDateDelta 0

		if $scope.options.duration.value is 'last7'
			$scope.graphOptions.start = getMidnightOfDateDelta -7
			$scope.graphOptions.end = getMidnightOfDateDelta 0

		if $scope.options.duration.value is 'last30'
			$scope.graphOptions.start = getMidnightOfDateDelta -30
			$scope.graphOptions.end = getMidnightOfDateDelta 0

		if $scope.options.duration.value is 'last365'
			$scope.graphOptions.start = getMidnightOfDateDelta -365
			$scope.graphOptions.end = getMidnightOfDateDelta 0

	updateAllowedIntervals = () ->
		if $scope.options.duration.value is 'today' then $scope.options.intervals = [allowedIntervals.hour]
		if $scope.options.duration.value is 'yesterday' then $scope.options.intervals = [allowedIntervals.hour]
		if $scope.options.duration.value is 'last7' then $scope.options.intervals = [allowedIntervals.hour, allowedIntervals.day]
		if $scope.options.duration.value is 'last30' then $scope.options.intervals = [allowedIntervals.day, allowedIntervals.week]
		if $scope.options.duration.value is 'last365' then $scope.options.intervals = [allowedIntervals.day, allowedIntervals.week, allowedIntervals.month]

		matchingInterval = $scope.options.intervals.some (interval) -> return interval.value is $scope.options.interval.value
		$scope.options.interval = $scope.options.intervals[0] if not matchingInterval

	getRepositories = () ->
		rpc 'repositories', 'read', 'getRepositories', null, (error, repositories) ->
			$scope.options.repositories = [allRepository].concat repositories
			updateChangeFinishedListeners()
			getChanges()

	getRepositoryIdsToDisplay = () ->
		if $scope.options.repository.id >= 0 then return [$scope.options.repository.id]
		else
			return $scope.options.repositories
				.map((repository) -> return repository.id)
				.filter((repositoryId) -> repositoryId >= 0)

	getChanges = () ->
		return if $scope.options.repositories.every (repository) -> return repository.id < 0

		$scope.graphOptions.changes = []

		requestParams =
			repositories: getRepositoryIdsToDisplay()
			startTimestamp: $scope.graphOptions.start.getTime()
			endTimestamp: $scope.graphOptions.end.getTime()
		rpc 'changes', 'read', 'getChangesBetweenTimestamps', requestParams, (error, changes) ->
			$scope.graphOptions.changes = changes

	getChangeWithId = (id) ->
		return (change for change in $scope.graphOptions.changes when change.id is id)[0]

	handleChangeFinished = (data) ->
		return if not data.resourceId in getRepositoryIdsToDisplay()
		$scope.graphOptions.changes.push data if not getChangeWithId(data.id)?

	changeFinishedListeners = []
	updateChangeFinishedListeners = () ->
		changeFinishedListener.unsubscribe() for changeFinishedListener in changeFinishedListeners
		changeFinishedListeners = []

		return if $scope.options.duration.value isnt 'today'

		for repositoryId in getRepositoryIdsToDisplay()
			changeFinishedListener = events('repositories', 'change finished', repositoryId).setCallback(handleChangeFinished).subscribe()
			changeFinishedListeners.push changeFinishedListener
	$scope.$on '$destroy', () -> changeFinishedListener.unsubscribe() for changeFinishedListener in changeFinishedListeners

	getRepositories()

	$scope.$watch 'options.graph', () ->
		$scope.graphOptions.graphType = $scope.options.graph.value

	$scope.$watch 'options.repository', () -> 
		updateChangeFinishedListeners()
		getChanges()

	$scope.$watch 'options.duration', () ->
		updateAllowedIntervals()
		setGraphBounds()
		updateChangeFinishedListeners()
		getChanges()

	$scope.$watch 'options.interval', () ->
		$scope.graphOptions.interval = $scope.options.interval.value
]
