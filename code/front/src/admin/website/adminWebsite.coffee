'use strict'

window.AdminWebsite = ['$scope', 'rpc', 'events', 'notification', ($scope, rpc, events, notification) ->
	$scope.domain = {}
	$scope.ssl = {}

	getWebsiteSettings = () ->
		rpc 'systemSettings', 'read', 'getWebsiteSettings', null, (error, websiteSettings) ->
			$scope.domain = websiteSettings

	submitDomainName = () ->
		rpc 'systemSettings', 'update', 'setWebsiteSettings', $scope.domain, (error) ->
			if error? then notification.error error
			else notification.success 'Updated website domain'

	submitSslCertificate = () ->
		rpc 'systemSettings', 'update', 'setSslCertificate', $scope.ssl, (error) ->
			if error? then notification.error error
			else notification.success 'Updated website ssl certificates'

	handleDomainNameUpdated = (data) ->
		$scope.domain.domainName = data

	getWebsiteSettings()

	domainNameUpdatedEvents = events('systemSettings', 'domain name updated', null).setCallback(handleDomainNameUpdated).subscribe()
	$scope.$on '$destroy', domainNameUpdatedEvents.unsubscribe

	$scope.submit = () ->
		submitDomainName() if $scope.domain.domainName? and $scope.domain.domainName isnt ''
		submitSslCertificate() if $scope.ssl.certificate? and $scope.ssl.certificate isnt '' and $scope.ssl.privateKey? and $scope.ssl.privateKey isnt ''
]
