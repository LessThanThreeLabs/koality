'use strict'

angular.module('koality.service.rpc', []).
	factory('BuildsRpc', ['$http', 'integerConverter', ($http, integerConverter) ->
		class BuildsRpc
			_requestIdCounter: 0
			_noMoreBuildsToRequest: false

			_currentQuery: null
			_currentCallback: null
			_nextQuery: null
			_nextCallback: null

			constructor: (@_numBuildsToRequest) ->
				assert.ok typeof @_numBuildsToRequest is 'number'

			_createBuildsQuery: (requestId, repositoryIds, group, query, startIndex) =>
				requestId: requestId
				repositoryIds: repositoryIds
				group: group
				query: query
				startIndex: startIndex
				numToRetrieve: @_numBuildsToRequest

			_shiftBuildsRequest: () =>
				if not @_nextQuery?
					@_currentQuery = null
					@_currentCallback = null
				else
					@_currentQuery = @_nextQuery
					@_currentCallback = @_nextCallback
					@_nextQuery = null
					@_nextCallback = null

					@_retrieveMoreBuilds()

			_parseBuildTimes: (builds) =>
				for build in builds
					build.created = new Date build.created if build.created?
					build.started = new Date build.started if build.started?
					build.ended = new Date build.ended if build.ended?
				return builds

			_retrieveMoreBuilds: () =>
				assert.ok @_currentQuery?
				assert.ok @_currentCallback?

				@_noMoreBuildsToRequest = false if @_currentQuery.startIndex is 0

				if @_noMoreBuildsToRequest
					@_shiftBuildsRequest()
				else
					repositoryId = @_currentQuery.repositoryIds[0]
					console.log 'SHOULD BE PASSING MULTIPLE REPOSITORY IDS, and what about group and query?'
					# repositoryIds = @_currentQuery.repositoryIds.join ','
					offset = @_currentQuery.startIndex
					results = @_currentQuery.numToRetrieve
					request = $http.get("/app/verifications/tail?repositoryId=#{repositoryId}&offset=#{offset}&results=#{results}")
					request.success (data, status, headers, config) =>
						@_noMoreBuildsToRequest = data.length < @_numBuildsToRequest
						@_currentCallback null,
							requestId: @_currentQuery.requestId
							builds: @_parseBuildTimes data
						@_shiftBuildsRequest()
					request.error (data, status, headers, config) =>
						@_currentCallback data
						@_shiftBuildsRequest()

			hasMoreBuildsToRequest: () =>
				return not @_noMoreBuildsToRequest

			queueRequest: (repositoryIds, group, query, startIndex, callback) =>
				assert.ok typeof repositoryIds is 'object' and repositoryIds.length > 0
				assert.ok not group? or (typeof group is 'string' and (group is 'all' or group is 'me'))
				assert.ok not query? or (typeof query is 'string')
				assert.ok (group? and not query?) or (not group? and query?)
				assert.ok typeof startIndex is 'number'
				assert.ok typeof callback is 'function'

				repositoryIds = repositoryIds.map (repositoryId) -> return integerConverter.toInteger repositoryId

				newQuery = @_createBuildsQuery @_requestIdCounter, repositoryIds, group, query, startIndex
				@_requestIdCounter++

				if @_currentQuery?
					@_nextQuery = newQuery
					@_nextCallback = callback
				else
					@_currentQuery = newQuery
					@_currentCallback = callback
					@_retrieveMoreBuilds()

				return newQuery.requestId

		return create: (numBuildsToRequest=100) ->
			return new BuildsRpc numBuildsToRequest
	]).
	factory('ConsoleLinesRpc', ['$http', ($http) ->
		class ConsoleLinesRpc
			_requestIdCounter: 0
			_noMoreLinesToRequest: false

			_currentQuery: null
			_currentCallback: null
			_nextQuery: null
			_nextCallback: null

			constructor: (@_numLinesToRequest) ->
				assert.ok typeof @_numLinesToRequest is 'number'

			_createLinesQuery: (requestId, stageRunId, startIndex) =>
				requestId: requestId
				id: stageRunId
				startIndex: startIndex
				numToRetrieve: @_numLinesToRequest

			_shiftConsoleLinesRequest: () =>
				if not @_nextQuery?
					@_currentQuery = null
					@_currentCallback = null
				else
					@_currentQuery = @_nextQuery
					@_currentCallback = @_nextCallback
					@_nextQuery = null
					@_nextCallback = null

					@_retrieveMoreConsoleLines()

			_retrieveMoreConsoleLines: () =>
				assert.ok @_currentQuery?
				assert.ok @_currentCallback?

				@_noMoreLinesToRequest = false if @_currentQuery.startIndex is 0

				if @_noMoreLinesToRequest
					@_shiftConsoleLinesRequest()
				else
					stageRunId = @_currentQuery.id
					offset = @_currentQuery.startIndex
					results = @_currentQuery.numToRetrieve
					request = $http.get("/app/stageRuns/#{stageRunId}/lines?offset=#{offset}&results=#{results}&from=tail")
					request.success (data, status, headers, config) =>
						@_noMoreLinesToRequest = Object.keys(data).length < @_numLinesToRequest
						@_currentCallback null,
							requestId: @_currentQuery.requestId
							lines: data
						@_shiftConsoleLinesRequest()
					request.error (data, status, headers, config) =>
						@_currentCallback data
						@_shiftConsoleLinesRequest()

			hasMoreLinesToRequest: () =>
				return not @_noMoreLinesToRequest

			notifyLinesRemoved: () =>
				@_noMoreLinesToRequest = false

			queueRequest: (stageRunId, startIndex, callback) =>
				assert.ok typeof stageRunId is 'number'
				assert.ok typeof startIndex is 'number'
				assert.ok typeof callback is 'function'

				newQuery = @_createLinesQuery @_requestIdCounter, stageRunId, startIndex
				@_requestIdCounter++

				if @_currentQuery?
					@_nextQuery = newQuery
					@_nextCallback = callback
				else
					@_currentQuery = newQuery
					@_currentCallback = callback
					@_retrieveMoreConsoleLines()

				return newQuery.requestId

		return create: (numLinesToRequest=1000) ->
			return new ConsoleLinesRpc numLinesToRequest
	])