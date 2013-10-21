package shell

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	AnsiReset  = 0
	AnsiBold   = 1
	AnsiUnBold = 22
)
const (
	AnsiFgBlack = 30 + iota
	AnsiFgRed
	AnsiFgGreen
	AnsiFgYellow
	AnsiFgBlue
	AnsiFgMagenta
	AnsiFgCyan
	AnsiFgWhite
	AnsiFgDefault
)

const (
	AnsiBgBlack = 40 + iota
	AnsiBgRed
	AnsiBgGreen
	AnsiBgYellow
	AnsiBgBlue
	AnsiBgMagenta
	AnsiBgCyan
	AnsiBgWhite
	AnsiBgDefault
)

func AnsiFormat(ansiCodes ...int) string {
	modifierStrings := make([]string, len(ansiCodes))
	for index, ansiCode := range ansiCodes {
		modifierStrings[index] = strconv.Itoa(ansiCode)
	}
	return fmt.Sprintf("\\x1b[%sm", strings.Join(modifierStrings, ";"))
}
