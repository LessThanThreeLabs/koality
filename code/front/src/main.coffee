'use strict'


window.Main = ['$scope', '$timeout', 'rpc', 'events', 'initialState', 'notification', ($scope, $timeout, rpc, events, initialState, notification) ->
	repositories = []

	checkSshKeyExists = () ->
		rpc 'users', 'read', 'getSshKeys', null, (error, sshKeys) ->
			if sshKeys.length is 0
				notification.warning 'You need to set up an SSH Key. <a href="/account?view=sshKeys">Click here to add one</a>', 60

	getRepositories = () ->
		rpc 'repositories', 'read', 'getRepositories', null, (error, repos) ->
			repositories = repos
			updateChangeFinishedListeners()

	createChangeFinishedHandler = (repository) ->
		return (data) ->
			if data.user.id is initialState.user.id
				message = "<a href='/repository/#{repository.id}?change=#{data.id}'>Repository #{repository.name} - Change #{data.headCommit.sha.substring(0, 4)}</a> #{data.aggregateStatus}"
				if data.aggregateStatus is 'passed' then notification.success message
				else if data.aggregateStatus is 'failed' then notification.error message
				else if data.aggregateStatus is 'skipped' then notification.warning message

	changeFinishedListeners = []
	updateChangeFinishedListeners = () ->
		changeFinishedListener.unsubscribe() for changeFinishedListener in changeFinishedListeners
		changeFinishedListeners = []

		return if not initialState.loggedIn

		for repository in repositories
			changeFinishedListener = events('repositories', 'change finished', repository.id).setCallback(createChangeFinishedHandler(repository)).subscribe()
			changeFinishedListeners.push changeFinishedListener
	$scope.$on '$destroy', () -> changeFinishedListener.unsubscribe() for changeFinishedListener in changeFinishedListeners

	checkLicenseStatus = () ->
		billingUpdateUrl = 'https://koalitycode.com/account/payment'

		getDaysRemaining = (futureTime) ->
			timeInDay = 1000 * 60 * 60 * 24
			return Math.floor (futureTime - Date.now()) / timeInDay

		handleTrialExpirationWithoutPaymentInfo = (trialDaysRemaining) ->
			if trialDaysRemaining <= 0
				notification.error "There are 0 days left on your free trial. <a href='#{billingUpdateUrl}'>Upgrade now to continue using Koality.</a>", 0
			else
				message = "There are #{trialDaysRemaining} days left on your free trial. <a href='#{billingUpdateUrl}'>Upgrade now</a> to continue using Koality uninterrupted."
				if trialDaysRemaining <= 7 then notification.warning message
				else notification.success message

		handleTrialExpirationWithPaymentInfo = (trialDaysRemaining) ->
			if trialDaysRemaining <= 0
				notification.success "There are 0 days left on your free trial. Your subscription will begin tomorrow."

		handleUnpaidExpiration = (unpaidDaysRemaining) ->
			if unpaidDaysRemaining <= 0
				notification.error "Your payment method has expired. <a href='#{billingUpdateUrl}'>Update your billing information.</a>", 0
			else if unpaidDaysRemaining <= 15
				notification.warning "Your payment method is about to expire. <a href='#{billingUpdateUrl}'>Update your billing information.</a>"

		rpc 'systemSettings', 'read', 'getLicenseInformation', null, (error, licenseInformation) ->
			if licenseInformation.trialExpirationTime?
				if licenseInformation.unpaidExpirationTime?
					handleTrialExpirationWithoutPaymentInfo getDaysRemaining licenseInformation.trialExpirationTime
				else
					handleTrialExpirationWithPaymentInfo getDaysRemaining licenseInformation.trialExpirationTime
			else if licenseInformation.unpaidExpirationTime?
				handleUnpaidExpiration getDaysRemaining licenseInformation.unpaidExpirationTime
			else if not licenseInformation.active
				notification.error "Your license has been deactivated. <a href='#{billingUpdateUrl}'>Update your billing information.</a>", 0

	if initialState.loggedIn
		getRepositories()

		$timeout (() ->
			checkSshKeyExists()
			checkLicenseStatus()
		), 5000
]
