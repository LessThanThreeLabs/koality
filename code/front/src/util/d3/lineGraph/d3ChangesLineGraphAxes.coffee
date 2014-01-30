window.D3ChangesLineGraphAxes = {}

window.D3ChangesLineGraphAxes.create = (svg, bounds, axisBuffer) ->
	d3ChangesLineGraphAxes = new @clazz svg, bounds, axisBuffer
	d3ChangesLineGraphAxes.initialize()
	return d3ChangesLineGraphAxes

window.D3ChangesLineGraphAxes.clazz = class D3ChangesLineGraphAxes
	constructor: (@svg, @bounds, @axisBuffer) ->
		assert.ok @svg? and typeof @svg is 'object'
		assert.ok @bounds? and typeof @bounds is 'object'
		assert.ok @axisBuffer? and typeof @axisBuffer is 'number'

	initialize: () =>
		@rootPanel = @svg.append 'g'

		@xAxisLabel = @rootPanel.append('g').attr('class', 'xAxis').attr 'transform', "translate(0, #{@bounds.bottom})"
		@yAxisLabel = @rootPanel.append('g').attr('class', 'yAxis').attr 'transform', "translate(#{@bounds.left}, 0)"

		@xReferenceLines = @rootPanel.append('g').attr 'class', 'xReferenceLines'
		@yReferenceLines = @rootPanel.append('g').attr 'class', 'yReferenceLines'

	_getTimeFormat: (d3Binner) ->
		timeInDay = 60 * 60 * 24 * 1000
		startTime = d3Binner.getTimeInterval().start
		endTime = d3Binner.getTimeInterval().end

		hourFormatter = (date) ->
			am = date.getHours() < 12
			hour = date.getHours() % 12
			hour = 12 if hour is 0
			return hour + if am then 'a' else 'p'

		if endTime - startTime <= timeInDay
			return {format: hourFormatter, ticks: 24, subdivide: 0}
		else if endTime - startTime <= 7 * timeInDay
			subdivide = if d3Binner.getIntervalName() is 'hour' then 3 else 0
			return {format: d3.time.format('%a'), ticks: 7, subdivide:  subdivide}
		else if endTime - startTime <= 30 * timeInDay
			ticks = if d3Binner.getIntervalName() is 'day' then 5 else 4
			subdivide = if d3Binner.getIntervalName() is 'day' then 6 else 0
			return {format: d3.time.format('%m/%d'), ticks: ticks, subdivide: subdivide}
		else
			return {format: d3.time.format('%b'), ticks: 12, subdivide: 0}

	_addXReferenceLines: (x, timeFormat) =>
		@xReferenceLines.selectAll('line').remove()
		
		data = x.ticks(timeFormat.ticks)
			.filter (tick) => return x(tick) - @bounds.left > 10 and @bounds.right - x(tick) > 10

		lines = @xReferenceLines.selectAll('line')
			.data(data)
		lines.enter()
			.append('line')
			.attr('x1', x)
			.attr('x2', x)
			.attr('y1', @bounds.bottom - @axisBuffer)
			.attr('y2', @bounds.top)
		lines.exit()
			.remove()

	_addYReferenceLines: (y) =>
		@yReferenceLines.selectAll('line').remove()
		
		lines = @yReferenceLines.selectAll('line')
			.data(y.ticks(5))
		lines.enter()
			.append('line')
			.attr('x1', @bounds.left + @axisBuffer)
			.attr('x2', @bounds.right)
			.attr('y1', y)
			.attr('y2', y)
		lines.exit()
			.remove()

	updateAxisLabels: (d3Binner, x, y, showYAsPercent=false) =>
		timeFormat = @_getTimeFormat d3Binner

		xAxis = d3.svg.axis().scale(x)
			.ticks(timeFormat.ticks)
			.tickSubdivide(timeFormat.subdivide)
			.tickFormat(timeFormat.format)
			.orient 'bottom'

		yAxis = d3.svg.axis().scale(y).ticks(10).orient 'left'
		if showYAsPercent then yAxis.tickFormat d3.format '.0%'
		else yAxis.tickFormat d3.format '.0'

		@xAxisLabel.call xAxis
		@yAxisLabel.call yAxis

		@_addXReferenceLines x, timeFormat
		@_addYReferenceLines y
