'use strict'

window.AdminRepositories = ['$scope', '$location', '$routeParams', '$timeout', 'initialState', 'rpc', 'events', 'notification', ($scope, $location, $routeParams, $timeout, initialState, rpc, events, notification) ->
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
		repository.newForwardUrl = repository.forwardUrl
		repository.build.newPre = repository.build.pre
		repository.build.newPush = repository.build.push
		repository.build.newPullRequest = repository.build.pullRequest
		return repository

	getRepositories = () ->
		rpc 'repositories', 'read', 'getRepositories', null, (error, repositories) ->
			if error? then notification.error error
			else
				$scope.repositories = (addRepositoryEditFields repository for repository in repositories)
				updateRepositoryForwardUrlUpdatedListeners()
				getMaxRepositoryCount()

	getMaxRepositoryCount = () ->
		rpc 'systemSettings', 'read', 'getMaxRepositoryCount', null, (error, maxRepositoryCount) ->
			if error? then notification.error error
			else
				$scope.maxRepositoryCount = maxRepositoryCount ? Number.POSITIVE_INFINITY
				updateRepositoryCountExceeded()

	updateRepositoryCountExceeded = () ->
		return if not $scope.maxRepositoryCount?
		$scope.exceededMaxRepositoryCount = $scope.repositories.length >= $scope.maxRepositoryCount
		showRepositoriesLimitWarningIfNecessary()

	showRepositoriesLimitWarningIfNecessary = () ->
		upgradeUrl = 'https://koalitycode.com/account/plan'
		if $routeParams.view is 'repositories' and $scope.repositories.length >= $scope.maxRepositoryCount
			notification.warning "Max number of repositories reached. <a href='#{upgradeUrl}'>Upgrade to increase this limit</a>"

	getPublicKey = () ->
		rpc 'repositories', 'read', 'getPublicKey', null, (error, publicKey) ->
			if error? then notification.error error
			else $scope.publicKey.key = publicKey

	getIsConnectedToGitHub = () ->
		$scope.retrievingGitHubInformation = true
		rpc 'users', 'read', 'isConnectedToGitHub', null, (error, connectedToGitHub) ->
			$scope.retrievingGitHubInformation = false
			if error? then notification.error error
			else
				$scope.isConnectedToGitHub = connectedToGitHub

				if connectedToGitHub and $scope.addRepository.setupType is 'gitHub'
					getGitHubRepositories() 

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

	handleAddedRepositoryUpdate = (data) ->
		return if data.resourceId isnt initialState.user.id

		$scope.repositories ?= []
		repositoryExists = (repository for repository in $scope.repositories when repository.id is data.id).length isnt 0
		$scope.repositories.push addRepositoryEditFields(data) if not repositoryExists

		updateRepositoryCountExceeded()

	handleRemovedRepositoryUpdate = (data) ->
		return if data.resourceId isnt initialState.user.id

		repositoryToRemoveIndex = (index for repository, index in $scope.repositories when repository.id is data.id)[0]
		$scope.repositories.splice repositoryToRemoveIndex, 1 if repositoryToRemoveIndex?

		updateRepositoryCountExceeded()

	createRepositoryForwardUrlUpdateHandler = (repository) ->
		return (data) ->
			repository.forwardUrl = data.forwardUrl
			repository.newForwardUrl = data.forwardUrl

	repositoryForwardUrlUpdatedListeners = []
	updateRepositoryForwardUrlUpdatedListeners = () ->
		repositoryForwardUrlUpdatedListener.unsubscribe() for repositoryForwardUrlUpdatedListener in repositoryForwardUrlUpdatedListeners
		repositoryForwardUrlUpdatedListeners = []

		for repository in $scope.repositories
			repositoryForwardUrlUpdatedListener = events('repositories', 'forward url updated', repository.id).setCallback(createRepositoryForwardUrlUpdateHandler(repository)).subscribe()
			repositoryForwardUrlUpdatedListeners.push repositoryForwardUrlUpdatedListener
	$scope.$on '$destroy', () -> repositoryForwardUrlUpdatedListener.unsubscribe() for repositoryForwardUrlUpdatedListener in repositoryForwardUrlUpdatedListeners

	addRepositoryEvents = events('users', 'repository added', initialState.user.id).setCallback(handleAddedRepositoryUpdate).subscribe()
	removeRepositoryEvents = events('users', 'repository removed', initialState.user.id).setCallback(handleRemovedRepositoryUpdate).subscribe()
	$scope.$on '$destroy', addRepositoryEvents.unsubscribe
	$scope.$on '$destroy', removeRepositoryEvents.unsubscribe

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
			requestParams =
				id: repository.id
				forwardUrl: repository.newForwardUrl
			rpc 'repositories', 'update', 'setForwardUrl', requestParams, (error) ->
				if error? then callback error, false
				else
					repository.forwardUrl = repository.newForwardUrl
					callback null, true

		updateBuildHook = (callback) ->
			requestParams =
				id: repository.id
				pushEnabled: repository.build.newPush
				pullRequestEnabled: repository.build.newPullRequest
			rpc 'repositories', 'update', 'setGitHubBuildHook', requestParams, (error) ->
				if error? then callback error, false
				else
					repository.build.push = repository.build.newPush
					repository.build.pullRequest = repository.build.newPullRequest
					callback null, true

		return if repository.saving
		repository.saving = true

		await
			if repository.forwardUrl isnt repository.newForwardUrl
				updateForwardUrl defer forwardUrlError, forwardUrlSuccess

			if repository.build.push isnt repository.build.newPost or 
				repository.build.pullRequest isnt repository.build.newPullRequest
					updateBuildHook defer buildHookError, buildHookSuccess

		repository.saving = false
		$scope.currentlyEditingRepositoryId = null

		if forwardUrlError? then notification.error forwardUrlError
		else if buildHookError?
			if buildHookError.redirect? then window.location.href = buildHookError.redirect
			else notification.error buildHookError
		else if forwardUrlSuccess or buildHookSuccess
			notification.success "Repository #{repository.name} successfully updated"


	$scope.deleteRepository = (repository) ->
		if not repository.password? or repository.password is ''
			notification.error 'You must provide a password to delete repository: ' + repository.name
			return

		return if repository.makingDeleteRequest
		repository.makingDeleteRequest = true

		requestParams =
			id: repository.id
			password: repository.password
		rpc 'repositories', 'delete', 'deleteRepository', requestParams, (error) ->
			repository.makingDeleteRequest = false
			if error? then notification.error error
			else notification.success 'Successfully deleted repository: ' + repository.name

	$scope.createManualRepository = () ->
		return if $scope.addRepository.manual.makingRequest
		$scope.addRepository.manual.makingRequest = true

		rpc 'repositories', 'create', 'createRepository', $scope.addRepository.manual, false, (error, repositoryId) ->
			$scope.addRepository.manual.makingRequest = false
			if error? then notification.error error
			else
				notification.success 'Created repository ' + $scope.addRepository.manual.name, 15
				$scope.clearAddRepository()

	$scope.createGitHubRepository = () ->
		return if $scope.addRepository.gitHub.makingRequest
		$scope.addRepository.gitHub.makingRequest = true

		rpc 'repositories', 'create', 'createGitHubRepository', $scope.addRepository.gitHub.repository, false, (error, repositoryInformation) ->
			$scope.addRepository.gitHub.makingRequest = false
			if error? then notification.error error
			else
				successString = 'Created repository ' + $scope.addRepository.gitHub.repository.name
				if not repositoryInformation.keyAlreadyAdded
					successString += '. A Koality SSH Key has been added to your account'
				notification.success successString, 15

				$scope.clearAddRepository()

	$scope.clearAddRepository = () ->
		$scope.addRepository.setupType = 'manual'
		$scope.addRepository.manual = {}
		$scope.addRepository.gitHub = {}
		$scope.currentlyOpenDrawer = null

	$scope.$watch 'addRepository.setupType', () ->
		if $scope.addRepository.setupType is 'gitHub'
			getGitHubRepositories()
]
