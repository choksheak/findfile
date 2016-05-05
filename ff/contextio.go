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
	"bufio"
	"os"
	"strconv"
)

/**************************************************************************/

// Types.

type (
	contextLines struct {
		startIndex int
		lines      []contextLine
	}

	contextLine struct {
		lineAsString        string
		lineAsStringIsValid bool
	}
)

/**************************************************************************/

// Variables.

var (
	preContextLines         contextLines
	postContextLines        contextLines
	contextLinesFileScanner *bufio.Scanner
)

/**************************************************************************/

// Setup context lines.

func setupContextLines() {
	if optionContextLines.value == 0 {
		return
	}

	preContextLines = createContextLines(optionContextLines.value)
	postContextLines = createContextLines(optionContextLines.value)
}

func createContextLines(numLines int) contextLines {
	newContextLines := contextLines{
		startIndex: 0,
		lines:      make([]contextLine, numLines),
	}

	return newContextLines
}

/**************************************************************************/

// Reading lines and maintaining context lines.

func setFileForScanning(fileHandle *os.File) {
	contextLinesFileScanner = bufio.NewScanner(fileHandle)
}

// This function will read one line whenever it gets called.
func hasNextLineInFile() bool {
	return contextLinesFileScanner.Scan()
}

func getNextLineFromFile() string {
	line := contextLinesFileScanner.Text()
	numBytesRead += int64(len(line))
	return line
}

func addToContextLineIndex(index, delta int) int {
	index += delta
	if index < 0 {
		index += optionContextLines.value
	} else if index >= optionContextLines.value {
		index -= optionContextLines.value
	}
	return index
}

func getContextLineByDelta(delta int) contextLine {
	if (delta < -optionContextLines.value) || (delta > optionContextLines.value) {
		panic("delta out of range: " + strconv.Itoa(delta))
	}
	if delta == 0 {
		panic("delta cannot be zero")
	}

	if delta < 0 {
		index := addToContextLineIndex(preContextLines.startIndex, delta+1)
		return preContextLines.lines[index]
	}

	index := addToContextLineIndex(postContextLines.startIndex, delta-1)
	return postContextLines.lines[index]
}

// Check context lines.
func hasNextLineInFileOrCache() bool {
	if optionContextLines.value == 0 {
		return hasNextLineInFile()
	}

	if postContextLines.lines[postContextLines.startIndex].lineAsStringIsValid {
		return true
	}

	return hasNextLineInFile()
}

// Maintain context lines.
func getNextLineFromFileOrCache() string {
	if optionContextLines.value == 0 {
		return getNextLineFromFile()
	}

	// Read from post context lines.
	if postContextLines.lines[postContextLines.startIndex].lineAsStringIsValid {
		postContextLines.lines[postContextLines.startIndex].lineAsStringIsValid = false
		line := postContextLines.lines[postContextLines.startIndex].lineAsString
		postContextLines.startIndex = addToContextLineIndex(postContextLines.startIndex, 1)
		return line
	}

	return getNextLineFromFile()
}

func pushToPreContextLines(line string) {
	if optionContextLines.value == 0 {
		return
	}
	preContextLines.startIndex = addToContextLineIndex(preContextLines.startIndex, 1)
	preContextLines.lines[preContextLines.startIndex].lineAsStringIsValid = true
	preContextLines.lines[preContextLines.startIndex].lineAsString = line
}

func fillPostContextLines() {
	if optionContextLines.value == 0 {
		return
	}

	i := postContextLines.startIndex
	for {
		if !postContextLines.lines[i].lineAsStringIsValid {
			if !hasNextLineInFile() {
				break
			}

			postContextLines.lines[i].lineAsString = getNextLineFromFile()
			postContextLines.lines[i].lineAsStringIsValid = true
		}

		i = addToContextLineIndex(i, 1)
		if i == postContextLines.startIndex {
			break
		}
	}
}

/**************************************************************************/
