'use strict'

angular.module('koality.filter', ['koality.service']).
	filter('emailToAlias', [() ->
		(input) ->
			return null if not input? or typeof input isnt 'string'

			if input.indexOf '@' isnt -1
				return input.substring 0, input.indexOf '@'
			else
				return input
	]).
	filter('ascii', [() ->
		(input) ->
			return null if not input? or typeof input isnt 'string'

			input = input.replace /\n/g, '<br>'
			input = input.replace /\t/g, '&nbsp;&nbsp;&nbsp;&nbsp;'
			input = input.replace /\040/g, '&nbsp;'
			return input
	]).
	filter('onlyFirstLine', [() ->
		(input) ->
			return null if not input? or typeof input isnt 'string'
			return input.split('\n')[0]
	]).
	filter('ansi', ['ansiParser', (ansiParser) ->
		(input) ->
			return null if not input? or typeof input isnt 'string'
			return ansiParsedLine = ansiParser.parse input
	]).
	filter('shaLink', [() ->
		(headSha, baseSha, repositoryInformation) ->
			return null if not headSha? or typeof headSha isnt 'string'
			return null if not repositoryInformation? or typeof repositoryInformation isnt 'object'

			generateLink = (path, text) ->
				url = 'http://' + domain + path
				return "<a href='#{url}' target='_blank'>#{text}</a>"

			getGitHubLink = () ->
				baseSha = null if baseSha is '0000000000000000000000000000000000000000'
				
				if typeof baseSha is 'string' and baseSha.length > 0 and baseSha isnt headSha
					return generateLink "/#{gitHubMatch[1]}/compare/#{baseSha}...#{headSha}", 'View Diff in GitHub'
				else
					return generateLink "/#{gitHubMatch[1]}/commit/#{headSha}", 'View Diff in GitHub'

			getBitBucketLink = () ->
				return generateLink "/#{bitBucketGitMatch[2]}/commits/#{headSha}", 'View Diff in BitBucket'

			forwardUrl = repositoryInformation.forwardUrl
			domain = forwardUrl.substring forwardUrl.indexOf('@') + 1, forwardUrl.indexOf(':')

			gitHubRegex = new RegExp "^git@#{domain}:(.+?)(\.git)?$"
			gitHubMatch = gitHubRegex.exec repositoryInformation.forwardUrl
			if (repositoryInformation.github? or domain is 'github.com') and gitHubMatch? and gitHubMatch[1]?
				return getGitHubLink()

			bitBucketRegex = new RegExp "^git@bitbucket.(org|com):(.+?)(\.git)?$"
			bitBucketGitMatch = bitBucketRegex.exec repositoryInformation.forwardUrl
			if bitBucketGitMatch? and bitBucketGitMatch[2]?
				return getBitBucketLink()

			return null
	])
