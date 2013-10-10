package shell

import (
	"fmt"
	"strings"
)

var shellSafeCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@%_-+=:,./"

func Quote(str string) string {
	switch {
	case str == "":
		return "''"
	case containsUnsafeCharacters(str):
		return fmt.Sprintf("'%s'", strings.Replace(str, "'", "'\"'\"'", -1))
	default:
		return str
	}
}

func containsUnsafeCharacters(str string) bool {
	for _, c := range str {
		if !strings.ContainsRune(shellSafeCharacters, c) {
			return true
		}
	}
	return false
}
