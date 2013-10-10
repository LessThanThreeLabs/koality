package shell

import (
	"fmt"
	"strings"
)

var shellSafeCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@%_-+=:,./"

func Quote(str string) string {
	if str == "" {
		return "''"
	}
	for _, c := range str {
		if !strings.ContainsRune(shellSafeCharacters, c) {
			return fmt.Sprintf("'%s'", strings.Replace(str, "'", "'\"'\"'", -1))
		}
	}
	return str
}
