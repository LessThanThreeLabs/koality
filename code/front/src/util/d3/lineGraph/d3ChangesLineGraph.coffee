window.D3ChangesLineGraph = {}

window.D3ChangesLineGraph.create = (element) ->
	d3ChangesLineGraph = new @clazz element
	d3ChangesLineGraph.initialize()
	return d3ChangesLineGraph

window.D3ChangesLineGraph.clazz = class D3ChangesLineGraph
	PADDING: {top: 10, left: 35, right: 10, bottom: 25}
	AXIS_BUFFER: 10

	constructor: (@element) ->
		assert.ok @element? and typeof @element is 'object'

	initialize: () =>
		@bounds =
			top: @PADDING.top
			left: @PADDING.left
			right: @element.width() - @PADDING.right
			bottom: @element.height() - @PADDING.bottom

		@svg = d3.select(@element[0]).select('svg')
			.attr('width', '100%')
			.attr('height', '100%')
		@root = @svg.append 'g'
		@axes = D3ChangesLineGraphAxes.create @root, @bounds, @AXIS_BUFFER
		@lines = D3ChangesLineGraphLines.create @element, @root

	_getTooltipTextGenerator: (d3Binner, changeType, isPercentage) =>
		return (firstDate, secondDate, value) ->
			getFirstLine = () ->
				secondDate ?= d3Binner.getTimeInterval().end
				if d3Binner.getIntervalName() is 'hour'
					return d3.time.format('%a %m/%d, %I:%M')(firstDate) + ' to ' + d3.time.format('%I:%M %p')(secondDate)
				else
					return d3.time.format('%a %m/%d')(firstDate) + ' to ' + d3.time.format('%a %m/%d')(secondDate)

			getSecondLine = () ->
				changeTag = if changeType is 'all' then 'total' else changeType
				if isPercentage then return d3.format('.0%')(value) + ' of changes ' + changeTag
				else return value + ' changes ' + changeTag

			return getFirstLine() + '<br>' + getSecondLine()

	drawGraph: (d3Binner, startFromZero) =>
		assert.ok d3Binner? and typeof d3Binner is 'object'

		allIntervals = d3Binner.getAllIntervals()
		histograms = d3Binner.getHistograms()

		x = d3.time.scale()
			.domain([allIntervals[0], allIntervals[allIntervals.length-1]])
			.range([@bounds.left + @AXIS_BUFFER, @bounds.right])
		y = d3.scale.linear()
			.domain([0, d3.max histograms.all])
			.range([@bounds.bottom - @AXIS_BUFFER, @bounds.top])

		allTooltipTextGenerator = @_getTooltipTextGenerator d3Binner, 'all', false
		@lines.updateLine 'all', histograms.all, x, y, allIntervals, allTooltipTextGenerator, startFromZero, 500

		passedTooltipTextGenerator = @_getTooltipTextGenerator d3Binner, 'passed', false
		@lines.updateLine 'passed', histograms.passed, x, y, allIntervals, passedTooltipTextGenerator, startFromZero, 750

		failedTooltipTextGenerator = @_getTooltipTextGenerator d3Binner, 'failed', false
		@lines.updateLine 'failed', histograms.failed, x, y, allIntervals, failedTooltipTextGenerator, startFromZero, 1000

		@axes.updateAxisLabels d3Binner, x, y

	drawPercentageGraph: (d3Binner, changeType, startFromZero=true) =>
		assert.ok d3Binner? and typeof d3Binner is 'object'
		assert.ok changeType? and changeType is 'passed' or changeType is 'failed'

		percentageHistograms = d3Binner.getPercentageHistograms()
		allIntervals = d3Binner.getAllIntervals()

		# We remove intervals that belong to NaN values so the
		# graph smoothes itself out, except for the endpoints
		fixHistogramAndIntervals = () ->
			isValidHistogramValue = (histogram) ->
				return (value, index) ->
					return true if index is 0 or index is histogram.length - 1
					return not isNaN value

			allIntervals = allIntervals.filter (interval, index) ->
				return true if index is 0 or index is allIntervals.length - 1
				if changeType is 'passed' then return not isNaN percentageHistograms.passed[index]
				if changeType is 'failed' then return not isNaN percentageHistograms.failed[index]

			percentageHistograms.passed = percentageHistograms.passed.filter isValidHistogramValue percentageHistograms.passed
			percentageHistograms.failed = percentageHistograms.failed.filter isValidHistogramValue percentageHistograms.failed

		fixHistogramAndIntervals()

		x = d3.time.scale()
			.domain([allIntervals[0], allIntervals[allIntervals.length-1]])
			.range([@bounds.left + @AXIS_BUFFER, @bounds.right])
		y = d3.scale.linear()
			.domain([0, 1])
			.range([@bounds.bottom - @AXIS_BUFFER, @bounds.top])

		@lines.hideLine 'all'

		if changeType is 'passed'
			passedTooltipTextGenerator = @_getTooltipTextGenerator d3Binner, 'passed', true
			@lines.updateLine 'passed', percentageHistograms.passed, x, y, allIntervals, passedTooltipTextGenerator, startFromZero, 750
			@lines.hideLine 'failed'

		if changeType is 'failed'
			failedTooltipTextGenerator = @_getTooltipTextGenerator d3Binner, 'failed', true
			@lines.updateLine 'failed', percentageHistograms.failed, x, y, allIntervals, failedTooltipTextGenerator, startFromZero, 750
			@lines.hideLine 'passed'

		@axes.updateAxisLabels d3Binner, x, y, true
