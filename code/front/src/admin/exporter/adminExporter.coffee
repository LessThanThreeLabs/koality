'use strict'

window.AdminExporter = ['$scope', 'rpc', 'events', 'notification', ($scope, rpc, events, notification) ->
	$scope.exporter = {}
	$scope.makingRequest = false

	getS3BucketName = () ->
		rpc 'systemSettings', 'read', 'getS3BucketName', null, (error, bucketName) ->
			$scope.exporter.s3BucketName = bucketName

	handleBucketNameUpdated = (data) ->
		$scope.exporter.s3BucketName = data

	getS3BucketName()

	s3BucketNameEvents = events('systemSettings', 's3 bucket name updated', null).setCallback(handleBucketNameUpdated).subscribe()
	$scope.$on '$destroy', s3BucketNameEvents.unsubscribe

	$scope.submit = () ->
		return if $scope.makingRequest
		$scope.makingRequest = true

		rpc 'systemSettings', 'update', 'setS3BucketName', bucketName: $scope.exporter.s3BucketName, (error) ->
			$scope.makingRequest = false
			if error? then notification.error error
			else notification.success 'Updated exporter settings'
]
