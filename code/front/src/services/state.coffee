'use strict'

angular.module('koality.service.state', []).
	factory('currentRepository', ['$http', 'integerConverter', ($http, integerConverter) ->
		class RepositoryManager
			_id: null
			_information: null

			clear: () =>
				@_id = null
				@_information = null

			setId: (repositoryId) =>
				@_id = integerConverter.toInteger repositoryId
				@_information = null

			getId: () =>
				return @_id

			setInformation: (repositoryInformation) =>
				assert.ok @_id?
				assert.ok repositoryInformation?
				@_information = repositoryInformation

			retrieveInformation: () =>
				assert.ok @_id

				request = $http.get("/app/repositories/#{@_id}")
				request.success (data, status, headers, config) =>
					@_information = data
				request.error (data, status, headers, config) =>
					console.error data

			getInformation: () =>
				return @_information

		return new RepositoryManager()
	]).
	factory('currentBuild', ['$http', 'events', 'integerConverter', ($http, events, integerConverter) ->
		class BuildManager
			_repositoryId: null
			_id: null
			_information: null

			# _startedListener: null
			# _finishedListener: null

			clear: () =>
				@_repositoryId = null
				@_id = null
				@_information = null

			setId: (repositoryId, buildId) =>
				@_repositoryId = integerConverter.toInteger repositoryId
				@_id = integerConverter.toInteger buildId
				@_information = null

			getId: () =>
				return @_id

			# We ONLY listen to these events when retrieving information, not when it is set.
			# If the information is set, it is coming from a source that is managing the
			# data and keeping it up to date.
			# _handleChangeStarted: (data) =>
			# 	return if data.id isnt @_id
			# 	$.extend true, @_information, data

			# _handleChangeFinished: (data) =>
			# 	return if data.id isnt @_id
			# 	$.extend true, @_information, data

			# _listenToEvents: () =>
			# 	assert.ok @_repositoryId?
			# 	assert.ok @_id?

			# 	@_stopListeningToEvents()
			# 	@_startedListener = events('repositories', 'change started', @_repositoryId).setCallback(@_handleChangeStarted).subscribe()
			# 	@_finishedListener = events('repositories', 'change finished', @_repositoryId).setCallback(@_handleChangeFinished).subscribe()
			
			# _stopListeningToEvents: () =>
			# 	@_startedListener.unsubscribe() if @_startedListener?
			# 	@_finishedListener.unsubscribe() if @_finishedListener?

			# 	@_startedListener = null
			# 	@_finishedListener = null

			setInformation: (changeInformation) =>
				assert.ok @_repositoryId?
				assert.ok @_id?
				assert.ok changeInformation?
				# @_stopListeningToEvents()
				@_information = changeInformation

			retrieveInformation: () =>
				assert.ok @_repositoryId?
				assert.ok @_id?

				request = $http.get("/app/builds/#{@_id}")
				request.success (data, status, headers, config) =>
					@_information = data
					# @_listenToEvents()
				request.error (data, status, headers, config) =>
					console.error data

			getInformation: () =>
				return @_information

		return new BuildManager()
	]).
	factory('currentStage', ['$http', 'events', 'integerConverter', ($http, events, integerConverter) ->
		class StageManager
			_repositoryId: null
			_buildId: null
			_id: null
			_information: null
			_summary: false

			# _updatedListener: null
			# _outputTypesListener: null

			clear: () =>
				@_repositoryId = null
				@_buildId = null
				@_id = null
				@_information = null

			setId: (repositoryId, buildId, stageId) =>
				@_repositoryId = integerConverter.toInteger repositoryId
				@_buildId = integerConverter.toInteger buildId
				@_id = integerConverter.toInteger stageId
				@_information = null
				@_summary = false

			getId: () =>
				return @_id

			# We ONLY listen to these events when retrieving information, not when it is set.
			# If the information is set, it is coming from a source that is managing the
			# data and keeping it up to date.
			# _handleUpdated: (data) =>
			# 	return if data.id isnt @_id
			# 	$.extend true, @_information, data

			# _handleOutputTypeAdded: (data) =>
			# 	return if data.id isnt @_id

			# 	if not (data.outputType in @_information.outputTypes)
			# 		@_information.outputTypes.push data.outputType

			# _listenToEvents: () =>
			# 	assert.ok @_repositoryId?
			# 	assert.ok @_buildId?
			# 	assert.ok @_id?

			# 	@_stopListeningToEvents()
			# 	@_updatedListener = events('changes', 'return code added', @_buildId).setCallback(@_handleUpdated).subscribe()
			# 	@_outputTypesListener = events('changes', 'output type added', @_buildId).setCallback(@_handleOutputTypeAdded).subscribe()
			
			# _stopListeningToEvents: () =>
			# 	@_updatedListener.unsubscribe() if @_updatedListener?
			# 	@_outputTypesListener.unsubscribe() if @_outputTypesListener?

			# 	@_updatedListener = null
			# 	@_outputTypesListener = null

			setInformation: (stageInformation) =>
				assert.ok @_repositoryId?
				assert.ok @_buildId?
				assert.ok @_id?
				assert.ok stageInformation?
				# @_stopListeningToEvents()
				@_information = stageInformation

			retrieveInformation: () =>
				assert.ok @_repositoryId?
				assert.ok @_buildId?
				assert.ok @_id?

				request = $http.get("/app/stages/#{@_id}")
				request.success (data, status, headers, config) =>
					@_information = data
					# @_listenToEvents()
				request.error (data, status, headers, config) =>
					console.error data

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

		return new StageManager()
	]).
	factory('currentStageRun', ['$http', 'events', 'integerConverter', ($http, events, integerConverter) ->
		class StageManager
			_repositoryId: null
			_buildId: null
			_stageId: null
			_id: null
			_information: null

			# _updatedListener: null
			# _outputTypesListener: null

			clear: () =>
				@_repositoryId = null
				@_buildId = null
				@_stageId = null
				@_id = null
				@_information = null

			setId: (repositoryId, buildId, stageId, stageRunId) =>
				@_repositoryId = integerConverter.toInteger repositoryId
				@_buildId = integerConverter.toInteger buildId
				@_stageId = integerConverter.toInteger stageId
				@_id = integerConverter.toInteger stageRunId
				@_information = null

			getId: () =>
				return @_id

			# We ONLY listen to these events when retrieving information, not when it is set.
			# If the information is set, it is coming from a source that is managing the
			# data and keeping it up to date.
			# _handleUpdated: (data) =>
			# 	return if data.id isnt @_id
			# 	$.extend true, @_information, data

			# _handleOutputTypeAdded: (data) =>
			# 	return if data.id isnt @_id

			# 	if not (data.outputType in @_information.outputTypes)
			# 		@_information.outputTypes.push data.outputType

			# _listenToEvents: () =>
			# 	assert.ok @_repositoryId?
			# 	assert.ok @_buildId?
			# 	assert.ok @_id?

			# 	@_stopListeningToEvents()
			# 	@_updatedListener = events('changes', 'return code added', @_buildId).setCallback(@_handleUpdated).subscribe()
			# 	@_outputTypesListener = events('changes', 'output type added', @_buildId).setCallback(@_handleOutputTypeAdded).subscribe()
			
			# _stopListeningToEvents: () =>
			# 	@_updatedListener.unsubscribe() if @_updatedListener?
			# 	@_outputTypesListener.unsubscribe() if @_outputTypesListener?

			# 	@_updatedListener = null
			# 	@_outputTypesListener = null

			setInformation: (stageRunInformation) =>
				assert.ok @_repositoryId?
				assert.ok @_buildId?
				assert.ok @_stageId?
				assert.ok @_id?
				assert.ok stageRunInformation?
				# @_stopListeningToEvents()
				@_information = stageRunInformation

			retrieveInformation: () =>
				assert.ok @_repositoryId?
				assert.ok @_buildId?
				assert.ok @_stageId?
				assert.ok @_id?

				request = $http.get("/app/stageRuns/#{@_id}")
				request.success (data, status, headers, config) =>
					@_information = data
					# @_listenToEvents()
				request.error (data, status, headers, config) =>
					console.error data

			getInformation: () =>
				return @_information

		return new StageManager()
	])
