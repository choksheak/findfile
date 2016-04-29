/*
The MIT License (MIT)

Copyright (c) 2016 Lau, Chok Sheak (for software "findfile")

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	"math"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

/**************************************************************************/

// Types.

type (
	matchIndexSpan struct {
		beginIndex int
		endIndex   int
	}

	matchIndexInfo struct {
		matched      bool
		minIndex     int
		matchIndexes []matchIndexSpan
	}
)

/**************************************************************************/

// Variables.

var (
	searchStringArgs              []string
	searchStringArgsToUse         []string
	searchStringArgsToUseCount    []int
	searchStringIntArrayToUse     [][]int
	searchStringArgsToExclude     []string
	searchStringIntArrayToExclude [][]int
	searchStringRegexesToUse      []*regexp.Regexp
	searchStringRegexesToExclude  []*regexp.Regexp

	fileNameIncludeFilters []string
	fileNameExcludeFilters []string
	dirNameIncludeFilters  []string
	dirNameExcludeFilters  []string

	currentLineMatchIndexInfo = matchIndexInfo{matchIndexes: make([]matchIndexSpan, 0, 20)}
)

/**************************************************************************/

// Prepare data used for matching.

func prepareSearchString() {
	if optionListAll.value {
		return
	}

	if len(nonOptionArguments) == 0 {
		printDefaultMessage()
		exit(1)
	}

	searchStringArgs = nonOptionArguments

	// Handle duplicate search strings.
	ss := make([]string, len(searchStringArgs))
	if optionIgnoreCase.value {
		for pos, s := range searchStringArgs {
			ss[pos] = strings.ToLower(s)
		}
	} else {
		copy(ss, searchStringArgs)
	}

	// Form array without duplicate strings and count of each string.
	searchStringArgsToUse = make([]string, 0, len(ss))
	searchStringArgsToUseCount = make([]int, 0, len(ss))

	for i := 0; i < len(ss); i++ {
		if ss[i] == "" {
			continue
		}
		count := 1
		for j := i + 1; j < len(ss); j++ {
			if ss[j] == "" {
				continue
			}
			if ss[i] == ss[j] {
				count++
				ss[j] = ""
			}
		}
		searchStringArgsToUse = append(searchStringArgsToUse, ss[i])
		searchStringArgsToUseCount = append(searchStringArgsToUseCount, count)
	}

	// Convert into integer array.
	searchStringIntArrayToUse = make([][]int, len(searchStringArgsToUse))

	for i := 0; i < len(searchStringArgsToUse); i++ {
		searchStringIntArrayToUse[i] = stringToIntArray(searchStringArgsToUse[i])
	}

	// Exclude strings.
	if optionExcludeStrings.value != "" {
		searchStringArgsToExclude = splitAndTrim(optionExcludeStrings.value, ";")

		if optionIgnoreCase.value {
			for i := 0; i < len(searchStringArgsToExclude); i++ {
				searchStringArgsToExclude[i] = strings.ToLower(searchStringArgsToExclude[i])
			}
		}

		searchStringIntArrayToExclude = make([][]int, len(searchStringArgsToExclude))
		for i := 0; i < len(searchStringArgsToExclude); i++ {
			searchStringIntArrayToExclude[i] = stringToIntArray(searchStringArgsToExclude[i])
		}
	}
}

func splitAndTrim(str, delimiter string) []string {
	a := strings.Split(str, delimiter)
	array := make([]string, 0, len(a))
	for i := 0; i < len(a); i++ {
		a[i] = strings.TrimSpace(a[i])

		// Remove empty strings.
		if a[i] != "" {
			array = append(array, a[i])
		}
	}
	return array
}

func prepareNameIncludeExcludeFilters() {
	if !hasAnyIncludeExcludeNameFilters() {
		return
	}

	fileNameIncludeFilters = []string{}
	fileNameExcludeFilters = []string{}
	dirNameIncludeFilters = []string{}
	dirNameExcludeFilters = []string{}

	// Include files.
	if optionIncludeFiles.value != "" {
		fileNameIncludeFilters = append(fileNameIncludeFilters, splitAndTrim(optionIncludeFiles.value, ";")...)
	}

	if optionIncludeDirs.value != "" {
		dirNameIncludeFilters = append(dirNameIncludeFilters, splitAndTrim(optionIncludeDirs.value, ";")...)
	}

	// Exclude files.
	if optionExcludeFiles.value != "" {
		fileNameExcludeFilters = append(fileNameExcludeFilters, splitAndTrim(optionExcludeFiles.value, ";")...)
	}

	if optionExcludeDirs.value != "" {
		dirNameExcludeFilters = append(dirNameExcludeFilters, splitAndTrim(optionExcludeDirs.value, ";")...)
	}
}

func prepareRegexMatching() {
	if !optionRegex.value {
		return
	}

	searchStringRegexesToUse = convertToRegexArray(searchStringArgsToUse)

	if searchStringArgsToExclude != nil {
		searchStringRegexesToExclude = convertToRegexArray(searchStringArgsToExclude)
	}
}

func convertToRegexArray(array []string) []*regexp.Regexp {
	regexes := make([]*regexp.Regexp, len(array))
	for pos, expr := range array {
		if optionWholeWord.value {
			expr = `\b` + expr + `\b`
		}

		regex, err := regexp.Compile(expr)
		if err != nil {
			putln("Invalid regex given: \"%v\"", expr)
			exit(1)
		}
		regexes[pos] = regex
	}
	return regexes
}

/**************************************************************************/

// Directory and file names glob matching.

func hasAnyIncludeExcludeNameFilters() bool {
	return optionIncludeFiles.value != "" ||
		optionIncludeDirs.value != "" ||
		optionExcludeFiles.value != "" ||
		optionExcludeDirs.value != ""
}

func matchesAnyGlob(baseName string, globs []string) bool {
	if len(globs) > 0 {
		for _, glob := range globs {
			matched, _ := filepath.Match(glob, baseName)
			if matched {
				return true
			}
		}
	}
	return false
}

func shouldIncludeFileByNameFilters(fileBaseName string) bool {
	if matchesAnyGlob(fileBaseName, fileNameIncludeFilters) {
		return true
	}

	if matchesAnyGlob(fileBaseName, fileNameExcludeFilters) {
		return false
	}

	if len(fileNameIncludeFilters) > 0 {
		return false
	}

	return true
}

func shouldIncludeDirByNameFilters(dirBaseName string) bool {
	if matchesAnyGlob(dirBaseName, dirNameIncludeFilters) {
		return true
	}

	if matchesAnyGlob(dirBaseName, dirNameExcludeFilters) {
		return false
	}

	if len(dirNameIncludeFilters) > 0 {
		return false
	}

	return true
}

/**************************************************************************/

// Matching utilities.

func isControlCharacter(char rune) bool {
	return (char == 127) || (((0 <= char) && (char <= 31)) && (char != 9) && (char != 10) && (char != 13))
}

func hasControlCharacters(line string) bool {
	for _, char := range line {
		if isControlCharacter(char) {
			return true
		}
	}
	return false
}

func isNonWordChar(char rune) bool {
	return unicode.IsControl(char) ||
		unicode.IsMark(char) ||
		unicode.IsPunct(char) ||
		unicode.IsSpace(char) ||
		unicode.IsSymbol(char)
}

func intArrayHasControlCharacters(array []int) bool {
	for _, char := range array {
		if isControlCharacter(rune(char)) {
			return true
		}
	}
	return false
}

/**************************************************************************/

// Matching logic.

func checkLineMatchFullInfo(line string, lineAsIntArray *[]int) {
	if line == "" {
		return
	}

	if optionIgnoreCase.value {
		line = strings.ToLower(line)
	}

	if optionRegex.value {
		getMatchIndexesByRegexMatch(line)
	} else {
		*lineAsIntArray = appendStringToIntArray((*lineAsIntArray)[:0], line)
		getMatchIndexesByExactMatch(*lineAsIntArray)
	}
}

func isLineMatching(line string, lineAsIntArray *[]int) bool {
	if line == "" {
		return false
	}

	if optionIgnoreCase.value {
		line = strings.ToLower(line)
	}

	if optionRegex.value {
		return isLineMatchingRegex(line)
	}

	*lineAsIntArray = appendStringToIntArray((*lineAsIntArray)[:0], line)
	return isLineMatchingIntArray(*lineAsIntArray)
}

func isLineMatchingWithFullInfo(line string, lineAsIntArray *[]int) bool {
	if line == "" {
		return false
	}

	// This might be initialized as an intended side effect of checkLineMatchFullInfo.
	*lineAsIntArray = (*lineAsIntArray)[:0]

	checkLineMatchFullInfo(line, lineAsIntArray)

	return currentLineMatchIndexInfo.matched != optionInvertMatch.value
}

func resetCurrentLineMatchIndexInfo() {
	// We still need to clear the match indexes when there are multiple search strings.
	currentLineMatchIndexInfo.matched = false
	currentLineMatchIndexInfo.minIndex = math.MaxInt32
	currentLineMatchIndexInfo.matchIndexes = currentLineMatchIndexInfo.matchIndexes[:0]
}

func addMatchSpan(beginIndex, endIndex int) {
	// Add match span.
	currentLineMatchIndexInfo.matchIndexes = append(
		currentLineMatchIndexInfo.matchIndexes,
		matchIndexSpan{beginIndex: beginIndex, endIndex: endIndex})

	// Maintain min and max.
	if currentLineMatchIndexInfo.minIndex > beginIndex {
		currentLineMatchIndexInfo.minIndex = beginIndex
	}
}

// Case-sensitive matching only.
func getMatchIndexesByRegexMatch(line string) {
	resetCurrentLineMatchIndexInfo()

	// Exclude if matches any.
	if searchStringRegexesToExclude != nil {
		for _, regex := range searchStringRegexesToExclude {
			if regex.MatchString(line) {
				return
			}
		}
	}

	// Include if matches all.
	for _, regex := range searchStringRegexesToUse {
		arrayOfIndexes := regex.FindAllStringIndex(line, -1)
		if arrayOfIndexes == nil {
			return
		}
		for _, indexes := range arrayOfIndexes {
			addMatchSpan(indexes[0], indexes[1])
		}
	}

	currentLineMatchIndexInfo.matched = true
}

func isLineMatchingRegex(line string) bool {
	// Exclude if matches any.
	if searchStringRegexesToExclude != nil {
		for _, regex := range searchStringRegexesToExclude {
			if regex.MatchString(line) {
				return false
			}
		}
	}

	// Include if matches all.
	for _, regex := range searchStringRegexesToUse {
		if !regex.MatchString(line) {
			return false
		}
	}

	return true
}

// Case-sensitive matching only.
// This function takes an int array as argument so that we can check word boundary matches.
// Using an int array also saves us from breaking down the string into runes multiple times.
func getMatchIndexesByExactMatch(array []int) {
	resetCurrentLineMatchIndexInfo()

	if searchStringIntArrayToExclude != nil && isMatchingIntArray(array, searchStringIntArrayToExclude, true) {
		return
	}

	for pos, searchStringIntArray := range searchStringIntArrayToUse {
		stringStartIndex := 0
		matchCount := 0
		for {
			beginIndex := intArrayIndexOf(array, searchStringIntArray, stringStartIndex)

			// Match whole words only.
			if beginIndex >= 0 {
				if optionWholeWord.value {
					if (beginIndex > 0) &&
						(!isNonWordChar(rune(array[beginIndex])) &&
							!isNonWordChar(rune(array[beginIndex-1]))) {
						stringStartIndex = beginIndex + len(searchStringIntArray)
						continue
					} else {
						endIndex := beginIndex + len(searchStringIntArray)
						if (endIndex < len(array)) &&
							(!isNonWordChar(rune(array[endIndex])) &&
								!isNonWordChar(rune(array[endIndex-1]))) {
							stringStartIndex = beginIndex + len(searchStringIntArray)
							continue
						}
					}
				}
			}

			if beginIndex < 0 {
				// Match at least the given number of times.
				if matchCount < searchStringArgsToUseCount[pos] {
					// No need to remove from match span because it will not be used anyway.
					return
				}
				break
			}

			endIndex := beginIndex + len(searchStringIntArray)
			addMatchSpan(beginIndex, endIndex)
			matchCount++

			if endIndex >= len(array) {
				// Match at least the given number of times.
				if matchCount < searchStringArgsToUseCount[pos] {
					// No need to remove from match span because it will not be used anyway.
					return
				}
				break
			}
			stringStartIndex = endIndex
		}
	}

	currentLineMatchIndexInfo.matched = true
}

func isLineMatchingIntArray(array []int) bool {
	if searchStringIntArrayToExclude != nil && isMatchingIntArray(array, searchStringIntArrayToExclude, true) {
		return false
	}
	return isMatchingIntArray(array, searchStringIntArrayToUse, false)
}

func isMatchingIntArray(array []int, toFindArray [][]int, matchAny bool) bool {
	for _, searchStringIntArray := range toFindArray {
		stringStartIndex := 0
		for {
			beginIndex := intArrayIndexOf(array, searchStringIntArray, stringStartIndex)

			// Match whole words only.
			if beginIndex >= 0 {
				if optionWholeWord.value {
					if (beginIndex > 0) &&
						(!isNonWordChar(rune(array[beginIndex])) &&
							!isNonWordChar(rune(array[beginIndex-1]))) {
						if matchAny {
							return true
						}
						stringStartIndex = beginIndex + len(searchStringIntArray)
						continue
					} else {
						endIndex := beginIndex + len(searchStringIntArray)
						if (endIndex < len(array)) &&
							(!isNonWordChar(rune(array[endIndex])) &&
								!isNonWordChar(rune(array[endIndex-1]))) {
							stringStartIndex = beginIndex + len(searchStringIntArray)
							if matchAny {
								return true
							}
							continue
						}
					}
				}
				if matchAny {
					return true
				}
				break
			} else {
				if matchAny {
					break
				}
				return false
			}
		}
	}

	if matchAny {
		return false
	}
	return true
}

func intArrayIndexOf(toSearch, toFind []int, startIndex int) int {
	endIndex := len(toSearch) - len(toFind) + 1
Outer:
	for i := startIndex; i < endIndex; i++ {
		for j := 0; j < len(toFind); j++ {
			if toSearch[i+j] != toFind[j] {
				continue Outer
			}
		}
		return i
	}
	return -1
}

/**************************************************************************/
