'use strict'

window.AdminUsers = ['$scope', 'initialState', 'rpc', 'events', 'notification', ($scope, initialState, rpc, events, notification) ->
	$scope.orderByPredicate = 'privilege'
	$scope.orderByReverse = false

	$scope.userId = initialState.user?.id
	$scope.currentlyEditingUserId = null
	$scope.currentlyOpenDrawer = null

	$scope.addUsers =
		makingRequest: false

	getDomainName = () ->
		rpc 'systemSettings', 'read', 'getWebsiteSettings', null, (error, websiteSettings) ->
			$scope.addUsers.domainName = websiteSettings.domainName

	getAllowedConnectionTypes = () ->
		rpc 'systemSettings', 'read', 'getAllowedUserConnectionTypes', null, (error, allowedConnectionTypes) ->
			$scope.addUsers.connectionType = allowedConnectionTypes[0]
			$scope.addUsers.newConnectionType = allowedConnectionTypes[0]

	getAllowedEmailDomains = () ->
		rpc 'systemSettings', 'read', 'getAllowedUserEmailDomains', null, (error, allowedEmailDomains) ->
			$scope.addUsers.emailDomains = allowedEmailDomains.join ' '
			$scope.addUsers.newEmailDomains = allowedEmailDomains.join ' '

	addUserPrivilege = (user) ->
		user.privilege = if user.isAdmin then 'Admin' else 'User'
		user.newPrivilege = user.privilege  # used while in edit mode
		return user

	getUsers = () ->
		rpc 'users', 'read', 'getAllUsers', null, (error, users) ->
			$scope.users = (addUserPrivilege user for user in users)

	handleConnectionTypesUpdated = (data) ->
		$scope.addUsers.connectionType = data[0]
		$scope.addUsers.newConnectionType = data[0]

	handleEmailDomainsUpdated = (data) ->
		$scope.addUsers.emailDomains = data.join ' '
		$scope.addUsers.newEmailDomains = data.join ' '

	handleUserAdded = (data) ->
		$scope.users.push addUserPrivilege data

	handleUserRemoved = (data) ->
		userToRemoveIndex = (index for user, index in $scope.users when user.id is data.id)[0]
		$scope.users.splice userToRemoveIndex, 1 if userToRemoveIndex?

	handleUserAdminStatusChanged = (data) ->
		userToUpdate = (user for user in $scope.users when user.id is data.id)[0]
		privilege = if data.isAdmin then 'Admin' else 'User'
		userToUpdate.privilege = privilege if userToUpdate?
		userToUpdate.newPrivilege = privilege if userToUpdate?

	allowedConnectionTypesEvents = events('systemSettings', 'allowed connection types updated', null).setCallback(handleConnectionTypesUpdated).subscribe()
	allowedEmailDomainsEvents = events('systemSettings', 'allowed email domains updated', null).setCallback(handleEmailDomainsUpdated).subscribe()
	addUserEvents = events('users', 'user created', initialState.user.id).setCallback(handleUserAdded).subscribe()
	removeUserEvents = events('users', 'user removed', initialState.user.id).setCallback(handleUserRemoved).subscribe()
	adminStatusEvents = events('users', 'user admin status', initialState.user.id).setCallback(handleUserAdminStatusChanged).subscribe()
	$scope.$on '$destroy', allowedConnectionTypesEvents.unsubscribe
	$scope.$on '$destroy', allowedEmailDomainsEvents.unsubscribe
	$scope.$on '$destroy', addUserEvents.unsubscribe
	$scope.$on '$destroy', removeUserEvents.unsubscribe
	$scope.$on '$destroy', adminStatusEvents.unsubscribe

	getUsers()
	getDomainName()
	getAllowedConnectionTypes()
	getAllowedEmailDomains()

	$scope.toggleDrawer = (drawerName) ->
		if $scope.currentlyOpenDrawer is drawerName
			$scope.currentlyOpenDrawer = null
		else
			$scope.currentlyOpenDrawer = drawerName
			$scope.currentlyEditingUserId = null

	$scope.editUser = (user) ->
		$scope.currentlyEditingUserId = user?.id

	$scope.saveUser = (user) ->
		requestParams =
			id: user.id
			isAdmin: user.newPrivilege is 'Admin'
		rpc 'users', 'update', 'changeAdminStatus', requestParams, (error) ->
			$scope.currentlyEditingUserId = null
			if error? then notification.error error
			else 
				user.privilege = user.newPrivilege
				notification.success "Admin status changed for: #{user.firstName} #{user.lastName}"

	$scope.deleteUser = (user) ->
		rpc 'users', 'delete', 'deleteUser', id: user.id, (error) ->
			$scope.currentlyEditingUserId = null
			if error? then notification.error error
			else notification.success "Deleted user #{user.firstName} #{user.lastName}"

	$scope.saveAddUsersConfig = () ->
		return if $scope.addUsers.makingRequest
		$scope.addUsers.makingRequest = true

		emailDomains = []
		if $scope.addUsers.newEmailDomains isnt ''
			emailDomains = $scope.addUsers.newEmailDomains.split(/[,; ]/)
			emailDomains = emailDomains.filter (domain) -> return domain isnt ''

		await
			rpc 'systemSettings', 'update', 'setAllowedUserConnectionTypes', connectionTypes: [$scope.addUsers.newConnectionType], defer connectionTypeError
			rpc 'systemSettings', 'update', 'setAllowedUserEmailDomains', emailDomains: emailDomains, defer emailDomainsError

		$scope.addUsers.makingRequest = false

		if connectionTypeError then notification.error connectionTypeError
		else if emailDomainsError then notification.error emailDomainsError
		else
			$scope.addUsers.connectionType = $scope.addUsers.newConnectionType
			$scope.addUsers.emailDomains = $scope.addUsers.newEmailDomains
			notification.success 'Updated new user configuration'
			$scope.clearAddUserConfig()

	$scope.clearAddUserConfig = () ->
		$scope.addUsers.newConnectionType = $scope.addUsers.connectionType
		$scope.addUsers.newEmailDomains = $scope.addUsers.emailDomains
		$scope.currentlyOpenDrawer = null
]
