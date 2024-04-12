package main

import (
	"fmt"
	"regexp"
	"strings"
)

// ANSI escape codes for colors and effects
var ansiCodes = map[string]string{
	"red":       "\033[31m",
	"green":     "\033[32m",
	"yellow":    "\033[33m",
	"blue":      "\033[34m",
	"magenta":   "\033[35m",
	"cyan":      "\033[36m",
	"white":     "\033[37m",
	"bold":      "\033[1m",
	"underline": "\033[4m",
	"blink":     "\033[5m",
	"invert":    "\033[7m",
	"reset":     "\033[0m",
}

// parseTags replaces color and effect tags in a string with ANSI escape codes
func parseTags(s string) string {
	tagRegex := regexp.MustCompile(`\[\&([a-z]+)\]`)
	return tagRegex.ReplaceAllStringFunc(s, func(tag string) string {
		key := strings.Trim(tag, "[]&")
		if code, exists := ansiCodes[key]; exists {
			return code
		}
		return ""
	})
}

// indent adds spaces to the beginning of each line in a string
func indent(s string, padding int) string {
	pad := strings.Repeat(" ", padding*4)
	return pad + strings.ReplaceAll(s, "\n", "\n"+pad)
}

// kvfmt formats a key-value pair with ANSI color and style tags
func kvfmt(key, value string) string {
	return parseTags(fmt.Sprintf("%s[&white]:[&reset] %s", key, value))
}

// logMap logs a map of string keys to string values with formatting
func logMap(m map[string]string, name string) {
	fmt.Println(parseTags(fmt.Sprintf("[&underline][&blue]%s[&reset]", name)))
	for k, v := range m {
		fmt.Println(kvfmt(k, v))
	}
}
