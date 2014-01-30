'use strict'

angular.module('koality', ['ngSanitize', 
		'koality.service', 'koality.service.socket', 'koality.service.repository', 'koality.service.changes',
		'koality.filter',
		'koality.directive', 'koality.directive.changesMenu', 'koality.directive.panel', 'koality.directive.dropdown', 'koality.d3.directive']).
	config(['$routeProvider', ($routeProvider) ->
		$routeProvider.
			when('/login',
				templateUrl: "/html/login/login#{fileSuffix}.html"
				controller: Login
				redirectTo: if window.accountInformation.id is '' then null else '/'
			).
			when('/account',
				templateUrl: "/html/account/account#{fileSuffix}.html"
				controller: Account
				reloadOnSearch: false
				redirectTo: if window.accountInformation.id is '' then '/login' else null
			).
			when('/create/account',
				templateUrl: "/html/createAccount/createAccount#{fileSuffix}.html"
				controller: CreateAccount
				reloadOnSearch: false
				redirectTo: if window.accountInformation.id is '' then null else '/'
			).
			when('/resetPassword',
				templateUrl: "/html/resetPassword/resetPassword#{fileSuffix}.html"
				controller: ResetPassword
				redirectTo: if window.accountInformation.id is '' then null else '/'
			).
			when('/repository/:repositoryId',
				templateUrl: "/html/repository/repository#{fileSuffix}.html"
				controller: Repository
				reloadOnSearch: false
				redirectTo: if window.accountInformation.id is '' then '/login' else null
			).
			when('/dashboard',
				templateUrl: "/html/dashboard/dashboard#{fileSuffix}.html"
				controller: Dashboard
				reloadOnSearch: false
				redirectTo: if window.accountInformation.id is '' then '/login' else null
			).
			when('/analytics',
				templateUrl: "/html/analytics/analytics#{fileSuffix}.html"
				controller: Analytics
				reloadOnSearch: false
				redirectTo: if window.accountInformation.id is '' then '/login' else null
			).
			when('/admin',
				templateUrl: "/html/admin/admin#{fileSuffix}.html"
				controller: Admin
				reloadOnSearch: false
				redirectTo: if window.accountInformation.isAdmin then null else '/'
			).
			otherwise(
				redirectTo: '/dashboard'
			)
	]).
	config(['$locationProvider', ($locationProvider) ->
		$locationProvider.html5Mode true
	]).
	run(() ->
		# initialization happens here
	)
