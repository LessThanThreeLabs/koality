'use strict'

angular.module('koality.service.manager', []).
	factory('ChangesManager', ['initialState', 'ChangesRpc', 'events', (initialState, ChangesRpc, events) ->
		class ChangesManager
			_changes: []
			_changesCache: {}
			_gettingMoreChanges: false
			_currentRequestId: null

			_changeAddedListeners: []
			_changeStartedListeners: []
			_changeFinishedListeners: []

			constructor: (@repositoryIds, @searchModel) ->
				assert.ok typeof @repositoryIds is 'object'
				assert.ok typeof @searchModel is 'object'

				@changesRpc = ChangesRpc.create()

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

			_doesChangeMatchQuery: (change) =>
				if @searchModel.mode is 'me'
					return initialState.user.id is change.user.id
				else
					return true if @searchModel.query.trim() is ''

					stringsToMatch = @searchModel.query.trim().split(' ')
						.filter((string) -> return string isnt '')
						.map((string) -> return string.toLowerCase())

					return (change.user.name.first.toLowerCase() in stringsToMatch) or
						(change.user.name.last.toLowerCase() in stringsToMatch) or
						(change.headCommit.sha.toLowerCase() in stringsToMatch)

			_changesRetrievedHandler: (error, changesData) =>
				return if @_currentRequestId isnt changesData.requestId

				changesData.changes = changesData.changes.filter (change) =>
					return not @_changesCache[change.id]?

				@_changesCache[change.id] = change for change in changesData.changes
				@_changes = @_changes.concat changesData.changes
				@_gettingMoreChanges = false

			_handleChangeAdded: (data) =>
				return if not data.resourceId in @repositoryIds
				if @_doesChangeMatchQuery(data) and not @_changesCache[data.id]?
					@_changesCache[data.id] = data
					@_changes.unshift data

			_handleChangeStarted: (data) =>
				return if not data.resourceId in @repositoryIds
				change = @_changesCache[data.id]
				$.extend true, change, data if change?

			_handleChangeFinished: (data) =>
				return if not data.resourceId in @repositoryIds
				change = @_changesCache[data.id]
				$.extend true, change, data if change?

			_addListeners: (listeners, eventType, handler) =>
				@_removeListeners listeners

				for repositoryId in @repositoryIds
					listener = events('repositories', eventType, repositoryId).setCallback(handler).subscribe()
					listeners.push listener

			_removeListeners: (listeners) =>
				listener.unsubscribe() for listener in listeners
				listeners.length = 0

			retrieveInitialChanges: () =>
				@_changes = []
				@_changesCache = {}
				@_gettingMoreChanges = true
				@_currentRequestId = @changesRpc.queueRequest @repositoryIds, @_getGroupFromMode(), @_getQuery(), 0, @_changesRetrievedHandler

			retrieveMoreChanges: () =>
				return if @_changes.length is 0
				return if @_gettingMoreChanges
				return if not @changesRpc.hasMoreChangesToRequest()
				@_gettingMoreChanges = true
				@_currentRequestId = @changesRpc.queueRequest @repositoryIds, @_getGroupFromMode(), @_getQuery(), @_changes.length, @_changesRetrievedHandler	

			getChanges: () =>
				return @_changes

			isRetrievingChanges: () =>
				return @_gettingMoreChanges

			listenToEvents: () =>
				@_addListeners @_changeAddedListeners, 'change added', @_handleChangeAdded
				@_addListeners @_changeStartedListeners, 'change started', @_handleChangeStarted
				@_addListeners @_changeFinishedListeners, 'change finished', @_handleChangeFinished
					
			stopListeningToEvents: () =>
				@_removeListeners @_changeAddedListeners
				@_removeListeners @_changeStartedListeners
				@_removeListeners @_changeFinishedListeners


		return create: (repositoryIds, searchModel) ->
			return new ChangesManager repositoryIds, searchModel
	]).
	factory('StagesManager', ['rpc', 'events', (rpc, events) ->
		class StagesManager
			_changeId: null

			_stages: []
			_stagesCache: {}
			_gettingStages: false

			_stageAddedListener: null
			_stageUpdatedListener: null
			_stageOutputTypesListener: null

			_stagesRetrievedHandler: (error, stagesToAdd) =>
				stagesToAdd = stagesToAdd.filter (stage) => return not @_stagesCache[stage.id]?

				@_stagesCache[stage.id] = stage for stage in stagesToAdd
				@_stages = @_stages.concat stagesToAdd
				@_gettingStages = false

			_handleStageAdded: (data) =>
				return if data.resourceId isnt @_changeId
				if not @_stagesCache[data.id]?
					@_stagesCache[data.id] = data
					@_stages.push data

			_handleStageUpdated: (data) =>
				return if data.resourceId isnt @_changeId
				stage = @_stagesCache[data.id]
				$.extend true, stage, data if stage?

			_handleStageOutputTypeAdded: (data) =>
				return if data.resourceId isnt @_changeId
				stage = @_stagesCache[data.id]

				if stage? and not (data.outputType in stage.outputTypes)
					stage.outputTypes.push data.outputType

			setChangeId: (changeId) =>
				assert.ok not changeId? or typeof changeId is 'number'

				if @_changeId isnt changeId
					@_stages = []
					@_stagesCache = {}
					@_changeId = changeId
					@stopListeningToEvents()

			retrieveStages: () =>
				assert.ok @_changeId?

				@_stages = []
				@_stagesCache = {}
				@_gettingStages = true

				rpc 'buildConsoles', 'read', 'getBuildConsoles', changeId: @_changeId, @_stagesRetrievedHandler

			getStages: () =>
				return @_stages

			isRetrievingStages: () =>
				return @_gettingStages

			listenToEvents: () =>
				assert.ok @_changeId?

				@stopListeningToEvents()

				@_stageAddedListener = events('changes', 'new build console', @_changeId).setCallback(@_handleStageAdded).subscribe()
				@_stageUpdatedListener = events('changes', 'return code added', @_changeId).setCallback(@_handleStageUpdated).subscribe()
				@_stageOutputTypesListener = events('changes', 'output type added', @_changeId).setCallback(@_handleStageOutputTypeAdded).subscribe()
					
			stopListeningToEvents: () =>
				@_stageAddedListener.unsubscribe() if @_stageAddedListener?
				@_stageUpdatedListener.unsubscribe() if @_stageUpdatedListener?
				@_stageOutputTypesListener.unsubscribe() if @_stageOutputTypesListener

				@_stageAddedListener = null
				@_stageUpdatedListener = null
				@_stageOutputTypesListener = null

		return create: () ->
			return new StagesManager()
	]).
	# factory('ConsoleTextManager', ['$timeout', 'ConsoleTextRpc', 'events', 'stringHasher', 'integerConverter', ($timeout, ConsoleTextRpc, events, stringHasher, integerConverter) ->
	factory('ConsoleTextManager', ['$rootScope', '$timeout', 'ConsoleTextRpc', 'events', 'stringHasher', 'integerConverter', ($rootScope, $timeout, ConsoleTextRpc, events, stringHasher, integerConverter) ->
		class ConsoleTextManager
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
				return if @_currentRequestId isnt linesData.requestId

				@_processNewLines linesData.lines
				@_gettingMoreLines = false

				@_allowGettingMoreLines = false
				$timeout (() => @_allowGettingMoreLines = true), 100

			_handleLinesAdded: (data) =>
				return if data.resourceId isnt @_stageId
				@_processNewLines data.lines

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
				assert.ok @_stageId?

				@stopListeningToEvents()
				@_linesAddedListener = events('buildConsoles', 'new output', @_stageId).setCallback(@_handleLinesAdded).subscribe()
					
			stopListeningToEvents: () =>
				@_linesAddedListener.unsubscribe() if @_linesAddedListener?
				@_linesAddedListener = null

		return create: () ->
			return new ConsoleTextManager()
	])
