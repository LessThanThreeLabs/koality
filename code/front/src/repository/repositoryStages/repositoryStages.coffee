'use strict'

window.RepositoryStages = ['$scope', '$routeParams', 'StagesManager', 'currentRepository', 'currentChange', 'currentStage', 'localStorage', ($scope, $routeParams, StagesManager, currentRepository, currentChange, currentStage, localStorage) ->
	$scope.selectedRepository = currentRepository
	$scope.selectedChange = currentChange
	$scope.selectedStage = currentStage

	$scope.filter = type: localStorage.repositoryStagesFilterType ? 'all'

	$scope.stagesManager = StagesManager.create()
	$scope.$on '$destroy', $scope.stagesManager.stopListeningToEvents

	getMostImportantStageWithTypeAndName = (type, name) ->
		mostImportantStage = null

		for potentialStage in $scope.stagesManager.getStages()
			continue if potentialStage.type isnt type or potentialStage.name isnt name

			if not mostImportantStage?
				mostImportantStage = potentialStage
			else
				if potentialStage.status is 'failed' and mostImportantStage.status is 'failed'
					mostImportantStage = potentialStage if potentialStage.id < mostImportantStage.id
				else if potentialStage.status is 'failed' and mostImportantStage.status isnt 'failed'
					mostImportantStage = potentialStage
				else if potentialStage.status isnt 'failed' and mostImportantStage.status isnt 'failed'
					mostImportantStage = potentialStage if potentialStage.id < mostImportantStage.id

		return mostImportantStage

	bringFailedMirrorStageToForeground = () ->
		return if not $scope.selectedStage.getId()?
		return if $scope.selectedStage.getInformation().status is 'failed'

		mirrorsOfSelectedStage = $scope.stagesManager.getStages().filter (stage) ->
			return isMirrorStage stage, $scope.selectedStage.getInformation()

		for mirrorStage in mirrorsOfSelectedStage
			continue if $scope.selectedStage.getId() is mirrorStage.id

			if mirrorStage.status is 'failed'
				$scope.selectedStage.setId $scope.selectedRepository.getId(), $scope.selectedChange.getId(), mirrorStage.id
				$scope.selectedStage.setInformation mirrorStage
				return

	isMirrorStage = (stage1, stage2) ->
		return false if not stage1? or not stage2?
		return stage1.type is stage2.type and stage1.name is stage2.name

	$scope.stageSort = (stage) ->
		if stage.type is 'setup'
			return 10000 + stage.orderNumber
		else if stage.type is 'compile'
			return 20000 + stage.orderNumber
		else if stage.type is 'testFactory'
			return 30000 + stage.orderNumber
		else if stage.type is 'test'
			return 40000 + stage.orderNumber
		else if stage.type is 'export'
			return 50000 + stage.orderNumber
		else
			console.error 'Cannot sort stage'
			return 60000

	$scope.shouldStageBeVisible = (stage) ->
		return true if stage.id is $scope.selectedStage.getId()
		return false if $scope.filter.type is 'failed' and stage.status isnt 'failed'
		return false if isMirrorStage stage, $scope.selectedStage.getInformation()
		return true if stage.id is getMostImportantStageWithTypeAndName(stage.type, stage.name).id
		return false

	$scope.hasNoFailedStages = () ->
		return not $scope.stagesManager.getStages().some (stage) -> 
			return stage.status is 'failed'

	$scope.selectStage = (stage) ->
		$scope.selectedStage.setId $scope.selectedRepository.getId(), $scope.selectedChange.getId(), stage.id
		$scope.selectedStage.setInformation stage

	$scope.$watch 'filter.type', (newFilterType, oldFilterType) ->
		return if newFilterType is oldFilterType

		localStorage.repositoryStagesFilterType = $scope.filter.type

		if $scope.filter.type is 'failed'
			status = $scope.selectedStage.getInformation()?.status
			if status? and status isnt 'failed'
				$scope.selectedStage.setSummary()

	$scope.$watch 'selectedChange.getId()', (newChangeId, oldChangeId) ->
		$scope.selectedStage.setSummary() if newChangeId isnt oldChangeId

		$scope.stagesManager.setChangeId $scope.selectedChange.getId()
		$scope.stagesManager.stopListeningToEvents()

		if $scope.selectedChange.getId()?
			$scope.stagesManager.listenToEvents()
			$scope.stagesManager.retrieveStages()

	$scope.$watch 'stagesManager.getStages()', ((newValue, oldValue) ->
		return if newValue is oldValue

		if $scope.selectedStage.getId()?
			stagesContainSelectedStageId = $scope.stagesManager.getStages().some (stage) ->
				return $scope.selectedStage.getId() is stage.id

			if not stagesContainSelectedStageId
				$scope.selectedStage.setSummary()

		bringFailedMirrorStageToForeground()
	), true
]