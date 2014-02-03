'use strict'

angular.module('koality', ['ngSanitize', 
		'koality.service', 'koality.service.socket', 'koality.service.repository', 'koality.service.changes',
		'koality.filter',
		'koality.directive', 'koality.directive.changesMenu', 'koality.directive.panel', 'koality.directive.dropdown', 'koality.d3.directive']).
	config(['$routeProvider', ($routeProvider) ->
		$routeProvider.
			when('/login',
				templateUrl: "/html/login/login.html"
				controller: Login
				redirectTo: if window.userId? then '/' else null
			).
			when('/account',
				templateUrl: "/html/account/account.html"
				controller: Account
				reloadOnSearch: false
				redirectTo: if window.userId? then null else '/login'
			).
			when('/create/account',
				templateUrl: "/html/createAccount/createAccount.html"
				controller: CreateAccount
				reloadOnSearch: false
				redirectTo: if window.userId? then '/' else null
			).
			when('/resetPassword',
				templateUrl: "/html/resetPassword/resetPassword.html"
				controller: ResetPassword
				redirectTo: if window.userId? then '/' else null
			).
			when('/dashboard',
				templateUrl: "/html/dashboard/dashboard.html"
				controller: Dashboard
				reloadOnSearch: false
				redirectTo: if window.userId? then null else '/login'
			).
			when('/repository/:repositoryId',
				templateUrl: "/html/repository/repository.html"
				controller: Repository
				reloadOnSearch: false
				redirectTo: if window.userId? then null else '/login'
			).
			when('/analytics',
				templateUrl: "/html/analytics/analytics.html"
				controller: Analytics
				reloadOnSearch: false
				redirectTo: if window.userId? then null else '/login'
			).
			when('/admin',
				templateUrl: "/html/admin/admin.html"
				controller: Admin
				reloadOnSearch: false
				redirectTo: if window.isAdmin then null else '/'
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
