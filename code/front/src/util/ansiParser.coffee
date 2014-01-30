'use strict'

class ConsoleLine

    class ConsoleCharacter
        constructor: (character, style) ->
            @character = character
            @style = new ConsoleStyle style

    class ConsoleStyle
        constructor: (style) ->
            if style?
                foregroundColor = style.foregroundColor
                backgroundColor = style.backgroundColor
                isBold = style.isBold

            @foregroundColor = if foregroundColor? then foregroundColor else 'Default'
            @backgroundColor = if backgroundColor? then backgroundColor else 'Default'
            @isBold = if isBold? then isBold else false

        toCssStyle: () =>
            'foreground' + @foregroundColor + ' background' + @backgroundColor + if @isBold then ' bright' else ''

        equals: (style) =>
            style and @foregroundColor is style.foregroundColor and @backgroundColor is style.backgroundColor and @isBold is style.isBold

    constructor: () ->
        @buffer = []
        @position = 0
        @style = new ConsoleStyle()

    write: (str) =>
        for char, index in str
            if char is '\r'  # Carriage return
                @_carriage_return()
            else
                @buffer[@position] = new ConsoleCharacter char, @style
                @position++

    _carriage_return: () =>
        @position = 0

    setStyles: (styles) =>
        for style in styles
            @setStyle style

    setStyle: (style) =>
        if style.foregroundColor?
            @style.foregroundColor = style.foregroundColor
        if style.backgroundColor?
            @style.backgroundColor = style.backgroundColor
        if style.isBold?
            @style.isBold = style.isBold

    setPosition: (position) =>
        if position > @buffer.length
            for index in [@buffer.length..position]
                @buffer[index] = new ConsoleCharacter ' ', @style
        @position = position

    toString: () =>
        if not @buffer.length
            return ''
        str = ''
        style = null
        cssStyle = null
        for char in @buffer
            if not char.style.equals style
                if cssStyle?
                    str += '</span>'
                style = char.style
                cssStyle = @_stylesToHtml char.style.toCssStyle()
                str += if cssStyle then cssStyle else ''
            str += _escapeString char.character
        str += '</span>' if cssStyle
        return str

    _stylesToHtml: (styles) =>
        if not styles?
            return null
        '<span class="' + styles + '">'


window.ansiParse = (str) ->
    return str if not str?

    if not _stringContainsAnsiCodes str
        return '<span class="foregroundDefault backgroundDefault">' + _escapeString(str) + '</span>'

    consoleLine = new ConsoleLine()
    mode = 'text'
    index = 0
    while index < str.length
        if str[index] is '\x1b' and str[index + 1] is '['  # ESC[ defines a control sequence
            if mode is 'text'  # Begin a control sequence
                mode = 'control code'
                controlCodes = ['']
                index++
            else if mode is 'control code'  # Not sure how we got here, but we'll take the previous control code data as text
                if controlCodes?
                    consoleLine.write controlCodes.join ';'
                index++
        else if mode is 'control code'  # Gather control codes
            switch str[index]
                when ';'  # Another control code is included
                    controlCodes.push ''
                when 'm'  # End the formatter control codes
                    styles = _ansiToStyles controlCodes
                    if styles?
                        consoleLine.setStyles styles
                    mode = 'text'
                when 'G'  # Set column in the line
                    consoleLine.setPosition Number controlCodes[0]
                    mode = 'text'
                else  # Add a character to the last control code
                    controlCodes[controlCodes.length - 1] = controlCodes[controlCodes.length - 1] + str[index]
        else if mode is 'text'  # Just keep the text
            consoleLine.write str[index]
        index++
    return consoleLine.toString()


_stringContainsAnsiCodes = (string) ->
    string.indexOf('\x1b[') isnt -1 or string.indexOf('\r') isnt -1


_escapeString = (string) ->
    return string.replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;')
        .replace(/\//g, '&#x2F;')
        .replace(/\ /g, '&nbsp;')


_ansiToStyles = (controlCodes) ->
    styles = (ansiFormatMap[Number code] for code in controlCodes when code? and code.length > 0 and ansiFormatMap[Number code]?)
    if not styles.length
        return null
    return styles


ansiFormatMap =
    # Clear everything
    0:
        'foregroundColor': 'Default'
        'backgroundColor': 'Default'
        'isBold': false
    # Thickness
    1: 'isBold': true
    22: 'isBold': false
    # Foreground Colors
    30: 'foregroundColor' :'Black'
    31: 'foregroundColor' :'Red'
    32: 'foregroundColor' :'Green'
    33: 'foregroundColor' :'Yellow'
    34: 'foregroundColor' :'Blue'
    35: 'foregroundColor' :'Magenta'
    36: 'foregroundColor' :'Cyan'
    37: 'foregroundColor' :'White'
    39: 'foregroundColor' :'Default'
    # Background Colors
    40: 'backgroundColor': 'Black'
    41: 'backgroundColor': 'Red'
    42: 'backgroundColor': 'Green'
    43: 'backgroundColor': 'Yellow'
    44: 'backgroundColor': 'Blue'
    45: 'backgroundColor': 'Magenta'
    46: 'backgroundColor': 'Cyan'
    47: 'backgroundColor': 'White'
    49: 'backgroundColor': 'Default'
