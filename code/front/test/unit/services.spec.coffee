'use strict'

describe 'Koality services', () ->

	describe 'initial state', () ->

		it 'should have non-null values', () ->
			fileSuffix = '_d487ab5e'; csrfToken = '4ed9a4a31had'; userId = 17
			email ='email@address.com'; firstName = 'First'; lastName = 'Last'

			module 'koality.service', ($provide) ->
				mockedWindow = 
					fileSuffix: fileSuffix
					csrfToken: csrfToken
					accountInformation:
						id: userId
						email: email
						firstName: firstName
						lastName: lastName
				$provide.value '$window', mockedWindow
				return

			inject (initialState) ->
				expect(initialState.fileSuffix).toBe fileSuffix
				expect(initialState.csrfToken).toBe csrfToken
				expect(initialState.user.id).toBe userId
				expect(initialState.user.email).toBe email
				expect(initialState.user.firstName).toBe firstName
				expect(initialState.user.lastName).toBe lastName

		it 'should have null values', () ->
			module 'koality.service', ($provide) ->
				mockedWindow = 
					fileSuffix: ''
					csrfToken: ''
					accountInformation:
						id: ''
						email: ''
						firstName: ''
						lastName: ''
				$provide.value '$window', mockedWindow
				return

			inject (initialState) ->
				expect(initialState.fileSuffix).toBeNull()
				expect(initialState.csrfToken).toBeNull()
				expect(initialState.user.id).toBeNull()
				expect(initialState.user.email).toBeNull()
				expect(initialState.user.firstName).toBeNull()
				expect(initialState.user.lastName).toBeNull()

		it 'should not allow edits', () ->
			module 'koality.service'

			inject (initialState) ->
				addValueToInitialState = () ->
					initialState.blah = 'blah'

				removeValueFromInitialState = () ->
					delete initialState.user

				expect(addValueToInitialState).toThrow()
				expect(removeValueFromInitialState).toThrow()

	describe 'file suffix adder', () ->
		fileSuffix = '_hc8oeb1f'
		
		beforeEach () ->
			module 'koality.service', ($provide) ->
				mockedInitialState = fileSuffix: fileSuffix
				$provide.value 'initialState', mockedInitialState
				return
			
		it 'should properly add the correct file suffix for valid file urls', () ->
			inject (fileSuffixAdder) ->
				expect(fileSuffixAdder.addFileSuffix('a.gif')).toBe "a#{fileSuffix}.gif"
				expect(fileSuffixAdder.addFileSuffix('/img/a.png')).toBe "/img/a#{fileSuffix}.png"
				expect(fileSuffixAdder.addFileSuffix('/short.html')).toBe "/short#{fileSuffix}.html"
				expect(fileSuffixAdder.addFileSuffix('/img/longerName.fakeExtension')).toBe "/img/longerName#{fileSuffix}.fakeExtension"

		it 'should fail to add the correct file suffix for invalid file urls', () ->
			inject (fileSuffixAdder) ->
				expect(fileSuffixAdder.addFileSuffix('agif')).toBe 'agif'
				expect(fileSuffixAdder.addFileSuffix('/img/apng')).toBe '/img/apng'
				expect(fileSuffixAdder.addFileSuffix('/shorthtml')).toBe '/shorthtml'
				expect(fileSuffixAdder.addFileSuffix('/img/longerNamefakeExtension')).toBe '/img/longerNamefakeExtension'

	describe 'integer converter', () ->
		beforeEach module 'koality.service'

		it 'should return integeres given integers', () ->
			inject (integerConverter) ->
				expect(integerConverter.toInteger(5)).toBe 5
				expect(integerConverter.toInteger(-1)).toBe -1
				expect(integerConverter.toInteger(9001)).toBe 9001

		it 'should return null given floats', () ->
			inject (integerConverter) ->
				expect(integerConverter.toInteger(5.1)).toBeNull()
				expect(integerConverter.toInteger(-1.7)).toBeNull()
				expect(integerConverter.toInteger(9001.1238907)).toBeNull()

		it 'should return numbers given valid strings', () ->
			inject (integerConverter) ->
				expect(typeof integerConverter.toInteger('5')).toBe 'number'
			
		it 'should properly parse valid integers', () ->
			inject (integerConverter) ->
				expect(integerConverter.toInteger '5').toBe 5
				expect(integerConverter.toInteger '-1').toBe -1
				expect(integerConverter.toInteger '123456789').toBe 123456789

		it 'should return null for invalid integers', () ->
			inject (integerConverter) ->
				expect(integerConverter.toInteger '').toBeNull()
				expect(integerConverter.toInteger null).toBeNull()
				expect(integerConverter.toInteger undefined).toBeNull()
				expect(integerConverter.toInteger '1.3').toBeNull()
				expect(integerConverter.toInteger 'five').toBeNull()

	describe 'ansiParser service', () ->
		beforeEach module 'koality.service'

		it 'should wrap plaintext in default styling', () -> 
			inject (ansiParser) ->
				expect(ansiParser.parse 'thisIsATest').toBe '<span class="ansi"><span class="foregroundDefault backgroundDefault">thisIsATest</span></span>'

		it 'should escape spaces', () ->
			inject (ansiParser) ->
				expect(ansiParser.parse 'this is  A   test').toBe '<span class="ansi"><span class="foregroundDefault backgroundDefault">this&nbsp;is&nbsp;&nbsp;A&nbsp;&nbsp;&nbsp;test</span></span>'

		it 'should handle changing columns', () ->
			inject (ansiParser) ->
				expect(ansiParser.parse 'one\x1b[5Gtwo').toBe '<span class="ansi"><span class="foregroundDefault backgroundDefault">one&nbsp;&nbsp;two</span></span>'
				expect(ansiParser.parse 'one\x1b[0Gtwo').toBe '<span class="ansi"><span class="foregroundDefault backgroundDefault">two</span></span>'

		it 'should handle colors, bold, and clearing', () ->
			inject (ansiParser) ->
				expect(ansiParser.parse '\x1b[33;45;1mbright yellow on magenta, \x1b[0mdefault').toBe '<span class="ansi"><span class="foregroundYellow backgroundMagenta bright">bright&nbsp;yellow&nbsp;on&nbsp;magenta,&nbsp;</span>' +
					'<span class="foregroundDefault backgroundDefault">default</span></span>'

		it 'should persist colors and bold through other styles', () ->
			inject (ansiParser) ->
				expect(ansiParser.parse 'default, \x1b[31;1mbright red, \x1b[34;46mbright blue on cyan, \x1b[22mblue on cyan')
					.toBe '<span class="ansi"><span class="foregroundDefault backgroundDefault">default,&nbsp;</span>' +
						'<span class="foregroundRed backgroundDefault bright">bright&nbsp;red,&nbsp;</span>' +
						'<span class="foregroundBlue backgroundCyan bright">bright&nbsp;blue&nbsp;on&nbsp;cyan,&nbsp;</span>' +
						'<span class="foregroundBlue backgroundCyan">blue&nbsp;on&nbsp;cyan</span></span>'

		it 'should overwrite old characters with new styles', () ->
			inject (ansiParser) ->
				expect(ansiParser.parse 'plain\x1b[0G\x1b[32mgreen').toBe '<span class="ansi"><span class="foregroundGreen backgroundDefault">green</span></span>'

		it 'should handle carriage returns', () ->
			inject (ansiParser) ->
				expect(ansiParser.parse '123456789\r9876').toBe '<span class="ansi"><span class="foregroundDefault backgroundDefault">987656789</span></span>'

		it 'should escape dangerous characters', () ->
			inject (ansiParser) ->
				expect(ansiParser.parse '&<>"\'/').toBe '<span class="ansi"><span class="foregroundDefault backgroundDefault">&amp;&lt;&gt;&quot;&#39;&#x2F;</span></span>'

	describe 'rpc', () ->
		mockedSocket = null

		beforeEach () ->
			jasmine.Clock.useMock()

			mockedSocket =
				makeRequest: jasmine.createSpy('makeRequest').andCallFake (resource, requestType, methodName, data, callback) ->
					setTimeout (() -> if callback? then callback null, 'ok'), 100

			module 'koality.service', ($provide) ->
				$provide.value 'socket', mockedSocket
				return
			
		it 'should properly call socket when making rpc requests', () ->
			inject (rpc) ->
				rpc 'users', 'update', 'login', id: 9001
				expect(mockedSocket.makeRequest).toHaveBeenCalled()
				expect(mockedSocket.makeRequest.callCount).toBe 1

				rpc 'users', 'update', 'logout', id: 9001
				expect(mockedSocket.makeRequest).toHaveBeenCalled()
				expect(mockedSocket.makeRequest.callCount).toBe 2

		it 'should have callback called after some delay', () ->
			inject (rpc) ->
				fakeCallback = jasmine.createSpy 'fakeCallback'

				rpc 'users', 'update', 'login', id: 9001, fakeCallback
				expect(mockedSocket.makeRequest).toHaveBeenCalled()
				expect(mockedSocket.makeRequest.callCount).toBe 1
				expect(fakeCallback).not.toHaveBeenCalled()
				jasmine.Clock.tick 101
				expect(fakeCallback).toHaveBeenCalled()

	describe 'events', () ->
		it 'should request to listen on eventName given from the socket', () ->
			eventToFireOn = 'aeoutnhuaensuha'
			mockedSocket =
				makeRequest: jasmine.createSpy('makeRequest').andCallFake (resource, requestType, methodName, data, callback) ->
					callback null, eventToFireOn
				respondTo: jasmine.createSpy('respondTo')

			module 'koality.service', ($provide) ->
				$provide.value 'socket', mockedSocket
				return

			inject (events) ->
				fakeCallback = () ->
				events('users', 'basic information', 9001).setCallback(fakeCallback).subscribe()

				expect(mockedSocket.respondTo.mostRecentCall.args[0]).toBe eventToFireOn
				expect(mockedSocket.makeRequest.callCount).toBe 1
				expect(mockedSocket.respondTo.callCount).toBe 1

		it 'should handle subscribe and unsubscribe', () ->
			jasmine.Clock.useMock()

			interval = null
			eventToFireOn = 'aeoutnhuaensuha'
			mockedSocket =
				makeRequest: jasmine.createSpy('makeRequest').andCallFake (resource, requestType, methodName, data, callback) ->
					if requestType is 'subscribe' then callback null, eventToFireOn
					if requestType is 'unsubscribe' then clearInterval interval
				respondTo: jasmine.createSpy('respondTo').andCallFake (eventName, callback) ->
					assert.ok eventName is eventToFireOn
					interval = setInterval (() -> callback 'hello'), 100

			module 'koality.service', ($provide) ->
				$provide.value 'socket', mockedSocket
				return

			inject (events) ->
				fakeCallback = jasmine.createSpy 'fakeCallback'
				fakeEvents = events('users', 'basic information', 9001).setCallback(fakeCallback).subscribe()

				expect(fakeCallback.callCount).toBe 0
				jasmine.Clock.tick 101
				expect(fakeCallback.callCount).toBe 1
				jasmine.Clock.tick 100
				expect(fakeCallback.callCount).toBe 2

				fakeEvents.unsubscribe()
				jasmine.Clock.tick 500
				expect(fakeCallback.callCount).toBe 2	

	describe 'changes rpc', () ->
		beforeEach () ->
			numChanges = 107
			mockedRpc = (resource, requestType, methodName, data, callback) ->
					endIndex = Math.min numChanges, data.startIndex + 100
					callback null, (num for num in [data.startIndex...endIndex])

			module 'koality.service', ($provide) ->
				$provide.value 'rpc', mockedRpc
				return
			
		it 'should receive changes', () ->
			inject (changesRpc) ->
				fakeCallback = jasmine.createSpy 'fakeCallback'

				changesRpc.queueRequest 17, 'all', null, 0, fakeCallback
				expect(fakeCallback.callCount).toBe 1
				expect(fakeCallback.mostRecentCall.args[1].length).toBe 100


		it 'should stop receiving changes if no more to receive', () ->
			inject (changesRpc) ->
				fakeCallback = jasmine.createSpy 'fakeCallback'

				changesRpc.queueRequest 17, 'all', null, 0, fakeCallback
				expect(fakeCallback.callCount).toBe 1
				expect(fakeCallback.mostRecentCall.args[1].length).toBe 100

				changesRpc.queueRequest 17, 'all', null, 100, fakeCallback
				expect(fakeCallback.callCount).toBe 2
				expect(fakeCallback.mostRecentCall.args[1].length).toBe 7

				changesRpc.queueRequest 17, 'all', null, 107, fakeCallback
				expect(fakeCallback.callCount).toBe 2

				changesRpc.queueRequest 17, 'all', null, 0, fakeCallback
				expect(fakeCallback.callCount).toBe 3
				expect(fakeCallback.mostRecentCall.args[1].length).toBe 100
