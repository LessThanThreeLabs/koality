'use strict'

window.AdminPlan = ['$scope', 'rpc', 'notification', ($scope, rpc, notification) ->
	$scope.license = {}

	getLicenseKey = () ->
		rpc 'systemSettings', 'read', 'getLicenseKey', null, (error, licenseKey) ->
			$scope.license.key = licenseKey

	getLicenseInformation = () ->
		rpc 'systemSettings', 'read', 'getLicenseInformation', null, (error, licenseInformation) ->
			$scope.license.type = licenseInformation.type

	getNumUsers = () ->
		rpc 'users', 'read', 'getAllUsers', null, (error, users) ->
			$scope.license.numUsers = users.length

	getLicenseKey()
	getLicenseInformation()
	getNumUsers()
]
