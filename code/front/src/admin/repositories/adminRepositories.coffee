'use strict'

window.AdminRepositories = ['$scope', '$location', '$routeParams', '$http', '$timeout', 'events', 'notification', ($scope, $location, $routeParams, $http, $timeout, events, notification) ->
	$scope.orderByPredicate = 'name'
	$scope.orderByReverse = false

	$scope.currentlyEditingRepositoryId = null
	$scope.currentlyOpenDrawer = null

	$scope.isConnectedToGitHub = false

	$scope.repositories = []

	$scope.addRepository =
		setupType: 'manual'
		manual: {}
		gitHub: {}

	$scope.publicKey =
		key: null

	if $routeParams.addGitHubRepository
		$location.search 'addGitHubRepository', null
		$timeout (() ->
			$scope.addRepository.setupType = 'gitHub'
			$scope.toggleDrawer 'addRepository'
		), 500

	addRepositoryEditFields = (repository) ->
		repository.newRemoteUri = repository.remoteUri
		if repository.gitHub?
			repository.gitHub.push = repository.gitHub?.hookTypes? and 'push' in repository.gitHub.hookTypes
			repository.gitHub.pullRequest = repository.gitHub?.hookTypes? and 'pull_request' in repository.gitHub.hookTypes
			repository.gitHub.newPush = repository.gitHub.push
			repository.gitHub.newPullRequest = repository.gitHub.pullRequest
		return repository

	getRepositories = () ->
		request = $http.get "/app/repositories/"
		request.success (data, status, headers, config) =>
			$scope.repositories = (addRepositoryEditFields repository for repository in data)
			# updateRepositoryForwardUrlUpdatedListeners()
		request.error (data, status, headers, config) =>
			notification.error data

	getPublicKey = () ->
		request = $http.get "/app/settings/repositoryKeyPair"
		request.success (data, status, headers, config) =>
			$scope.publicKey.key = data.publicKey
		request.error (data, status, headers, config) =>
			notification.error data

	getIsConnectedToGitHub = () ->
		console.log 'need to determine if connected to GitHub...'
		# $scope.retrievingGitHubInformation = true
		# rpc 'users', 'read', 'isConnectedToGitHub', null, (error, connectedToGitHub) ->
		# 	$scope.retrievingGitHubInformation = false
		# 	if error? then notification.error error
		# 	else
		# 		$scope.isConnectedToGitHub = connectedToGitHub

		# 		if connectedToGitHub and $scope.addRepository.setupType is 'gitHub'
		# 			getGitHubRepositories() 

	hasRequestedGitHubRepositories = false
	getGitHubRepositories = () ->
		return if not $scope.isConnectedToGitHub
		return if hasRequestedGitHubRepositories

		hasRequestedGitHubRepositories = true
		$scope.retrievingGitHubInformation = true
		rpc 'repositories', 'read', 'getGitHubRepositories', null, (error, gitHubRepositories) ->
			$scope.retrievingGitHubInformation = false
			if error? then notification.error error
			else
				$scope.gitHubRepositories = gitHubRepositories
				for repository in $scope.gitHubRepositories
					repository.displayName = "#{repository.owner}/#{repository.name}"

	# handleAddedRepositoryUpdate = (data) ->
	# 	return if data.resourceId isnt initialState.user.id

	# 	$scope.repositories ?= []
	# 	repositoryExists = (repository for repository in $scope.repositories when repository.id is data.id).length isnt 0
	# 	$scope.repositories.push addRepositoryEditFields(data) if not repositoryExists

	# 	updateRepositoryCountExceeded()

	# handleRemovedRepositoryUpdate = (data) ->
	# 	return if data.resourceId isnt initialState.user.id

	# 	repositoryToRemoveIndex = (index for repository, index in $scope.repositories when repository.id is data.id)[0]
	# 	$scope.repositories.splice repositoryToRemoveIndex, 1 if repositoryToRemoveIndex?

	# 	updateRepositoryCountExceeded()

	# createRepositoryForwardUrlUpdateHandler = (repository) ->
	# 	return (data) ->
	# 		repository.remoteUri = data.remoteUri
	# 		repository.newRemoteUri = data.remoteUri

	# repositoryForwardUrlUpdatedListeners = []
	# updateRepositoryForwardUrlUpdatedListeners = () ->
	# 	repositoryForwardUrlUpdatedListener.unsubscribe() for repositoryForwardUrlUpdatedListener in repositoryForwardUrlUpdatedListeners
	# 	repositoryForwardUrlUpdatedListeners = []

	# 	for repository in $scope.repositories
	# 		repositoryForwardUrlUpdatedListener = events('repositories', 'forward url updated', repository.id).setCallback(createRepositoryForwardUrlUpdateHandler(repository)).subscribe()
	# 		repositoryForwardUrlUpdatedListeners.push repositoryForwardUrlUpdatedListener
	# $scope.$on '$destroy', () -> repositoryForwardUrlUpdatedListener.unsubscribe() for repositoryForwardUrlUpdatedListener in repositoryForwardUrlUpdatedListeners

	# addRepositoryEvents = events('users', 'repository added', initialState.user.id).setCallback(handleAddedRepositoryUpdate).subscribe()
	# removeRepositoryEvents = events('users', 'repository removed', initialState.user.id).setCallback(handleRemovedRepositoryUpdate).subscribe()
	# $scope.$on '$destroy', addRepositoryEvents.unsubscribe
	# $scope.$on '$destroy', removeRepositoryEvents.unsubscribe

	getRepositories()
	getPublicKey()
	getIsConnectedToGitHub()

	$scope.toggleDrawer = (drawerName) ->
		if $scope.currentlyOpenDrawer is drawerName
			$scope.currentlyOpenDrawer = null
		else
			$scope.currentlyOpenDrawer = drawerName
			$scope.currentlyEditingRepositoryId = null

	$scope.connectToGitHub = () ->
		rpc 'repositories', 'read', 'getGitHubConnectRedirectUri', null, (error, redirectUri) ->
			if error? then notification.error error
			else window.location.href = redirectUri

	$scope.editRepository = (repository) ->
		otherRepository.deleting = false for otherRepository in $scope.repositories
		$scope.currentlyEditingRepositoryId = repository?.id

	$scope.saveRepository = (repository) ->
		updateForwardUrl = (callback) ->
			request = $http.put "/app/repositories/#{repository.id}/remoteUri", repository.newRemoteUri
			request.success (data, status, headers, config) =>
				repository.remoteUri = repository.newRemoteUri
				callback null, true
			request.error (data, status, headers, config) =>
				callback data, false

		updateGitHubHook = (callback) ->
			hookTypes = []
			hookTypes.append 'push' if repository.gitHub.newPush
			hookTypes.append 'pull_request' if repository.gitHub.newPullRequest

			if hookTypes.length is 0
				request = $http.put "/app/repositories/#{repository.id}/gitHub/clearHook"
				request.success (data, status, headers, config) =>
					repository.gitHub.push = repository.gitHub.newPush
					repository.gitHub.pullRequest = repository.gitHub.newPullRequest
					callback null, true
				request.error (data, status, headers, config) =>
					callback data, false
			else
				request = $http.put "/app/repositories/#{repository.id}/gitHub/setHook", hookTypes
				request.success (data, status, headers, config) =>
					repository.gitHub.push = repository.gitHub.newPush
					repository.gitHub.pullRequest = repository.gitHub.newPullRequest
					callback null, true
				request.error (data, status, headers, config) =>
					callback data, false

		return if repository.saving
		repository.saving = true

		await
			if repository.remoteUri isnt repository.newRemoteUri
				updateForwardUrl defer remoteUriError, remoteUriSuccess

			if repository.gitHub? and 
				(repository.gitHub.push isnt repository.gitHub.newPost or
				repository.gitHub.pullRequest isnt repository.gitHub.newPullRequest)
					updateGitHubHook defer gitHubHookError, gitHubHookSuccess

		repository.saving = false
		$scope.currentlyEditingRepositoryId = null

		if remoteUriError? then notification.error remoteUriError
		else if gitHubHookError?
			if gitHubHookError.redirect? then window.location.href = gitHubHookError.redirect
			else notification.error gitHubHookError
		else if remoteUriSuccess or gitHubHookSuccess
			notification.success "Repository #{repository.name} successfully updated"

	$scope.deleteRepository = (repository) ->
		if not repository.deleteName? or repository.deleteName is ''
			notification.error 'You must confirm with repository name to delete ' + repository.name
			return
		else if repository.name isnt repository.deleteName
			notification.error 'Incorrect repository name'
			return

		return if repository.makingDeleteRequest
		repository.makingDeleteRequest = true

		request = $http.delete "/app/repositories/#{repository.id}"
		request.success (data, status, headers, config) =>
			repository.makingDeleteRequest = false
			notification.success 'Successfully deleted repository: ' + repository.name
		request.error (data, status, headers, config) =>
			repository.makingDeleteRequest = false
			notification.error data

	$scope.createManualRepository = () ->
		return if $scope.addRepository.manual.makingRequest
		$scope.addRepository.manual.makingRequest = true

		request = $http.post "/app/repositories/create", $scope.addRepository.manual
		request.success (data, status, headers, config) =>
			$scope.addRepository.manual.makingRequest = false
			notification.success 'Created repository ' + $scope.addRepository.manual.name, 15
			$scope.clearAddRepository()
		request.error (data, status, headers, config) =>
			$scope.addRepository.manual.makingRequest = false
			notification.error data

	$scope.createGitHubRepository = () ->
		return if $scope.addRepository.gitHub.makingRequest
		$scope.addRepository.gitHub.makingRequest = true

		request = $http.post "/app/repositories/gitHub/create", $scope.addRepository.gitHub.repository
		request.success (data, status, headers, config) =>
			$scope.addRepository.gitHub.makingRequest = false
			notification.success "Created repository #{$scope.addRepository.gitHub.repository.name}. A Koality SSH Key has been added to your account", 15
			$scope.clearAddRepository()
		request.error (data, status, headers, config) =>
			$scope.addRepository.gitHub.makingRequest = false
			notification.error data

	$scope.clearAddRepository = () ->
		$scope.addRepository.setupType = 'manual'
		$scope.addRepository.manual = {}
		$scope.addRepository.gitHub = {}
		$scope.currentlyOpenDrawer = null

	$scope.$watch 'addRepository.setupType', () ->
		if $scope.addRepository.setupType is 'gitHub'
			getGitHubRepositories()
]
