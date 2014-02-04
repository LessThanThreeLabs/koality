'use strict'

angular.module('koality.service.managers', []).
	factory('BuildsManager', ['initialState', 'BuildsRpc', 'events', (initialState, BuildsRpc, events) ->
		class BuildsManager
			_builds: []
			_buildsCache: {}
			_gettingMoreBuilds: false
			_currentRequestId: null

			# _changeAddedListeners: []
			# _changeStartedListeners: []
			# _changeFinishedListeners: []

			constructor: (@repositoryIds, @searchModel) ->
				assert.ok typeof @repositoryIds is 'object'
				assert.ok typeof @searchModel is 'object'

				@buildsRpc = BuildsRpc.create()

			_getGroupFromMode: () =>
				if @searchModel.mode is 'all' or @searchModel.mode is 'me'
					return @searchModel.mode
				if @searchModel.query.trim() is ''
					return 'all'
				return null

			_getQuery: () =>
				if @searchModel.mode isnt 'search' then return null

				query = @searchModel.query.trim()
				if query is '' then return null
				else return query

			_doesBuildMatchQuery: (build) =>
				console.log 'NEED TO FIX THIS!'
				return true
				# if @searchModel.mode is 'me'
				# 	return initialState.user.id is change.user.id
				# else
				# 	return true if @searchModel.query.trim() is ''

				# 	stringsToMatch = @searchModel.query.trim().split(' ')
				# 		.filter((string) -> return string isnt '')
				# 		.map((string) -> return string.toLowerCase())

				# 	return (change.user.name.first.toLowerCase() in stringsToMatch) or
				# 		(change.user.name.last.toLowerCase() in stringsToMatch) or
				# 		(change.headCommit.sha.toLowerCase() in stringsToMatch)

			_buildsRetrievedHandler: (error, buildsData) =>
				if error? then console.error error
				else
					return if @_currentRequestId isnt buildsData.requestId
					buildsData.builds = buildsData.builds.filter (build) =>
						return not @_buildsCache[build.id]?
					@_buildsCache[build.id] = build for build in buildsData.builds
					@_builds = @_builds.concat buildsData.builds
					@_gettingMoreBuilds = false

			# _handleChangeAdded: (data) =>
			# 	return if not data.resourceId in @repositoryIds
			# 	if @_doesBuildMatchQuery(data) and not @_buildsCache[data.id]?
			# 		@_buildsCache[data.id] = data
			# 		@_builds.unshift data

			# _handleChangeStarted: (data) =>
			# 	return if not data.resourceId in @repositoryIds
			# 	change = @_buildsCache[data.id]
			# 	$.extend true, change, data if change?

			# _handleChangeFinished: (data) =>
			# 	return if not data.resourceId in @repositoryIds
			# 	change = @_buildsCache[data.id]
			# 	$.extend true, change, data if change?

			# _addListeners: (listeners, eventType, handler) =>
			# 	@_removeListeners listeners

			# 	for repositoryId in @repositoryIds
			# 		listener = events('repositories', eventType, repositoryId).setCallback(handler).subscribe()
			# 		listeners.push listener

			# _removeListeners: (listeners) =>
			# 	listener.unsubscribe() for listener in listeners
			# 	listeners.length = 0

			retrieveInitialBuilds: () =>
				@_builds = []
				@_buildsCache = {}
				@_gettingMoreBuilds = true
				@_currentRequestId = @buildsRpc.queueRequest @repositoryIds, @_getGroupFromMode(), @_getQuery(), 0, @_buildsRetrievedHandler

			retrieveMoreChanges: () =>
				return if @_builds.length is 0
				return if @_gettingMoreBuilds
				return if not @buildsRpc.hasMoreBuildsToRequest()
				@_gettingMoreBuilds = true
				@_currentRequestId = @buildsRpc.queueRequest @repositoryIds, @_getGroupFromMode(), @_getQuery(), @_builds.length, @_buildsRetrievedHandler	

			getBuilds: () =>
				return @_builds

			isRetrievingBuilds: () =>
				return @_gettingMoreBuilds

			listenToEvents: () =>
			# 	@_addListeners @_changeAddedListeners, 'change added', @_handleChangeAdded
			# 	@_addListeners @_changeStartedListeners, 'change started', @_handleChangeStarted
			# 	@_addListeners @_changeFinishedListeners, 'change finished', @_handleChangeFinished
					
			stopListeningToEvents: () =>
			# 	@_removeListeners @_changeAddedListeners
			# 	@_removeListeners @_changeStartedListeners
			# 	@_removeListeners @_changeFinishedListeners

		return create: (repositoryIds, searchModel) ->
			return new BuildsManager repositoryIds, searchModel
	]).
	factory('StagesManager', ['$http', 'events', ($http, events) ->
		class StagesManager
			_buildId: null

			_stages: []
			_stagesCache: {}
			_gettingStages: false

			# _stageAddedListener: null
			# _stageUpdatedListener: null
			# _stageOutputTypesListener: null

			_stagesRetrievedHandler: (error, stagesToAdd) =>
				if error? then console.error error
				else
					stagesToAdd = stagesToAdd.filter (stage) => return not @_stagesCache[stage.id]?
					@_stagesCache[stage.id] = stage for stage in stagesToAdd
					@_stages = @_stages.concat stagesToAdd
					@_gettingStages = false

			_handleStageAdded: (data) =>
				return if data.resourceId isnt @_buildId
				if not @_stagesCache[data.id]?
					@_stagesCache[data.id] = data
					@_stages.push data

			_handleStageUpdated: (data) =>
				return if data.resourceId isnt @_buildId
				stage = @_stagesCache[data.id]
				$.extend true, stage, data if stage?

			_handleStageOutputTypeAdded: (data) =>
				return if data.resourceId isnt @_buildId
				stage = @_stagesCache[data.id]

				if stage? and not (data.outputType in stage.outputTypes)
					stage.outputTypes.push data.outputType

			setBuildId: (buildId) =>
				assert.ok not buildId? or typeof buildId is 'number'

				if @_buildId isnt buildId
					@_stages = []
					@_stagesCache = {}
					@_buildId = buildId
					@stopListeningToEvents()

			retrieveStages: () =>
				assert.ok @_buildId?

				@_stages = []
				@_stagesCache = {}
				@_gettingStages = true

				request = $http.get("/app/stages/?verificationId=#{@_buildId}")
				request.success (data, status, headers, config) =>
					@_stagesRetrievedHandler null, data
				request.error (data, status, headers, config) =>
					@_stagesRetrievedHandler data

			getStages: () =>
				return @_stages

			isRetrievingStages: () =>
				return @_gettingStages

			listenToEvents: () =>
			# 	assert.ok @_buildId?

			# 	@stopListeningToEvents()

			# 	@_stageAddedListener = events('changes', 'new build console', @_buildId).setCallback(@_handleStageAdded).subscribe()
			# 	@_stageUpdatedListener = events('changes', 'return code added', @_buildId).setCallback(@_handleStageUpdated).subscribe()
			# 	@_stageOutputTypesListener = events('changes', 'output type added', @_buildId).setCallback(@_handleStageOutputTypeAdded).subscribe()
					
			stopListeningToEvents: () =>
			# 	@_stageAddedListener.unsubscribe() if @_stageAddedListener?
			# 	@_stageUpdatedListener.unsubscribe() if @_stageUpdatedListener?
			# 	@_stageOutputTypesListener.unsubscribe() if @_stageOutputTypesListener

			# 	@_stageAddedListener = null
			# 	@_stageUpdatedListener = null
			# 	@_stageOutputTypesListener = null

		return create: () ->
			return new StagesManager()
	]).
	factory('ConsoleLinesManager', ['$rootScope', '$timeout', 'ConsoleLinesRpc', 'events', 'stringHasher', 'integerConverter', ($rootScope, $timeout, ConsoleLinesRpc, events, stringHasher, integerConverter) ->
		class ConsoleLinesManager
			_stageId: null
			_currentRequestId: null

			_oldLines: {}
			_newLines: {}
			_allowGettingMoreLines: true
			_gettingMoreLines: false

			_linesAddedListener: null

			constructor: () ->
				@consoleTextRpc = ConsoleTextRpc.create()

			_linesRetrievedHandler: (error, linesData) =>
				if error? then console.error error
				else 
					return if @_currentRequestId isnt linesData.requestId
					@_processNewLines linesData.lines
					@_gettingMoreLines = false
					@_allowGettingMoreLines = false
					$timeout (() => @_allowGettingMoreLines = true), 100

			# _handleLinesAdded: (data) =>
			# 	return if data.resourceId isnt @_stageId
			# 	@_processNewLines data.lines

			_processNewLines: (data) =>
				@_mergeNewLinesWithOldLines()
				@_newLines = {}

				for lineNumber, lineText of data
					lineNumber = integerConverter.toInteger lineNumber
					lineHash = stringHasher.hash lineText
					@_newLines[lineNumber] =
						text: lineText
						hash: lineHash

			_mergeNewLinesWithOldLines: () =>
				for lineNumber, line of @_newLines
					@_oldLines[lineNumber] = line

			clear: () =>
				@_stageId = null
				@_newLines = {}
				@_oldLines = {}
				@_currentRequestId = null
				@stopListeningToEvents()

			setStageId: (stageId) =>
				assert.ok not stageId? or typeof stageId is 'number'

				if @_stageId isnt stageId
					@_stageId = stageId
					@_newLines = {}
					@_oldLines = {}
					@_currentRequestId = null
					@stopListeningToEvents()

			retrieveInitialLines: () =>
				assert.ok @_stageId?

				@_newLines = {}
				@_oldLines = {}
				@_gettingMoreLines = true
				@_currentRequestId = @consoleTextRpc.queueRequest @_stageId, 0, @_linesRetrievedHandler

			retrieveMoreLines: () =>
				getStartIndex = () =>
					startIndex = Object.keys(@_oldLines).length
					for lineNumber in Object.keys(@_newLines)
						startIndex++ if not @_oldLines[lineNumber]?
					return startIndex

				return if Object.keys(@_newLines).length is 0
				return if not @_allowGettingMoreLines
				return if @_gettingMoreLines
				return if not @consoleTextRpc.hasMoreLinesToRequest()

				@_gettingMoreLines = true
				@_currentRequestId = @consoleTextRpc.queueRequest @_stageId, getStartIndex(), @_linesRetrievedHandler

			getNewLines: () =>
				return @_newLines

			getOldLines: () =>
				return @_oldLines

			removeLines: (startIndex, numLines) =>
				startIndex = integerConverter.toInteger startIndex
				numLines = integerConverter.toInteger numLines

				for lineNumber in [startIndex...(startIndex+numLines)]
					delete @_oldLines[lineNumber]
				@consoleTextRpc.notifyLinesRemoved()

			isRetrievingLines: () =>
				return @_gettingMoreLines

			listenToEvents: () =>
				# assert.ok @_stageId?

				# @stopListeningToEvents()
				# @_linesAddedListener = events('buildConsoles', 'new output', @_stageId).setCallback(@_handleLinesAdded).subscribe()
					
			stopListeningToEvents: () =>
				# @_linesAddedListener.unsubscribe() if @_linesAddedListener?
				# @_linesAddedListener = null

		return create: () ->
			return new ConsoleLinesManager()
	])
