'use strict'

window.AdminExporter = ['$scope', '$http', 'events', 'notification', ($scope, $http, events, notification) ->
	$scope.s3Exporter = {}
	$scope.makingRequest = false

	getS3Exporter = () ->
		request = $http.get "/app/settings/s3Exporter"
		request.success (data, status, headers, config) =>
			$scope.s3Exporter.accessKey = data.accessKey
			$scope.s3Exporter.secretKey = data.secretKey
			$scope.s3Exporter.bucketName = data.bucketName
			$scope.s3Exporter.enabled = if data? and Object.keys(data).length > 0 then 'yes' else 'no'
		request.error (data, status, headers, config) =>
			notification.error data
			$scope.s3Exporter.enabled = 'no'

	# handleBucketNameUpdated = (data) ->
	# 	$scope.exporter.s3BucketName = data

	getS3Exporter()

	# s3BucketNameEvents = events('systemSettings', 's3 bucket name updated', null).setCallback(handleBucketNameUpdated).subscribe()
	# $scope.$on '$destroy', s3BucketNameEvents.unsubscribe

	$scope.submit = () ->
		return if $scope.makingRequest
		$scope.makingRequest = true

		request = null
		if $scope.s3Exporter.enabled is 'yes'
			request = $http.put "/app/settings/s3Exporter", $scope.s3Exporter
		else
			request = $http.delete "/app/settings/s3Exporter"

		request.success (data, status, headers, config) =>
			$scope.makingRequest = false	
			notification.success 'Updated exporter settings'
		request.error (data, status, headers, config) =>
			$scope.makingRequest = false
			notification.error data
]
