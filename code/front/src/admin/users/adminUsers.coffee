'use strict'

window.AdminUsers = ['$scope', '$http', 'events', 'notification', ($scope, $http, events, notification) ->
	$scope.orderByPredicate = 'privilege'
	$scope.orderByReverse = false

	$scope.currentlyEditingUserId = null
	$scope.currentlyOpenDrawer = null

	$scope.addUsers =
		makingRequest: false

	getDomainName = () ->
		request = $http.get "/app/settings/domainName"
		request.success (data, status, headers, config) =>
			$scope.addUsers.domainName = data
		request.error (data, status, headers, config) =>
			notification.error data

	getAuthenticationSettings = () ->
		request = $http.get "/app/settings/authentication"
		request.success (data, status, headers, config) =>
			$scope.addUsers.connectionType = if data.googleAccountsAllowed then 'google' else 'default'
			$scope.addUsers.newConnectionType = $scope.addUsers.connectionType
			$scope.addUsers.emailDomains = data.allowedDomains?.join ' '
			$scope.addUsers.newEmailDomains = $scope.addUsers.emailDomains
		request.error (data, status, headers, config) =>
			notification.error data

	addUserPrivilege = (user) ->
		user.privilege = if user.isAdmin then 'Admin' else 'User'
		user.newPrivilege = user.privilege  # used while in edit mode
		return user

	getUsers = () ->
		request = $http.get "/app/users/"
		request.success (data, status, headers, config) =>
			$scope.users = (addUserPrivilege user for user in data)
		request.error (data, status, headers, config) =>
			notification.error data

	# handleConnectionTypesUpdated = (data) ->
	# 	$scope.addUsers.connectionType = data[0]
	# 	$scope.addUsers.newConnectionType = data[0]

	# handleEmailDomainsUpdated = (data) ->
	# 	$scope.addUsers.emailDomains = data.join ' '
	# 	$scope.addUsers.newEmailDomains = data.join ' '

	# handleUserAdded = (data) ->
	# 	$scope.users.push addUserPrivilege data

	# handleUserRemoved = (data) ->
	# 	userToRemoveIndex = (index for user, index in $scope.users when user.id is data.id)[0]
	# 	$scope.users.splice userToRemoveIndex, 1 if userToRemoveIndex?

	# handleUserAdminStatusChanged = (data) ->
	# 	userToUpdate = (user for user in $scope.users when user.id is data.id)[0]
	# 	privilege = if data.isAdmin then 'Admin' else 'User'
	# 	userToUpdate.privilege = privilege if userToUpdate?
	# 	userToUpdate.newPrivilege = privilege if userToUpdate?

	# allowedConnectionTypesEvents = events('systemSettings', 'allowed connection types updated', null).setCallback(handleConnectionTypesUpdated).subscribe()
	# allowedEmailDomainsEvents = events('systemSettings', 'allowed email domains updated', null).setCallback(handleEmailDomainsUpdated).subscribe()
	# addUserEvents = events('users', 'user created', initialState.user.id).setCallback(handleUserAdded).subscribe()
	# removeUserEvents = events('users', 'user removed', initialState.user.id).setCallback(handleUserRemoved).subscribe()
	# adminStatusEvents = events('users', 'user admin status', initialState.user.id).setCallback(handleUserAdminStatusChanged).subscribe()
	# $scope.$on '$destroy', allowedConnectionTypesEvents.unsubscribe
	# $scope.$on '$destroy', allowedEmailDomainsEvents.unsubscribe
	# $scope.$on '$destroy', addUserEvents.unsubscribe
	# $scope.$on '$destroy', removeUserEvents.unsubscribe
	# $scope.$on '$destroy', adminStatusEvents.unsubscribe

	getUsers()
	getAuthenticationSettings()
	getDomainName()

	$scope.toggleDrawer = (drawerName) ->
		if $scope.currentlyOpenDrawer is drawerName
			$scope.currentlyOpenDrawer = null
		else
			$scope.currentlyOpenDrawer = drawerName
			$scope.currentlyEditingUserId = null

	$scope.editUser = (user) ->
		$scope.currentlyEditingUserId = user?.id

	$scope.saveUser = (user) ->
		privilege = user.newPrivilege is 'Admin'
		request = $http.put "/app/users/#{user.id}/admin", privilege.toString()
		request.success (data, status, headers, config) =>
			$scope.currentlyEditingUserId = null
			user.privilege = user.newPrivilege
			notification.success "Admin status changed for: #{user.firstName} #{user.lastName}"
		request.error (data, status, headers, config) =>
			$scope.currentlyEditingUserId = null
			notification.error data

	$scope.deleteUser = (user) ->
		request = $http.delete "/app/users/#{user.id}"
		request.success (data, status, headers, config) =>
			$scope.currentlyEditingUserId = null
			notification.success "Deleted user #{user.firstName} #{user.lastName}"
		request.error (data, status, headers, config) =>
			$scope.currentlyEditingUserId = null
			notification.error data

	# $scope.saveAddUsersConfig = () ->
	# 	return if $scope.addUsers.makingRequest
	# 	$scope.addUsers.makingRequest = true

	# 	emailDomains = []
	# 	if $scope.addUsers.newEmailDomains isnt ''
	# 		emailDomains = $scope.addUsers.newEmailDomains.split(/[,; ]/)
	# 		emailDomains = emailDomains.filter (domain) -> return domain isnt ''

	# 	await
	# 		rpc 'systemSettings', 'update', 'setAllowedUserConnectionTypes', connectionTypes: [$scope.addUsers.newConnectionType], defer connectionTypeError
	# 		rpc 'systemSettings', 'update', 'setAllowedUserEmailDomains', emailDomains: emailDomains, defer emailDomainsError

	# 	$scope.addUsers.makingRequest = false

	# 	if connectionTypeError then notification.error connectionTypeError
	# 	else if emailDomainsError then notification.error emailDomainsError
	# 	else
	# 		$scope.addUsers.connectionType = $scope.addUsers.newConnectionType
	# 		$scope.addUsers.emailDomains = $scope.addUsers.newEmailDomains
	# 		notification.success 'Updated new user configuration'
	# 		$scope.clearAddUserConfig()

	$scope.clearAddUserConfig = () ->
		$scope.addUsers.newConnectionType = $scope.addUsers.connectionType
		$scope.addUsers.newEmailDomains = $scope.addUsers.emailDomains
		$scope.currentlyOpenDrawer = null
]
