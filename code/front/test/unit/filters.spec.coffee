'use strict'

describe 'Koality filters', () ->

	describe 'fileSuffix filter', () ->
		fileSuffix = null
		suffixString = '_qa8aset32'

		beforeEach module 'koality.filter', ($provide) ->
			mockedInitialState = fileSuffix: suffixString
			$provide.value 'initialState', mockedInitialState
			return

		beforeEach () ->
			inject (fileSuffixFilter) ->
				fileSuffix = fileSuffixFilter

		it 'should correctly add file suffix for valid file urls', () ->
			expect(fileSuffix 'hello.png').toBe "hello#{suffixString}.png"
			expect(fileSuffix 'hello/there.jpg').toBe "hello/there#{suffixString}.jpg"
			expect(fileSuffix '/hello/there/sir.gif').toBe "/hello/there/sir#{suffixString}.gif"

		it 'should fail to add the correct file suffix for invalid file urls', () ->
			expect(fileSuffix 'hellopng').toBe 'hellopng'
			expect(fileSuffix 'hello/therejpg').toBe 'hello/therejpg'
			expect(fileSuffix '/hello/there/sirgif').toBe '/hello/there/sirgif'

	describe 'ascii filter', () ->
		beforeEach module 'koality.filter'

		it 'should handle invalid values', () ->
			inject (asciiFilter) ->
				expect(asciiFilter null).toBe null
				expect(asciiFilter 15).toBe null
				expect(asciiFilter 15.1).toBe null
				expect(asciiFilter {}).toBe null

		it 'should change spaces to &nbsp;', () ->
			inject (asciiFilter) ->
				expect(asciiFilter '').toBe ''
				expect(asciiFilter 'hello').toBe 'hello'
				expect(asciiFilter 'hello there').toBe 'hello&nbsp;there'
				expect(asciiFilter ' hello').toBe '&nbsp;hello'
				expect(asciiFilter 'hello ').toBe 'hello&nbsp;'

		it 'should change tabs to &nbsp;', () ->
			inject (asciiFilter) ->
				expect(asciiFilter '').toBe ''
				expect(asciiFilter 'hello').toBe 'hello'
				expect(asciiFilter 'hello\tthere').toBe 'hello&nbsp;&nbsp;&nbsp;&nbsp;there'
				expect(asciiFilter '\thello').toBe '&nbsp;&nbsp;&nbsp;&nbsp;hello'
				expect(asciiFilter 'hello\t').toBe 'hello&nbsp;&nbsp;&nbsp;&nbsp;'

		it 'should add <br>\'s correctly', () ->
			inject (asciiFilter) ->
				expect(asciiFilter '\n').toBe '<br>'
				expect(asciiFilter 'hello\n').toBe 'hello<br>'
				expect(asciiFilter '\nhello').toBe '<br>hello'
				expect(asciiFilter 'hello\nthere').toBe 'hello<br>there'
				expect(asciiFilter '\nhello\nthere').toBe '<br>hello<br>there'
				expect(asciiFilter '\nhello\nthere\n').toBe '<br>hello<br>there<br>'
