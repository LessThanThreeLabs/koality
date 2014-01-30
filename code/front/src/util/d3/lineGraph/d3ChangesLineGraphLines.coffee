window.D3ChangesLineGraphLines = {}

window.D3ChangesLineGraphLines.create = (element, svg) ->
	d3ChangesLineGraphLines = new @clazz element, svg
	d3ChangesLineGraphLines.initialize()
	return d3ChangesLineGraphLines

window.D3ChangesLineGraphLines.clazz = class D3ChangesLineGraphLines
	PADDING: {top: 10, left: 35, right: 10, bottom: 25}
	AXIS_BUFFER: 10

	constructor: (@element, @svg) ->
		assert.ok @element? and typeof @element is 'object'
		assert.ok @svg? and typeof @svg is 'object'

	initialize: () =>
		@rootPanel = @svg.append 'g'

		@allLine = @svg.append('path').attr 'class', 'allLine'
		@passedLine = @svg.append('path').attr 'class', 'passedLine'
		@failedLine = @svg.append('path').attr 'class', 'failedLine'

		@allLineDots = @svg.append('g').attr 'class', 'allLineDots'
		@passedLineDots = @svg.append('g').attr 'class', 'passedLineDots'
		@failedLineDots = @svg.append('g').attr 'class', 'failedLineDots'

		@allTooltip = d3.select(@element[0]).append('xhtml:div')
			.html('<span class="prettyTooltip">hello</span>')
			.attr 'class', 'allTooltip'
		@passedTooltip = d3.select(@element[0]).append('xhtml:div')
			.html('<span class="prettyTooltip">hello</span>')
			.attr 'class', 'passedTooltip'
		@failedTooltip = d3.select(@element[0]).append('xhtml:div')
			.html('<span class="prettyTooltip">hello</span>')
			.attr 'class', 'failedTooltip'

	_computeChangeLine: (x, y, allIntervals) ->
		return d3.svg.line()
			.defined((d) -> return not isNaN d)
			.x((d, index) -> return x allIntervals[index])
			.y((d) -> return y d)

	_getLineFromChangeType: (changeType) =>
		if changeType is 'all' then return @allLine
		else if changeType is 'passed' then return @passedLine
		else if changeType is 'failed' then return @failedLine
		else throw 'unexpected change type: ' + changeType

	_getDotsFromChangeType: (changeType) =>
		if changeType is 'all' then return @allLineDots
		else if changeType is 'passed' then return @passedLineDots
		else if changeType is 'failed' then return @failedLineDots
		else throw 'unexpected change type: ' + changeType		

	_getTooltipFromChangeType: (changeType) =>
		if changeType is 'all' then return @allTooltip
		else if changeType is 'passed' then return @passedTooltip
		else if changeType is 'failed' then return @failedTooltip
		else throw 'unexpected change type: ' + changeType	

	_updatePath: (changeType, data, x, y, allIntervals, startFromZero, transitionTime) =>
		line = @_getLineFromChangeType changeType

		line.attr 'display', 'inline'

		previouslyNoData = not line.datum()?
		line = line.datum(data)
		if previouslyNoData or startFromZero
			line = line.attr 'd', @_computeChangeLine x, ((d) -> return y 0), allIntervals
		line = line.transition().duration(transitionTime)
		line = line.attr('d', @_computeChangeLine x, y, allIntervals)

	_updateDots: (changeType, data, x, y, allIntervals, tooltipTextGenerator) =>
		dots = @_getDotsFromChangeType changeType
		tooltip = @_getTooltipFromChangeType changeType

		dots.selectAll('circle').remove()
		newDots = dots.selectAll('circle').data(data)
		newDots.enter()
			.append('circle')
			.attr('class', 'smallDot')
			.attr('cx', (d, index) -> return x allIntervals[index])
			.attr('cy', (d) -> return if isNaN d then 0 else y d)
			.attr('r', (d) -> return if isNaN d then 0 else 1.5)
		newDots.enter()
			.append('circle')
			.attr('class', 'largeDot')
			.attr('cx', (d, index) -> return x allIntervals[index])
			.attr('cy', (d) -> return if isNaN d then 0 else y d)
			.attr('r', (d) -> return if isNaN d then 0 else 5.0)
			.on('mouseover', (d, index) =>
				tooltip
					.style('left', (x(allIntervals[index]) - 130) + 'px')
					.style('top', if isNaN d then '0' else (y(d) - 50) + 'px')
					.classed('visible', true)
				tooltip.select('.prettyTooltip').html tooltipTextGenerator allIntervals[index], allIntervals[index+1], d
			)
			.on('mouseout', (d, index) =>
				tooltip.classed('visible', false)
			)
		newDots.exit()
			.remove()

	updateLine: (changeType, data, x, y, allIntervals, tooltipTextGenerator, startFromZero, transitionTime) =>
		@_updatePath changeType, data, x, y, allIntervals, startFromZero, transitionTime
		@_updateDots changeType, data, x, y, allIntervals, tooltipTextGenerator

	hideLine: (changeType) =>
		line = @_getLineFromChangeType changeType
		line.attr 'display', 'none'

		dots = @_getDotsFromChangeType changeType
		dots.selectAll('circle').remove()
