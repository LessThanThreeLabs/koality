'use strict'

angular.module('koality.service.repository', []).
	factory('currentRepository', ['rpc', 'integerConverter', (rpc, integerConverter) ->

		class RepositoryManager
			_id: null
			_information: null

			_forwardUrlListener: null

			clear: () =>
				@_id = null
				@_information = null

			setId: (repositoryId) =>
				@_id = integerConverter.toInteger repositoryId
				@_information = null

			getId: () =>
				return @_id

			# We ONLY listen to these evets when retrieving information, not when it is set.
			# If the information is set, it is coming from a source that is managing the
			# data and keeping it up to date.
			_handleForwardUrlUpdated: (data) =>
				return if data.id isnt @_id
				$.extend true, @_information, data

			_listenToEvents: () =>
				assert.ok @_id?

				@_stopListeningToEvents()
				@_forwardUrlListener = events('repositories', 'forward url updated', @_id).setCallback(@_handleForwardUrlUpdated).subscribe()
			
			_stopListeningToEvents: () =>
				@_forwardUrlListener.unsubscribe() if @_forwardUrlListener?
				@_forwardUrlListener = null

			setInformation: (repositoryInformation) =>
				assert.ok @_id?
				assert.ok repositoryInformation?
				@_information = repositoryInformation

			retrieveInformation: () =>
				assert.ok @_id
				rpc 'repositories', 'read', 'getMetadata', id: @_id, (error, repositoryInformation) =>
					@_information = repositoryInformation

			getInformation: () =>
				return @_information

		return new RepositoryManager()
	]).
	factory('currentChange', ['rpc', 'events', 'integerConverter', (rpc, events, integerConverter) ->

		class ChangeManager
			_repositoryId: null
			_id: null
			_information: null

			_startedListener: null
			_finishedListener: null

			clear: () =>
				@_repositoryId = null
				@_id = null
				@_information = null

			setId: (repositoryId, changeId) =>
				@_repositoryId = integerConverter.toInteger repositoryId
				@_id = integerConverter.toInteger changeId
				@_information = null

			getId: () =>
				return @_id

			# We ONLY listen to these evets when retrieving information, not when it is set.
			# If the information is set, it is coming from a source that is managing the
			# data and keeping it up to date.
			_handleChangeStarted: (data) =>
				return if data.id isnt @_id
				$.extend true, @_information, data

			_handleChangeFinished: (data) =>
				return if data.id isnt @_id
				$.extend true, @_information, data

			_listenToEvents: () =>
				assert.ok @_repositoryId?
				assert.ok @_id?

				@_stopListeningToEvents()
				@_startedListener = events('repositories', 'change started', @_repositoryId).setCallback(@_handleChangeStarted).subscribe()
				@_finishedListener = events('repositories', 'change finished', @_repositoryId).setCallback(@_handleChangeFinished).subscribe()
			
			_stopListeningToEvents: () =>
				@_startedListener.unsubscribe() if @_startedListener?
				@_finishedListener.unsubscribe() if @_finishedListener?

				@_startedListener = null
				@_finishedListener = null

			setInformation: (changeInformation) =>
				assert.ok @_repositoryId?
				assert.ok @_id?
				assert.ok changeInformation?
				@_stopListeningToEvents()
				@_information = changeInformation

			retrieveInformation: () =>
				assert.ok @_repositoryId?
				assert.ok @_id?

				requestData =
					repositoryId: @_repositoryId
					id: @_id
				rpc 'changes', 'read', 'getChange', requestData, (error, changeInformation) =>
					@_information = changeInformation
					@_listenToEvents()

			getInformation: () =>
				return @_information

		return new ChangeManager()
	]).
	factory('currentStage', ['rpc', 'events', 'integerConverter', (rpc, events, integerConverter) ->

		class StageManager
			_repositoryId: null
			_changeId: null
			_id: null
			_information: null
			_summary: false
			_skipped: false
			_merge: false
			_debug: false

			_updatedListener: null
			_outputTypesListener: null

			clear: () =>
				@_repositoryId = null
				@_id = null
				@_information = null

			setId: (repositoryId, changeId, stageId) =>
				@_repositoryId = integerConverter.toInteger repositoryId
				@_changeId = integerConverter.toInteger changeId
				@_id = integerConverter.toInteger stageId
				@_information = null
				@_summary = false
				@_skipped = false
				@_merge = false
				@_debug = false

			getId: () =>
				return @_id

			# We ONLY listen to these evets when retrieving information, not when it is set.
			# If the information is set, it is coming from a source that is managing the
			# data and keeping it up to date.
			_handleUpdated: (data) =>
				return if data.id isnt @_id
				$.extend true, @_information, data

			_handleOutputTypeAdded: (data) =>
				return if data.id isnt @_id

				if not (data.outputType in @_information.outputTypes)
					@_information.outputTypes.push data.outputType

			_listenToEvents: () =>
				assert.ok @_repositoryId?
				assert.ok @_changeId?
				assert.ok @_id?

				@_stopListeningToEvents()
				@_updatedListener = events('changes', 'return code added', @_changeId).setCallback(@_handleUpdated).subscribe()
				@_outputTypesListener = events('changes', 'output type added', @_changeId).setCallback(@_handleOutputTypeAdded).subscribe()
			
			_stopListeningToEvents: () =>
				@_updatedListener.unsubscribe() if @_updatedListener?
				@_outputTypesListener.unsubscribe() if @_outputTypesListener?

				@_updatedListener = null
				@_outputTypesListener = null

			setInformation: (stageInformation) =>
				assert.ok @_repositoryId?
				assert.ok @_changeId?
				assert.ok @_id?
				assert.ok stageInformation?
				@_stopListeningToEvents()
				@_information = stageInformation

			retrieveInformation: () =>
				assert.ok @_repositoryId?
				assert.ok @_changeId?
				assert.ok @_id?

				requestData =
					repositoryId: @_repositoryId
					id: @_id
				rpc 'buildConsoles', 'read', 'getBuildConsole', requestData, (error, stageInformation) =>
					@_information = stageInformation
					@_listenToEvents()

			getInformation: () =>
				return @_information

			setSummary: () =>
				@_id = null
				@_information = null
				@_summary = true
				@_skipped = false
				@_merge = false
				@_debug = false

			isSummary: () =>
				return @_summary

			setSkipped: () =>
				@_id = null
				@_information = null
				@_summary = false
				@_skipped = true
				@_merge = false
				@_debug = false

			isSkipped: () =>
				return @_skipped

			setMerge: () =>
				@_id = null
				@_information = null
				@_summary = false
				@_skipped = false
				@_merge = true
				@_debug = false

			isMerge: () =>
				return @_merge

			setDebug: () =>
				@_id = null
				@_information = null
				@_summary = false
				@_skipped = false
				@_merge = false
				@_debug = true

			isDebug: () =>
				return @_debug

		return new StageManager()
	])