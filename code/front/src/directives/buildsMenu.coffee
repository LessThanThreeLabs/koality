angular.module('koality.directive.buildsMenu', []).
	directive('buildsMenu', () ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<div class="buildsMenu unselectable" ng-transclude>
				<div class="buildsMenuBackgroundPanel"></div>
				<div class="buildsMenuFooter"></div>
			</div>'
	).
	directive('buildsMenuHeader', () ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<div class="buildsMenuHeader">
				<div class="buildsMenuHeaderContent" ng-transclude></div>
				<div class="buildsMenuHeaderBuffer"></div>
			</div>'
		link: (scope, element, attributes) ->
			element.addClass 'noPadding' if attributes.noPadding?
	).
	directive('buildsMenuOptions', () ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<div class="buildsMenuOptions">
				<div class="buildsMenuOptionsScrollWrapper onScrollToBottomDirectiveAnchor" ng-transclude></div>
			</div>'
	).
	directive('buildsMenuEmptyMessage', () ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<div class="buildsMenuEmptyMessage" ng-transclude></div>'
	).
	directive('buildsMenuRetrievingMore', () ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<div class="buildsMenuRetrievingMore" ng-transclude></div>'
	).
	directive('buildsMenuOption', () ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<div class="buildsMenuOption">
				<div class="buildsMenuOptionContents">
					<div class="buildsMenuOptionTextContainer" ng-transclude></div>
					<div class="buildsMenuOptionArrow"></div>
					<spinner class="buildsMenuOptionSpinner" running="spinning"></spinner>
				</div>
				<div class="buildsMenuOptionTooth"></div>
			</div>'
		link: (scope, element, attributes) ->
			attributes.$observe 'menuOptionSpinning', (spinning) ->
				# typeof spinning is 'string'...
				scope.spinning = spinning is 'true'

			scope.$watch 'spinning', () ->
				element.find('.buildsMenuOptionTextContainer').toggleClass 'spinnerTextPadding', scope.spinning
	)