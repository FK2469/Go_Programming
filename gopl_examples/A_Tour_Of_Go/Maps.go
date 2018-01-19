package main

import (
	"strings"

	"golang.org/x/tour/wc"
)

func WordCount(s string) map[string]int {
	ret := make(map[string]int)
	arr := strings.Fields(s)
	// Fields splits the string s around each instance of one or more consecutive white space characters, as defined by unicode.IsSpace, returning an array of substrings of s or an empty list if s contains only white space.
	for _, val := range arr {
		ret[val]++
	}
	return ret
}

func main() {
	wc.Test(WordCount)
}
