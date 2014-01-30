angular.module('koality.directive.panel', []).
	directive('panel', () ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<div class="panel unselectable" ng-transclude>
				<div class="panelFooter"></div>
			</div>'
		link: (scope, element, attributes) ->
			if element.find('.panelOptions').length > 0
				element.addClass 'optionsVisible'
	).
	directive('panelOptions', () ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<div class="panelOptions" ng-transclude></div>'
	).
	directive('panelOption', () ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<div class="panelOption" ng-transclude></div>'
		link: (scope, element, attributes) ->
			scope.$watch attributes.selected, (selected=false) ->
				element.toggleClass 'selected', selected
	).
	directive('panelHeader', () ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<div class="panelHeader" ng-transclude></div>'
		link: (scope, element, attributes) ->
			if attributes.panelHeaderPadding?
				element.addClass 'panelHeaderPadding'
	).
	directive('panelHeaderSubtext', () ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<div class="panelHeaderSubtext" ng-transclude></div>'
	).
	directive('panelBody', ['$timeout', ($timeout) ->
		restrict: 'E'
		replace: true
		transclude: true
		scope:
			openDrawer: '='
		template: '<div class="panelBody">
				<div class="panelBodyContents textSelectable" ng-transclude></div>
			</div>'
		link: (scope, element, attributes) ->
			hideOtherDrawersTimeoutPromise = null

			openDrawer = (drawerToOpen) ->
				panelDrawer = element.find(".panelDrawer[drawer-name='#{drawerToOpen}']")
				otherPanelDrawers = element.find(".panelDrawer:not([drawer-name='#{drawerToOpen}'])")
				panelBodyContents = element.find('.panelBodyContents')

				assert.ok panelDrawer.length <= 1
				return if panelBodyContents.length is 0

				$timeout.cancel hideOtherDrawersTimeoutPromise if hideOtherDrawersTimeoutPromise?

				hideOtherDrawers = () ->
					otherPanelDrawers.css 'z-index', 0
					otherPanelDrawers.css 'display', 'none'

				bringDrawerToTop = () ->
					panelDrawer.css 'z-index', 1
					panelDrawer.css 'display', 'block'

				if panelDrawer.length isnt 0
					hideOtherDrawers()
					bringDrawerToTop()
					panelBodyContents.css 'top', panelDrawer.outerHeight() + 'px'
				else
					panelBodyContents.css 'top', '0'
					hideOtherDrawersTimeoutPromise = $timeout (() -> hideOtherDrawers()), 1000

			moveDrawerOutOfNgTranslate = () ->
				element.append element.find('.panelDrawer')

			moveDrawerOutOfNgTranslate() if element.find('.panelDrawer').length > 0

			element.addClass 'noScroll' if attributes.noScroll?

			scope.$watch 'openDrawer', (drawerToOpen) ->
				openDrawer drawerToOpen
	]).
	directive('panelDrawer', () ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<div class="panelDrawer" ng-transclude></div>'
	)
