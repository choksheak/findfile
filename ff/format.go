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
	"fmt"
	"sort"
	"strconv"
	"strings"
)

/**************************************************************************/

// Variables.

var (
	needColoring                       bool
	needMatchDecorations               bool
	showFileNamesOnly                  bool
	currentMatchCount                  int
	currentLineNumber                  int
	currentLineIntArray                = make([]int, 0, 1000)
	beginColorIndexes                  = make(sort.IntSlice, 0, 20)
	endColorIndexes                    = make(sort.IntSlice, 0, 20)
	contextColumnTruncationMark        = stringToIntArray("...")
	contextColumnBeginIndex            int
	contextColumnEndIndex              int
	contextColumnActualIndexes         []int
	contextColumnActualLength          int
	contextColumnFirstColorActualIndex int
	contextColumnLastColorActualIndex  int
	outputFormatString                 = outputFormatDefault
	outputFormatFuncArray              []func()
)

/**************************************************************************/

// Prepare output formatting.

func prepareOutputFormat() {
	// Init global variables.
	needColoring = (optionNoColor.value == false) && isTerminal
	needMatchDecorations = (!optionInvertMatch.value) &&
		(needColoring || optionShowBrackets.value || optionContextColumns.value != 0)

	// Show filename only will use a different code flow.
	if optionFormat2ShowFileNamesAndCounts.value || optionFormat3ShowFileNamesOnly.value {
		showFileNamesOnly = true
		return
	}

	currentMatchesPhrase = selectString(optionInvertMatch.value, "does not match", "matches")

	// Set format string from arguments.
	if optionFormat0ShowLinesOnly.value {
		outputFormatString = outputFormat0
	} else if optionFormat1ShowFileNamesAndLines.value {
		outputFormatString = outputFormat1
	} else if optionFormat.value != "" {
		outputFormatString = optionFormat.value
	}

	// Compile string to make sure it is valid.
	funcs := []func(){}
	escape := false

	for _, char := range outputFormatString {
		// Escape sequence with %.
		if char == '%' {
			if escape {
				funcs = append(funcs, func() {
					putc('%')
				})
				escape = false
			} else {
				escape = true
			}
			continue
		}

		// Append literal character as-is.
		if !escape {
			c := char
			funcs = append(funcs, func() {
				putc(c)
			})
			continue
		}

		// Interpret escape sequence.
		switch char {
		case 'i':
			funcs = append(funcs, func() {
				puts(strconv.Itoa(currentMatchCount))
			})
		case 'p':
			funcs = append(funcs, func() {
				puts(currentFilePath)
			})
		case 'l':
			funcs = append(funcs, func() {
				puts(strconv.Itoa(currentLineNumber))
			})
		case 'c':
			funcs = append(funcs, func() {
				puts(strconv.Itoa(currentLineMatchIndexInfo.minIndex + 1))
			})
		case 's':
			funcs = append(funcs, func() {
				putIntArrayWithColors(currentLineIntArray)
			})
		case 'n':
			funcs = append(funcs, func() {
				puts(osNewLine)
			})
		default:
			putln("Unrecognized escape sequence %%%c", char)
			exit(1)
		}

		escape = false
	}

	if escape {
		putln("Unterminated '%%' at end of output format string: \"%v\"", outputFormatString)
		exit(1)
	}

	outputFormatFuncArray = funcs
}

/**************************************************************************/

// Writing output line.

func writeFormattedOutputLine() {
	for _, fn := range outputFormatFuncArray {
		fn()
	}

	// Flush every search result so that the user can see it immediately.
	flush()
}

func writePathNameOutputLine(baseName, numMatchesAsString string, isDir bool) {
	if optionFormat3ShowFileNamesOnly.value {
		puts(currentFilePath)
		putBlankLine()
	} else if optionFormat2ShowFileNamesAndCounts.value {
		putln("%15v : %v", numMatchesAsString, currentFilePath)
	} else {
		if baseName == "" {
			panic("Impossible case in show filename only condition")
		}
		currentLineNumber = 0
		currentLineMatchIndexInfo.minIndex = -1
		fileOrDir := selectString(isDir, "dir", "file")
		line := fmt.Sprintf("%v - %v name %v", baseName, fileOrDir, currentMatchesPhrase)
		currentLineIntArray = insertMatchDecorations(currentLineIntArray[:0], line)
		writeFormattedOutputLine()
	}
}

/**************************************************************************/

// Output coloring.

func insertMatchDecorations(array []int, line string) []int {
	if !needMatchDecorations {
		return appendStringToIntArray(array, line)
	}

	// Get marker indexes.
	numMatches := len(currentLineMatchIndexInfo.matchIndexes)

	beginColorIndexes = beginColorIndexes[:0]
	endColorIndexes = endColorIndexes[:0]

	for _, span := range currentLineMatchIndexInfo.matchIndexes {
		beginColorIndexes = append(beginColorIndexes, span.beginIndex)
		endColorIndexes = append(endColorIndexes, span.endIndex)
	}

	beginColorIndexes.Sort()
	endColorIndexes.Sort()

	// Write string to output buffer with color markers.
	b, e := 0, 0

	for pos, char := range line {
		// End color goes first because it is for the previous match.
		if e < numMatches {
			if pos == endColorIndexes[e] {
				array = appendMatchDecorationsEnd(array)
				e++

				// Skip duplicate indexes.
				for e < numMatches && endColorIndexes[e] == pos {
					array = appendMatchDecorationsEnd(array)
					e++
				}
			}
		}

		// Begin color goes next because it is for the next match.
		if b < numMatches {
			if pos == beginColorIndexes[b] {
				array = appendMatchDecorationsBegin(array)
				b++

				// Skip duplicate indexes.
				for b < numMatches && beginColorIndexes[b] == pos {
					array = appendMatchDecorationsBegin(array)
					b++
				}
			}
		}

		array = append(array, int(char))
	}

	// Terminate the color.
	for e < b {
		array = appendMatchDecorationsEnd(array)
		e++
	}

	return array
}

func appendMatchDecorationsBegin(array []int) []int {
	if needColoring {
		array = append(array, color1RuneBegin)
	}
	if optionShowBrackets.value {
		array = append(array, '[')
	}
	return array
}

func appendMatchDecorationsEnd(array []int) []int {
	if optionShowBrackets.value {
		array = append(array, ']')
	}
	if needColoring {
		array = append(array, colorRuneEnd)
	}
	return array
}

/**************************************************************************/

// Format output lines.

func isOutputFormatStringBeginWithNewLine() bool {
	return strings.HasPrefix(outputFormatString, "%n")
}

func transformOutputLine() {
	// Context columns are calculated based on the original line.
	currentLineIntArray = transformSingleOutputLine(currentLineIntArray, true)

	// Simple case if context lines are not needed.
	if optionContextLines.value == 0 {
		return
	}

	// Save the original line to be used later.
	matchingLineIntArrayTempBuffer = append(matchingLineIntArrayTempBuffer[:0], currentLineIntArray...)
	currentLineIntArray = currentLineIntArray[:0]

	// Add pre-context lines.
	for i := optionContextLines.value - 1; i >= 0; i-- {
		contextLine := getContextLineByDelta(-i - 1)
		if contextLine.lineAsStringIsValid {
			j := i + 1
			currentLineIntArray = appendStringToIntArray(
				currentLineIntArray,
				fmt.Sprintf("%v:-%v: ", currentLineNumber-j, j))

			contextLineIntArrayTempBuffer = appendStringToIntArray(contextLineIntArrayTempBuffer[:0], contextLine.lineAsString)
			contextLineIntArrayTempBuffer = transformSingleOutputLine(contextLineIntArrayTempBuffer, false)

			currentLineIntArray = append(currentLineIntArray, contextLineIntArrayTempBuffer...)
			currentLineIntArray = appendStringToIntArray(currentLineIntArray, osNewLine)
		}
	}

	// Add the matching line itself.
	currentLineIntArray = appendStringToIntArray(
		currentLineIntArray,
		fmt.Sprintf("%v: 0: ", currentLineNumber))
	currentLineIntArray = append(currentLineIntArray, matchingLineIntArrayTempBuffer...)
	currentLineIntArray = appendStringToIntArray(currentLineIntArray, osNewLine)

	// Pre-read post-context lines from the file if needed.
	fillPostContextLines()

	// Add post-context lines.
	for i := 0; i < optionContextLines.value; i++ {
		contextLine := getContextLineByDelta(i + 1)
		if !contextLine.lineAsStringIsValid {
			break
		}
		j := i + 1
		currentLineIntArray = appendStringToIntArray(
			currentLineIntArray,
			fmt.Sprintf("%v:+%v: ", currentLineNumber+j, j))

		contextLineIntArrayTempBuffer = appendStringToIntArray(contextLineIntArrayTempBuffer[:0], contextLine.lineAsString)
		contextLineIntArrayTempBuffer = transformSingleOutputLine(contextLineIntArrayTempBuffer, false)

		currentLineIntArray = append(currentLineIntArray, contextLineIntArrayTempBuffer...)
		currentLineIntArray = appendStringToIntArray(currentLineIntArray, osNewLine)
	}
}

func transformSingleOutputLine(line []int, needCalculateContextColumns bool) []int {
	// The ordering here is important:
	// 1. Tabs will expand before control characters are removed.
	// 2. Control characters are removed before we extract the column context window.
	line = transformTabs(line)
	line = replaceControlCharacters(line)

	if needCalculateContextColumns {
		calculateContextColumns(line)
	}

	line = truncateToContextColumns(line)

	// Debug.
	//putln(intArrayToDebugString(line))

	return line
}

/**************************************************************************/

// Format context columns.

// Returns a new array with the mapping, and the actual length of the string.
func createContextColumnsLineIndexMapping(line []int) {
	// No need to recalculate for the same line.
	if contextColumnActualIndexes != nil {
		return
	}

	// Maps the index in line to the actual index of the printed string.
	contextColumnActualIndexes = make([]int, len(line))
	contextColumnActualLength = 0
	contextColumnFirstColorActualIndex = -1
	contextColumnLastColorActualIndex = -1

	for pos, char := range line {
		if char >= 0 {
			contextColumnActualIndexes[pos] = contextColumnActualLength
			contextColumnActualLength++
		} else if char == color1RuneBegin {
			// Attach color begin to the next character.
			contextColumnActualIndexes[pos] = contextColumnActualLength
			if contextColumnFirstColorActualIndex < 0 {
				contextColumnFirstColorActualIndex = contextColumnActualLength
			}
		} else if char == colorRuneEnd {
			// Attach color end to the previous character.
			contextColumnActualIndexes[pos] = contextColumnActualLength - 1
			contextColumnLastColorActualIndex = contextColumnActualLength
		}
	}
}

func calculateContextColumns(line []int) {
	if optionContextColumns.value == 0 {
		return
	}

	// Reset to indicate it was not calculated yet.
	contextColumnActualIndexes = nil

	// If the expanded string is already shorter, then return.
	// "line" is an expanded string because it could contain color escape codes,
	// which does not take up any space when printed.
	if len(line) <= optionContextColumns.value {
		contextColumnBeginIndex = 0
		contextColumnEndIndex = optionContextColumns.value
		//putln("calculateContextColumns0: min %v max %v len %v", contextColumnBeginIndex, contextColumnEndIndex, contextColumnEndIndex-contextColumnBeginIndex)
		return
	}

	// Find the actual indexes and length.
	createContextColumnsLineIndexMapping(line)

	// If the printed string is really shorter, then return.
	if contextColumnActualLength <= optionContextColumns.value {
		contextColumnBeginIndex = 0
		contextColumnEndIndex = optionContextColumns.value
		//putln("calculateContextColumns1: min %v max %v len %v", contextColumnBeginIndex, contextColumnEndIndex, contextColumnEndIndex-contextColumnBeginIndex)
		return
	}

	// If matched string is longer than context column, shrink to matched string.
	if contextColumnFirstColorActualIndex >= 0 &&
		contextColumnLastColorActualIndex >= 0 &&
		contextColumnLastColorActualIndex-contextColumnFirstColorActualIndex >= optionContextColumns.value {

		contextColumnBeginIndex = contextColumnFirstColorActualIndex
		contextColumnEndIndex = contextColumnLastColorActualIndex
		//putln("calculateContextColumns2: min %v max %v len %v", contextColumnBeginIndex, contextColumnEndIndex, contextColumnEndIndex-contextColumnBeginIndex)
		return
	}

	// Find the first and last actual indexes of the matching subsequence.
	var minActualIndex int
	var maxActualIndex int

	if needMatchDecorations {
		minActualIndex = contextColumnFirstColorActualIndex
		if minActualIndex == -1 {
			panic(fmt.Sprintf("Begin color missing: line=%v", intArrayToDebugString(line)))
		}

		maxActualIndex = contextColumnLastColorActualIndex
		if maxActualIndex == -1 {
			panic(fmt.Sprintf("End color missing: line=%v", intArrayToDebugString(line)))
		}
	} else {
		// Get first N characters for non-matches.
		minActualIndex = 0
		maxActualIndex = len(line)

		if maxActualIndex > optionContextColumns.value {
			maxActualIndex = optionContextColumns.value
		}
	}

	contextActualLength := maxActualIndex - minActualIndex

	//putln("calculateContextColumns3: min %v max %v len %v actualLength %v", minActualIndex, maxActualIndex, contextActualLength, contextColumnActualLength)

	// Expand the context if the matching subsequence is too short.
	if contextActualLength < optionContextColumns.value {
		adjust := (optionContextColumns.value - contextActualLength) / 2
		minActualIndex -= adjust
		maxActualIndex = minActualIndex + optionContextColumns.value
		//putln("adjust %v min %v max %v contextActualLength %v", adjust, minActualIndex, maxActualIndex, contextActualLength)

		if minActualIndex < 0 {
			minActualIndex = 0
			maxActualIndex = optionContextColumns.value
		} else if maxActualIndex > contextColumnActualLength {
			maxActualIndex = contextColumnActualLength
			minActualIndex = maxActualIndex - optionContextColumns.value
			if minActualIndex < 0 {
				minActualIndex = 0
			}
		}
	}

	if false {
		putln("calculateContextColumns4: min %v max %v len %v", minActualIndex, maxActualIndex, contextActualLength)
		for pos, char := range line {
			puts(fmt.Sprintf("[%c,%v,%v]", char, pos, contextColumnActualIndexes[pos]))
		}
		putBlankLine()
		flush()
	}

	contextColumnBeginIndex = minActualIndex
	contextColumnEndIndex = maxActualIndex
}

func truncateToContextColumns(line []int) []int {
	if optionContextColumns.value == 0 {
		return line
	}

	// Line is empty, so return empty.
	if len(line) == 0 {
		return line[:0]
	}

	// Line is too short, so make it empty.
	if len(line) <= contextColumnBeginIndex {
		return append(line[:0], contextColumnTruncationMark...)
	}

	// Find the actual indexes and length.
	createContextColumnsLineIndexMapping(line)

	// If really too short, just return.
	if contextColumnActualLength <= contextColumnBeginIndex {
		return append(line[:0], contextColumnTruncationMark...)
	}

	// Get end index.
	needEndDots := true
	actualEndIndex := contextColumnEndIndex
	if actualEndIndex >= contextColumnActualLength {
		actualEndIndex = contextColumnActualLength
		needEndDots = false

		// If we just need the whole string, just return now.
		if contextColumnBeginIndex == 0 {
			return line
		}
	}

	// Find begin index.
	contextBeginIndex := 0
	if contextColumnBeginIndex > 0 {
		contextBeginIndex = len(contextColumnTruncationMark)
	}

	// Convert actual begin index into fake begin index.
	beginIndex := -1
	for i := 0; i < len(contextColumnActualIndexes); i++ {
		if contextColumnBeginIndex == contextColumnActualIndexes[i] {
			beginIndex = i
			break
		}
	}
	if beginIndex == -1 {
		panic(fmt.Sprintf(
			"First fake index missing: contextColumnBeginIndex=%v, actualIndexes=%v, line=%v",
			contextColumnBeginIndex,
			contextColumnActualIndexes,
			intArrayToDebugString(line)))
	}

	// Convert actual end index into fake end index.
	endIndex := -1
	if actualEndIndex >= contextColumnActualLength {
		// The endIndex is always one past the last position.
		endIndex = len(line)
	} else {
		for i := len(contextColumnActualIndexes) - 1; i >= 0; i-- {
			if actualEndIndex == contextColumnActualIndexes[i] {
				endIndex = i
				break
			}
		}
	}
	if endIndex == -1 {
		panic(fmt.Sprintf(
			"Last fake index missing: actualEndIndex=%v, actualIndexes=%v, line=%v",
			actualEndIndex,
			contextColumnActualIndexes,
			intArrayToDebugString(line)))
	}

	/*
		putln("truncateToContextColumns1: begin %v end %v endIndex %v len(line) %v: %v",
			contextColumnBeginIndex, contextColumnEndIndex, endIndex, len(line), intArrayToString(line))
		for pos, char := range line {
			puts(fmt.Sprintf("[%c,%v,%v]", char, pos, actualIndexes[pos]))
		}
		putBlankLine()
		flush()
	*/

	length := endIndex - beginIndex + contextBeginIndex
	copy(line[contextBeginIndex:length], line[beginIndex:endIndex])
	line = line[0:length]

	if contextBeginIndex != 0 {
		copy(line[0:contextBeginIndex], contextColumnTruncationMark)
	}

	if needEndDots {
		line = append(line, contextColumnTruncationMark...)
	}

	return line
}

/**************************************************************************/

// Format tabs.

func intArrayContainsTabs(array []int) bool {
	for _, char := range array {
		if char == '\t' {
			return true
		}
	}
	return false
}

func transformTabs(array []int) []int {
	if !intArrayContainsTabs(array) {
		return array
	}

	col := 0
	origArray := make([]int, len(array))
	copy(origArray, array)
	array = array[:0]

	for _, char := range origArray {
		if char == '\t' {
			array = appendTab(array, optionTabSpacing.value-col)
			col = 0
			continue
		}

		array = append(array, char)

		// Don't count color markers.
		if char >= 0 {
			col++
			if col >= optionTabSpacing.value {
				col = 0
			}
		}
	}

	return array
}

func appendTab(array []int, tabSize int) []int {
	if optionShowTabs.value {
		for i := tabSize - 1; i > 0; i-- {
			array = append(array, int('-'))
		}
		array = append(array, int('>'))
	} else {
		for i := tabSize; i > 0; i-- {
			array = append(array, int(' '))
		}
	}
	return array
}

/**************************************************************************/

// Format control characters.

func replaceControlCharacters(array []int) []int {
	if optionShowControlChars.value {
		return array
	}

	if !intArrayHasControlCharacters(array) {
		return array
	}

	newArray := make([]int, 0, len(array))
	copy(newArray, array)
	array = array[:0]

	for _, char := range newArray {
		if isControlCharacter(rune(char)) {
			// Retain the same spacing to avoid confusion in the output.
			array = append(array, ' ')
		} else {
			array = append(array, char)
		}
	}

	return array
}

/**************************************************************************/
