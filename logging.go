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

var logFormats = map[string]string{
	"warn":      "[&white][[&yellow]WARN[&white]][&reset] %s[&reset]",
	"error":     "[&white][[&red]ERRO[&white]][&reset] %s[&reset]",
	"debug":     "[&white][[&blue]DEBG[&white]][&reset] %s[&reset]",
	"key_value": "[&cyan]%s:[&reset]%s\t%s",
	"map":       "[&underline][&blue]%s[&reset]",
	"break":     "",
}

// parseTags replaces color and effect tags in a string with ANSI escape codes
func parseTags(text string) string {
	tagRegex := regexp.MustCompile(`\[&([a-z]+)]`)
	return tagRegex.ReplaceAllStringFunc(text, func(tag string) string {
		key := strings.Trim(tag, "[]&")
		return ansiCodes[key] // if the key does not exist, it will return an empty string
	})
}

// printFormatted prints formatted and tagged strings to the console
func printFormatted(format string, args ...interface{}) {
	fmt.Println(parseTags(fmt.Sprintf(format, args...)))
}

// logMap logs a map of string keys to string values with formatting
func logMap(m map[string]string, name string) {
	maxKeyLen := 0
	for k := range m {
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
	}

	printFormatted(logFormats["map"], name)
	for k, v := range m {
		printFormatted("[&cyan]%s:[&reset]%s\t%s", k, strings.Repeat(" ", maxKeyLen-len(k)), v)
	}
	fmt.Println(ansiCodes["reset"])
}

func logWarning(msg string) {
	printFormatted(logFormats["warn"], msg)
}

func logError(msg string) {
	printFormatted(logFormats["error"], msg)
}

func logBreak() {
	fmt.Println(logFormats["break"])
}
