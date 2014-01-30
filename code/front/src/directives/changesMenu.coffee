angular.module('koality.directive.changesMenu', []).
	directive('changesMenu', () ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<div class="changesMenu unselectable" ng-transclude>
				<div class="changesMenuBackgroundPanel"></div>
				<div class="changesMenuFooter"></div>
			</div>'
	).
	directive('changesMenuHeader', () ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<div class="changesMenuHeader">
				<div class="changesMenuHeaderContent" ng-transclude></div>
				<div class="changesMenuHeaderBuffer"></div>
			</div>'
		link: (scope, element, attributes) ->
			element.addClass 'noPadding' if attributes.noPadding?
	).
	directive('changesMenuOptions', () ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<div class="changesMenuOptions">
				<div class="changesMenuOptionsScrollWrapper onScrollToBottomDirectiveAnchor" ng-transclude></div>
			</div>'
	).
	directive('changesMenuEmptyMessage', () ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<div class="changesMenuEmptyMessage" ng-transclude></div>'
	).
	directive('changesMenuRetrievingMore', () ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<div class="changesMenuRetrievingMore" ng-transclude></div>'
	).
	directive('changesMenuOption', () ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<div class="changesMenuOption">
				<div class="changesMenuOptionContents">
					<div class="changesMenuOptionTextContainer" ng-transclude></div>
					<div class="changesMenuOptionArrow"></div>
					<spinner class="changesMenuOptionSpinner" running="spinning"></spinner>
				</div>
				<div class="changesMenuOptionTooth"></div>
			</div>'
		link: (scope, element, attributes) ->
			attributes.$observe 'menuOptionSpinning', (spinning) ->
				# typeof spinning is 'string'...
				scope.spinning = spinning is 'true'

			scope.$watch 'spinning', () ->
				element.find('.changesMenuOptionTextContainer').toggleClass 'spinnerTextPadding', scope.spinning
	)