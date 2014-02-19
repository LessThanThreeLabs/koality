'use strict'

window.AdminUpgrade = ['$scope', '$http', '$timeout', 'events', 'notification', ($scope, $http, $timeout, events, notification) ->
	$scope.upgradeStatus = {}
	$scope.performingUpgrade = false
	$scope.makingRequest = false

	listenForWebserverComingBackUp = () ->
		intervalTime = 500

		checkIfWebserverIsDown = () ->
			request = $http.get '/ping', timeout: intervalTime
			request.success (data, status, headers, config) ->
				$timeout checkIfWebserverIsDown, intervalTime
			request.error (data, status, headers, config) ->
				$timeout checkIfWebserverIsUp, intervalTime
				console.log 'Webserver is down. Listening for webserver to come back up...'

		checkIfWebserverIsUp = () ->
			request = $http.get '/ping', timeout: intervalTime
			request.success (data, status, headers, config) ->
				notification.success 'Upgrade successful! Your browser will automatically refresh in 10 seconds', 10
				$timeout (() -> location.reload()), 10000
			request.error (data, status, headers, config) ->
				$timeout checkIfWebserverIsUp, intervalTime

		checkIfWebserverIsDown()

	getUpgradeStatus = () ->
		request = $http.get "/app/settings/upgradeStatus"
		request.success (data, status, headers, config) =>
			$scope.upgradeStatus = data
		request.error (data, status, headers, config) =>
			notification.error data

	# handleUpgradeStatus = (upgradeStatus) ->
	# 	return if not upgradeStatus?

	# 	$scope.version =
	# 		current: upgradeStatus.currentVersion
	# 		future: upgradeStatus.upgradeVersion

	# 	lastUpgradeStatus = upgradeStatus.lastUpgradeStatus
	# 	upgradeAvailable = upgradeStatus.upgradeAvailable ? false

	# 	if lastUpgradeStatus is 'running'
	# 		$scope.upgrade.message = 'An upgrade is currently in progress. You should expect the system to restart in a few minutes.'
	# 		$scope.upgrade.upgradeAllowed = false
	# 	else if lastUpgradeStatus is 'failed'
	# 		$scope.upgrade.message = 'The last upgrade failed. Contact support if this happens again.'
	# 		$scope.upgrade.upgradeAllowed = upgradeAvailable
	# 	else if upgradeAvailable
	# 		$scope.upgrade.message = 'An upgrade to Koality is available. Upgrading will shut down the server and may take several minutes before restarting.'
	# 		$scope.upgrade.upgradeAllowed = true
	# 	else
	# 		$scope.upgrade.message = 'There are no upgrades available at this time.'
	# 		$scope.upgrade.upgradeAllowed = false

	# handleSystemSettingsUpdate = (data) ->
	# 	if data.resource is 'deployment' and data.key is 'upgrade_status'
	# 		handleUpgradeStatus lastUpgradeStatus: data.value

	# changedSystemSetting = events('systemSettings', 'system setting updated', initialState.user.id).setCallback(handleSystemSettingsUpdate).subscribe()
	# $scope.$on '$destroy', changedSystemSetting.unsubscribe

	getUpgradeStatus()

	$scope.performUpgrade = () ->
		return if $scope.makingRequest
		$scope.makingRequest = true
		downloadingMessage = "Downloading upgrade to " + $scope.upgradeStatus.nextVersion + "..."
		$scope.upgradeStatus.statusMessage = downloadingMessage
		notification.success downloadingMessage

		request = $http.post "/app/settings/upgrade", $scope.upgradeStatus.nextVersion
		request.success (data, status, headers, config) =>
			$scope.makingRequest = false
			$scope.performingUpgrade = true

			upgradeMessage = "Upgrade to " + $scope.upgradeStatus.nextVersion + " in progress..."
			$scope.upgradeStatus.statusMessage = upgradeMessage
			notification.success upgradeMessage

			listenForWebserverComingBackUp()
		request.error (data, status, headers, config) =>
			$scope.makingRequest = false
			$scope.upgradeStatus.statusMessage = "Failed to download upgrade file."
			notification.error data
]
