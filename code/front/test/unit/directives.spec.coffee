'use strict'

describe 'Koality directives', () ->

	describe 'dropdown directive', () ->
		element = null
		scope = null

		beforeEach module 'koality.directive'

		beforeEach () ->
			inject ($rootScope, $compile) ->
				scope = $rootScope.$new()
				element = angular.element '<dropdown alignment="left" dropdown-options="dropdownOptions" dropdown-show="dropdownShow" dropdown-option-click="handleClick(dropdownOption)" />'
				$compile(element)(scope)
				scope.$digest()

		it 'should render the correct number of options', () ->
			createOption = (optionNum) ->
				title: optionNum
				name: optionNum

			numOptions = 5
			scope.$apply () -> scope.dropdownOptions = (createOption optionNum for optionNum in [0...numOptions])
			options = element.find('.prettyDropdownOption')
			expect(options.length).toBe numOptions + 1  # to account for "-- empty --" option

			numOptions = 7
			scope.$apply () -> scope.dropdownOptions = (createOption optionNum for optionNum in [0...numOptions])
			options = element.find('.prettyDropdownOption')
			expect(options.length).toBe numOptions + 1  # to account for "-- empty --" option

		it 'should handle option click properly', () ->
			scope.$apply () -> scope.dropdownOptions = [{title: 'First', name: 'first'}, {title: 'Second', name: 'second'}]
			
			scope.handleClick = (dropdownOptionName) ->
			spyOn(scope, 'handleClick')

			options = element.find '.prettyDropdownOption'

			options.eq(0).click()
			expect(scope.handleClick).toHaveBeenCalledWith 'first'
			expect(scope.handleClick.calls.length).toBe 1

			options.eq(1).click()
			expect(scope.handleClick).toHaveBeenCalledWith 'second'
			expect(scope.handleClick.calls.length).toBe 2

			options.eq(1).click()
			expect(scope.handleClick.calls.length).toBe 3

	describe 'modal directive', () ->
		element = null
		scope = null

		beforeEach module 'koality.directive'

		it 'should become invisible when background is clicked', () ->
			inject ($rootScope, $compile) ->
				scope = $rootScope.$new()
				scope.modalVisible = true
				element = angular.element '<modal modal-visible="modalVisible">blah</modal>'
				$compile(element)(scope)
				scope.$digest()

			expect(scope.modalVisible).toBe true
			element.find('.prettyModalBackdrop').click()
			expect(scope.modalVisible).toBe false

	describe 'tooltip directive', () ->
		element = null
		scope = null

		beforeEach module 'koality.directive'

		it 'should render the correct tooltip text when valid text is provided', () ->
			tooltipText = 'blah'

			inject ($rootScope, $compile) ->
				scope = $rootScope.$new()
				element = angular.element "<span tooltip='#{tooltipText}'>hello</span>"
				$compile(element)(scope)
				scope.$digest()

			expect(element.find('.prettyTooltip').html()).toBe tooltipText

		it 'should render no tooltip text when no valid text is provided', () ->
			inject ($rootScope, $compile) ->
				scope = $rootScope.$new()
				element = angular.element "<span tooltip>hello</span>"
				$compile(element)(scope)
				scope.$digest()

			expect(element.find('.prettyTooltip').html()).toBe ''
			