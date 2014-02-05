'use strict'

angular.module('koality.directive', []).
	directive('focused', () ->
		return (scope, element, attributes) ->
			element.focus()
	).
	directive('inputFocusOnClick', () ->
		return (scope, element, attributes) ->
			element.bind 'click', (event) ->
				element.find('input').focus()
	).
	directive('highlightOnClick', () ->
		return (scope, element, attributes) ->
			highlightText = () ->
				element.select()
				element.unbind 'click', highlightText if attributes.firstOnly?

			element.bind 'click', highlightText
	).
	directive('autoFillableField', () ->
		# This is a hack until the following bug is fixed:
		# https://github.com/angular/angular.js/issues/1460
		return (scope, element, attributes) ->
			element.on 'change.autofill DOMAttrModified.autofill keydown.autofill propertychange.autofill', () ->
				element.trigger 'input' if element.val() isnt ''
	).
	directive('hideOnExternalClick', ['$document', '$timeout', ($document, $timeout) ->
		restrict: 'A'
		link: (scope, element, attributes) ->
			addListeners = () ->
				element.bind 'click', handleInternalClick
				$timeout (() -> $document.bind 'click', handleExternalClick)

			removeListeners = () ->
				element.unbind 'click', handleInternalClick
				$document.unbind 'click', handleExternalClick

			handleInternalClick = (event) ->
				event.stopPropagation()

			handleExternalClick = () ->
				scope.$apply attributes.hideFunction

			scope.$watch attributes.hideOnExternalClick, () ->
				if scope.$eval attributes.hideOnExternalClick then addListeners()
				else removeListeners()
	]).
	directive('licenseKey', () ->
		restrict: 'A'
		require: 'ngModel'
		link: (scope, element, attributes, control) ->
			control.$parsers.unshift (viewValue) ->
				return viewValue if not viewValue?

				cleanedValue = viewValue.replace /[^a-zA-Z0-9]/g, ''
				control.$setValidity 'licenseKey', cleanedValue.length is 16

				return if cleanedValue.length is 16 then cleanedValue else undefined
	).
	directive('fadingContent', ['$window', '$timeout', ($window, $timeout) ->
		restrict: 'A'
		replace: true
		transclude: true
		template: '<div class="fadingContentContainer">
				<div class="fadingContentTopBuffer"></div>
				<div class="fadingContentScrollWrapper onScrollToTopDirectiveAnchor onScrollToBottomDirectiveAnchor autoScrollToBottomDirectiveAnchor">
					<div class="fadingContent" ng-transclude></div>
				</div>
				<div class="fadingContentBottomBuffer"></div>
			</div>'
		link: (scope, element, attributes) ->
			fadingContentElement = element.find('.fadingContent')
			
			setFadingContentWidth = () ->
				fadingContentElement.width element.width() if element.width() > 100

			setFadingContentWidth()
			$($window).resize () -> setFadingContentWidth()
	]).
	directive('autoScrollToBottom', ['$timeout', 'integerConverter', ($timeout, integerConverter) ->
		restrict: 'A'
		link: (scope, element, attributes) ->
			if element.find('.autoScrollToBottomDirectiveAnchor').length > 0
				element = element.find('.autoScrollToBottomDirectiveAnchor')[0]

			scrollToBottom = () ->
				$timeout (() -> element.scrollTop = element.scrollHeight)

			scrollBottomBuffer = integerConverter.toInteger(attributes.autoScrollToBottomBuffer) ? 20

			scope.$watch attributes.autoScrollToBottom, (newValue, oldValue) ->
				isFirstRender = () ->
					return not oldValue? or oldValue.length is 0 or Object.keys(oldValue).length is 0

				if isFirstRender()
					scrollToBottom() if attributes.startAtBottom?
				else
					isScrolledToBottomIsh = element.scrollTop + element.offsetHeight + scrollBottomBuffer >= element.scrollHeight
					scrollToBottom() if isScrolledToBottomIsh
	]).
	directive('onScrollToTop', [() ->
		restrict: 'A'
		link: (scope, element, attributes) ->
			if element.find('.onScrollToTopDirectiveAnchor').length > 0
				element = element.find '.onScrollToTopDirectiveAnchor'

			addScrollListener = () ->
				element.bind 'scroll', (event) ->
					scope.$apply attributes.onScrollToTop if element[0].scrollTop is 0

			addScrollListener()
	]).
	directive('onScrollToBottom', [() ->
		restrict: 'A'
		link: (scope, element, attributes) ->
			if element.find('.onScrollToBottomDirectiveAnchor').length > 0
				element = element.find '.onScrollToBottomDirectiveAnchor'

			addScrollListener = () ->
				element.bind 'scroll', (event) ->
					scrolledToBottom = element[0].scrollTop + element[0].offsetHeight >= element[0].scrollHeight
					scope.$apply attributes.onScrollToBottom if scrolledToBottom

			addScrollListener()
	]).
	directive('busyButton', ['$timeout', ($timeout) ->
		restrict: 'E'
		replace: true
		transclude: true
		scope:
			show: '=show'
			busy: '=busy'
			disabled: '=disabled'
			clickHandler: '&click'
		template: '<div class="busyButton" ng-show="(show != null && show) || (show == null)">
				<button ng-show="!busy" ng-mouseup="click($event)" ng-click="click($event)" ng-disabled="disabled" ng-transclude></button>
				<spinner white running="busy" class="busyButtonSpinner"></spinner>
			</div>'
		link: (scope, element, attributes) ->
			element.css
				position: 'relative'
				'z-index': 1

			button = element.find 'button'
			button.addClass 'centered' if attributes.centered?
			button.addClass 'fullWidth' if attributes.fullWidth?
			button.addClass 'red' if attributes.red?

			spinner = element.find '.busyButtonSpinner'
			spinner.css
				position: 'absolute'
				top: 0
				left: 0
				right: 0
				bottom: 0
				'z-index': -1

			element.width Math.max button.outerWidth() + 1, 25 if not attributes.centered?
			element.height Math.max button.outerHeight(), 25

			# This allows us to have more responsive buttons that respond to
			# mouseup (because buttons are buggy in osx...) while still supporting
			# form submission logic (clicking the enter key while focus is in a form)
			ignoreClickEvents = false
			scope.click = (event) ->
				return if ignoreClickEvents
				ignoreClickEvents = true

				scope.clickHandler()

				$timeout (() -> ignoreClickEvents = false), 100
	]).
	directive('spinner', () ->
		restrict: 'E'
		replace: true
		scope: running: '=running'
		template: '<div class="spinnerContainer"></div>'
		link: (scope, element, attributes) ->
			spinner = null

			spinnerOptions =
				lines: 7
				length: 3
				width: 3
				radius: 5
				corners: 1
				rotate: 14
				color: '#FFFFFF'
				speed: 1.2
				trail: 30
				shadow: false
				hwaccel: true
				className: 'spinner'
				zIndex: 2e9
				top: 'auto'
				left: 'auto'

			scope.$watch 'running', () ->
				if scope.running then startSpinner() else stopSpinner()

			startSpinner = () ->
				spinner = new Spinner(spinnerOptions).spin(element[0])

			stopSpinner = () ->
				spinner.stop() if spinner?
	).
	directive('notification', ['$timeout', ($timeout) ->
		restrict: 'E'
		replace: true
		transclude: true
		scope:
			type: '@'
			durationInSeconds: '@'
		template: '<div class="notification" ng-class="{green: type == \'success\', orange: type == \'warning\', red: type == \'error\'}">
					<div class="notificationContent growingCentered">
						<span class="notificationMessage" ng-transclude></span>
						<span class="notificationClose" ng-click="hide()">X</span>
					</div>
				</div>'
		link: (scope, element, attributes) ->
			getZIndex = (type, transient) ->
				minorIndex = 0
				switch type
					when 'success' then minorIndex = 100
					when 'warning' then minorIndex = 200
					when 'error' then minorIndex = 300
					else throw 'unexpected notification type ' + type
				return if transient then minorIndex + 1000 else minorIndex

			scope.hide = () ->
				element.addClass 'hiding'
				$timeout (() -> element.remove()), 5000

			element.css 'z-index', getZIndex attributes.type, attributes.durationInSeconds > 0
			$timeout scope.hide, attributes.durationInSeconds * 1000 if attributes.durationInSeconds > 0
	]).
	directive('consoleText', ['$timeout', 'ansiParser', ($timeout, ansiParser) ->
		restrict: 'E'
		replace: true
		scope: 
			oldLines: '=oldLines'
			newLines: '=newLines'
			removeLines: '&removeLines'
		template: '<div class="consoleText"></div>'
		link: (scope, element, attributes) ->
			oldLineNumberBounds = null
			scrollableElement = element.closest('.onScrollToTopDirectiveAnchor')[0]

			getLineNumberBounds = (lines) ->
				return null if Object.keys(lines).length is 0

				minLineNumber = Number.MAX_VALUE
				maxLineNumber = Number.MIN_VALUE

				for lineNumber in Object.keys lines
					minLineNumber = Math.min minLineNumber, lineNumber
					maxLineNumber = Math.max maxLineNumber, lineNumber

				return {min: minLineNumber, max: maxLineNumber}

			clearLines = () ->
				element.empty()
				oldLineNumberBounds = null

			generateLineHtml = (lineNumber, lineText) ->
				ansiParsedLine = ansiParser.parse (lineText ? '')
				return "<div class='line' number=#{lineNumber}><span class='number'>#{lineNumber}&nbsp;&nbsp;</span><span class='text'>#{ansiParsedLine}</span></div>"

			renderInitialLines = (lines) ->
				lineNumberBounds = getLineNumberBounds lines
				return if not lineNumberBounds?

				htmlToAppend = []
				for index in [lineNumberBounds.min..lineNumberBounds.max]
					htmlToAppend.push generateLineHtml index, lines[index]?.text

				element.append htmlToAppend.join ''

				oldLineNumberBounds = lineNumberBounds

			updateLines = (newLines, oldLines) ->
				newLineNumberBounds = getLineNumberBounds newLines
				return if not newLineNumberBounds?

				keepScrollPosition = () ->
					oldScrollHeight = scrollableElement.scrollHeight

					$timeout () -> 
						newScrollHeight = scrollableElement.scrollHeight
						scrollableElement.scrollTop = newScrollHeight - oldScrollHeight

				addLinesThatAreBeforeExistingLines = () ->
					return if newLineNumberBounds.min >= oldLineNumberBounds.min

					htmlToPrepend = []
					for index in [newLineNumberBounds.min...oldLineNumberBounds.min]
						htmlToPrepend.push generateLineHtml index, newLines[index]?.text

					element.prepend htmlToPrepend.join ''

				updateLinesThatAlreadyExist = () ->
					for lineNumber, line of newLines
						continue if oldLines[lineNumber]?.hash is newLines[lineNumber]?.hash

						ansiParsedLine = ansiParser.parse (line.text ? '')
						element.find(".line[number='#{lineNumber}']").find('.text').html ansiParsedLine

				addLinesThatAreAfterExistingLines = () ->
					return if newLineNumberBounds.max <= oldLineNumberBounds.max

					htmlToAppend = []
					for index in [(oldLineNumberBounds.max + 1)..newLineNumberBounds.max]
						htmlToAppend.push generateLineHtml index, newLines[index]?.text
					
					element.append htmlToAppend.join ''

				if newLineNumberBounds.min < oldLineNumberBounds.min and newLineNumberBounds.max <= oldLineNumberBounds.max
					keepScrollPosition()

				addLinesThatAreBeforeExistingLines()
				updateLinesThatAlreadyExist()
				addLinesThatAreAfterExistingLines()

				oldLineNumberBounds =
					min: Math.min oldLineNumberBounds.min, newLineNumberBounds.min
					max: Math.max oldLineNumberBounds.max, newLineNumberBounds.max

			removeLines = () =>
				numLinesToMaintain = 2000

				removeLinesFromView = (numLinesToRemove) =>
					element.children().slice(0, numLinesToRemove).remove()

				isScrolledToBottomIsh = scrollableElement.scrollTop + scrollableElement.offsetHeight + 100 >= scrollableElement.scrollHeight
				if isScrolledToBottomIsh and (oldLineNumberBounds.max - oldLineNumberBounds.min + 1 > numLinesToMaintain)
					numLinesToRemove = oldLineNumberBounds.max - oldLineNumberBounds.min - numLinesToMaintain + 1
					scope.removeLines 
						startIndex: oldLineNumberBounds.min
						numLines: numLinesToRemove
					removeLinesFromView numLinesToRemove
					oldLineNumberBounds.min = oldLineNumberBounds.min + numLinesToRemove

			handleLinesUpdate = () ->
				if (not scope.oldLines? or Object.keys(scope.oldLines).length is 0) and 
					(not scope.newLines? or Object.keys(scope.newLines).length is 0)
						clearLines()
				else if not scope.oldLines? or Object.keys(scope.oldLines).length is 0
					renderInitialLines scope.newLines
				else
					removeLines()
					updateLines scope.newLines, scope.oldLines

			scope.$watch 'newLines', handleLinesUpdate
	])
