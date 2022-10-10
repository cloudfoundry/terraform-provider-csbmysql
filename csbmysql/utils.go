package csbmysql

import (
	"fmt"
	"strings"
)

func quotedIdentifier(identifier string) string {
	return escapeStringEnclosingCharacter(identifier, "`")
}

func quotedString(originalString string) string {
	return escapeStringEnclosingCharacter(originalString, "'")
}

func escapeStringEnclosingCharacter(originalString string, character string) string {
	return fmt.Sprintf("%[1]s%[2]s%[1]s", character, strings.NewReplacer(character, character+character).Replace(originalString))
}
