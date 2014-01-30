angular.module('koality.directive.dropdown', []).
	directive('dropdown', [() ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<div class="dropdownContainer unselectable" ng-transclude></div>'
		link: (scope, element, attributes) ->
			element.addClass 'dropdownLeft' if attributes.left?
			element.addClass 'dropdownRight' if attributes.right?
			element.addClass 'dropdownLight' if attributes.light?
	]).
	directive('dropdownOptions', ['$document', ($document) ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<div class="dropdownOptions" ng-click="hideDropdownOptions()" ng-transclude></div>'
		link: (scope, element, attributes) ->
			removeTemporaryHideWindowListener = () ->
				element.removeClass 'temporaryHide'
				$document.unbind 'mousemove', removeTemporaryHideWindowListener

			scope.hideDropdownOptions = () ->
				element.addClass 'temporaryHide'
				$document.bind 'mousemove', removeTemporaryHideWindowListener

	]).
	directive('dropdownOption', [() ->
		restrict: 'E'
		replace: true
		transclude: true
		template: '<a class="dropdownOption" ng-transclude></a>'
	])
