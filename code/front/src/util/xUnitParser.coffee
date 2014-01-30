'use strict'

getAttribute = (testCaseString, attributeName) ->
    assert.ok typeof testCaseString is 'string'
    assert.ok typeof attributeName is 'string'

    tagStartIndex = testCaseString.indexOf(" #{attributeName}=\"")
    return null if tagStartIndex is -1

    nameStartIndex = tagStartIndex + " #{attributeName}=\"".length
    nameEndIndex = testCaseString.indexOf '"', nameStartIndex

    if nameStartIndex is -1 or nameEndIndex is -1 then return null
    else return testCaseString.substring nameStartIndex, nameEndIndex

extractTextFromCData = (text) ->
    assert.ok typeof text is 'string'

    cDataStartTag = '<![CDATA['
    cDataEndTag = ']]>'

    return text if text.indexOf(cDataStartTag) is -1 or text.indexOf(cDataEndTag) is -1

    textStartIndex = text.indexOf(cDataStartTag) + cDataStartTag.length
    textEndIndex = text.indexOf cDataEndTag

    return text.substring textStartIndex, textEndIndex

getTextInElement = (testCaseString, elementName) ->
    assert.ok typeof testCaseString is 'string'
    assert.ok typeof elementName is 'string'

    elementStartIndex = testCaseString.indexOf("<#{elementName}")
    return null if elementStartIndex is -1

    textStartIndex = testCaseString.indexOf('>', elementStartIndex) + 1
    textEndIndex = testCaseString.indexOf("</#{elementName}>", textStartIndex)

    text = testCaseString.substring textStartIndex, textEndIndex
    return extractTextFromCData text

getTestCaseString = (xunitOutput, startIndex) ->
    assert.ok typeof xunitOutput is 'string'
    assert.ok typeof startIndex is 'number'

    testCaseStartTag = '<testcase '
    testCaseEndTag = '</testcase>'
    testCaseSelfClose = '/>'

    testCaseStartIndex = xunitOutput.indexOf testCaseStartTag, startIndex
    return null if testCaseStartIndex is -1

    testCaseClosingTagIndex = xunitOutput.indexOf testCaseEndTag, testCaseStartIndex
    testCaseSelfClosingTagIndex = xunitOutput.indexOf testCaseSelfClose, testCaseStartIndex

    testCaseEndIndex = null
    if testCaseSelfClosingTagIndex isnt -1 and testCaseClosingTagIndex is -1
        testCaseEndIndex = testCaseSelfClosingTagIndex + testCaseSelfClose.length
    else if testCaseSelfClosingTagIndex is -1 and testCaseClosingTagIndex isnt -1
        testCaseEndIndex = testCaseClosingTagIndex + testCaseEndTag.length
    else if testCaseSelfClosingTagIndex isnt -1 and testCaseClosingTagIndex isnt -1
        testCaseEndIndex = Math.min testCaseSelfClosingTagIndex + testCaseSelfClose.length,
            testCaseClosingTagIndex + testCaseEndTag.length
    else
        throw 'Nonexistent closing tag'

    toReturn =
        start: testCaseStartIndex
        end: testCaseEndIndex
        text: xunitOutput.substring testCaseStartIndex, testCaseEndIndex
    return toReturn

window.xUnitParse = (xunitOutput) ->
    assert.ok typeof xunitOutput is 'string'
    
    currentTestCaseString = getTestCaseString xunitOutput, 0

    testCases = []
    while currentTestCaseString?
        className = getAttribute currentTestCaseString.text, 'classname'
        testName = getAttribute currentTestCaseString.text, 'name'

        testCase =
            name: if className? then className + '.' + testName else testName
            time: Number getAttribute currentTestCaseString.text, 'time'
            failure: getTextInElement currentTestCaseString.text, 'failure'
            error: getTextInElement currentTestCaseString.text, 'error'
            sysout: getTextInElement currentTestCaseString.text, 'system-out'
            syserr: getTextInElement currentTestCaseString.text, 'system-err'
        testCase.status = if testCase.failure? or testCase.error? then 'failed' else 'passed'

        testCases.push testCase
        currentTestCaseString = getTestCaseString xunitOutput, currentTestCaseString.end

    return testCases
