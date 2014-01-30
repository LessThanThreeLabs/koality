'use strict'

describe 'Koality redirection', () ->

	it 'should redirect to login if not logged in', () ->
		browser().navigateTo '/'
		# dump browser().location().url()
		expect(browser().location().url()).toBe '/events'
