'use strict'

window.RepositoryStageDetails = ['$scope', '$location', '$http', 'events', 'ConsoleLinesManager', 'xUnitParser', 'currentRepository', 'currentBuild', 'currentStage', 'currentStageRun', 'notification', ($scope, $location, $http, events, ConsoleLinesManager, xUnitParser, currentRepository, currentBuild, currentStage, currentStageRun, notification) ->
	$scope.selectedRepository = currentRepository
	$scope.selectedBuild = currentBuild
	$scope.selectedStage = currentStage
	$scope.selectedStageRun = currentStageRun

	$scope.output = type: null
	$scope.xunit =
		makingRequest: false
		testCases: []
		orderByPredicate: 'status'
		orderByReverse: false
		maxResults: 100
	$scope.debugInstance =
		durationInMinutes: 60
		makingRequest: false
	$scope.retrigger =
		makingRequest: false

	$scope.currentlyOpenDrawer = null

	$scope.consoleLinesManager = ConsoleLinesManager.create()
	$scope.$on '$destroy', $scope.consoleLinesManager.stopListeningToEvents

	updateUrl = () ->
		$scope.currentUrl = $location.absUrl()
	$scope.$on '$routeUpdate', updateUrl
	updateUrl()

	# retrieveCurrentChangeExportUris = () ->
	# 	$scope.exportUris = []
	# 	return if not $scope.selectedBuild.getId()?

	# 	rpc 'changes', 'read', 'getChangeExportUris', id: $scope.selectedBuild.getId(), (error, exportUris) ->
	# 		$scope.exportUris = exportUris

	# retrieveXUnitOutput = () ->
	# 	assert.ok $scope.output.type is 'xunit'
	# 	return if $scope.xunit.makingRequest

	# 	$scope.xunit.testCases = []
	# 	return if not $scope.selectedStage.getId()?

	# 	$scope.xunit.makingRequest = true
	# 	rpc 'buildConsoles', 'read', 'getXUnit', id: $scope.selectedStage.getId(), (error, xunitOutputs) ->
	# 		$scope.xunit.makingRequest = false
	# 		$scope.xunit.testCases = xUnitParser.getTestCases xunitOutputs

	# handleExportUrisAdded = (data) ->
	# 	return if data.resourceId isnt $scope.selectedBuild.getId()
	# 	$scope.exportUris ?= []
	# 	$scope.exportUris = $scope.exportUris.concat data.exportMetadata

	# addedExportUrisEvents = null
	# updateExportUrisAddedListener = () ->
	# 	if addedExportUrisEvents?
	# 		addedExportUrisEvents.unsubscribe()
	# 		addedExportUrisEvents = null

	# 	if $scope.selectedBuild.getId()?
	# 		addedExportUrisEvents = events('changes', 'export metadata added', $scope.selectedBuild.getId()).setCallback(handleExportUrisAdded).subscribe()
	# $scope.$on '$destroy', () -> addedExportUrisEvents.unsubscribe() if addedExportUrisEvents?

	$scope.toggleDrawer = (drawerName) ->
		if $scope.currentlyOpenDrawer is drawerName
			$scope.currentlyOpenDrawer = null
		else
			$scope.currentlyOpenDrawer = drawerName

	# $scope.retrigger = () ->
	# 	return if $scope.retrigger.makingRequest
	# 	$scope.retrigger.makingRequest = true

	# 	requestParams = id: $scope.selectedBuild.getId()
	# 	rpc 'changes', 'create', 'retrigger', requestParams, (error, createdChange) ->
	# 		$scope.retrigger.makingRequest = false
	# 		if error? then notification.error error
	# 		else
	# 			notification.success 'Change has been retriggered'
	# 			$scope.clearRetrigger()

	# 			$scope.selectedBuild.setId $scope.selectedRepository.getId(), createdChange.id
	# 			$scope.selectedBuild.retrieveInformation $scope.selectedRepository.getId(), createdChange.id

	$scope.clearRetrigger = () ->
		$scope.currentlyOpenDrawer = null

	# $scope.launchDebugInstance = () ->
	# 	return if $scope.debugInstance.makingRequest
	# 	$scope.debugInstance.makingRequest = true

	# 	requestParams =
	# 		id: $scope.selectedBuild.getId()
	# 		duration: $scope.debugInstance.durationInMinutes * 60 * 1000
	# 	rpc 'changes', 'create', 'launchDebugInstance', requestParams, (error) ->
	# 		$scope.debugInstance.makingRequest = false
	# 		if error? then notification.error error
	# 		else
	# 			notification.success 'You will be emailed shortly with information for accessing your debug instance', 30
	# 			$scope.clearLaunchDebugInstance()

	$scope.clearLaunchDebugInstance = () ->
		$scope.debugInstance.durationInMinutes = 60
		$scope.currentlyOpenDrawer = null

	$scope.getStageRunNumber = (stage, stageRun) ->
		return 0 if not stage?.runs? or not stageRun?

		for potentialStageRun, index in stage.runs
			return index + 1 if potentialStageRun.id is stageRun.id

		console.error 'Unable to determine number for stage run'
		return 10000

	$scope.selectStageRun = (stageRun) ->
		$scope.selectedStageRun.setId $scope.selectedRepository.getId(), $scope.selectedBuild.getId(), $scope.selectedStage.getId(), stageRun.id
		$scope.selectedStageRun.setInformation stageRun

	$scope.$watch 'selectedBuild.getId()', () ->
		# updateExportUrisAddedListener()
		# retrieveCurrentChangeExportUris()
		$scope.clearLaunchDebugInstance()

	$scope.$watch 'selectedStage.getId()', () ->
		$scope.output.type = null
		$scope.clearLaunchDebugInstance()

	$scope.$watch 'selectedStage.getInformation().runs', (() ->
		if not $scope.selectedStage.getInformation()?.runs?
			$scope.selectedStageRun.clear()
		else if $scope.selectedStage.getInformation().runs.length is 0
			$scope.selectedStageRun.clear()
		else
			selectedStageRunExists = $scope.selectedStage.getInformation().runs.some (stageRun) -> 
				return stageRun.id is $scope.selectedStageRun.getId()
			if not selectedStageRunExists
				firstStageRun = $scope.selectedStage.getInformation().runs[0]
				$scope.selectedStageRun.setId $scope.selectedRepository.getId(), $scope.selectedBuild.getId(), $scope.selectedStage.getId(), firstStageRun.id
				$scope.selectedStageRun.setInformation firstStageRun
	), true

	$scope.$watch 'selectedStageRun.getInformation()', (() ->
		console.log '...need to actually get output types somehow'
		$scope.selectedStageRun?.getInformation()?.outputTypes = ['console']

		return if not $scope.selectedStageRun.getInformation()?.outputTypes?

		$scope.output.hasConsole = 'console' in $scope.selectedStageRun.getInformation().outputTypes
		$scope.output.hasXUnit = 'xunit' in $scope.selectedStageRun.getInformation().outputTypes

		return if $scope.output.type?

		if 'xunit' in $scope.selectedStageRun.getInformation().outputTypes
			$scope.output.type = 'xunit'
		else if 'console' in $scope.selectedStageRun.getInformation().outputTypes
			$scope.output.type = 'console'
		else
			console.error 'No output type provided'
	), true

	$scope.$watch 'selectedStageRun.getId() + output.type', () ->
		return if not $scope.selectedStageRun.getId()?

		$scope.consoleLinesManager.clear()
		$scope.xunit.testCases = []

		if $scope.output.type is 'console'
			$scope.consoleLinesManager.setStageRunId $scope.selectedStageRun.getId()
			$scope.consoleLinesManager.listenToEvents()
			$scope.consoleLinesManager.retrieveInitialLines()

		if $scope.output.type is 'xunit'
			retrieveXUnitOutput()
]
