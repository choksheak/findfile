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
	"container/list"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"
)

/**************************************************************************/

// Global variables used during search.

var writeNoisyOutput = func(format string, a ...interface{}) {}
var searchStringArgs []string
var readableSearchString string
var searchStringArgsToUse []string
var searchStringArgsToUseCount []int
var searchStringIntArrayToUse [][]int
var searchStringArgsToExclude []string
var searchStringIntArrayToExclude [][]int
var searchStartTime time.Time
var showFileNamesOnly bool
var numDirsRead int64
var numFilesRead int64
var numBytesRead int64
var currentMatchCount int
var outputFileHandle *os.File
var outputFileWriter *bufio.Writer
var outputFileInfo os.FileInfo
var currentFilePath string
var currentLineNumber int
var currentLineText string
var currentLineIntArray = make([]int, 0, 1000)
var contextLineIntArrayTempBuffer []int
var matchingLineIntArrayTempBuffer []int
var beginColorIndexes = make(sort.IntSlice, 0, 20)
var endColorIndexes = make(sort.IntSlice, 0, 20)

type matchIndexSpan struct {
	beginIndex int
	endIndex   int
}

type matchIndexInfo struct {
	matched      bool
	minIndex     int
	matchIndexes []matchIndexSpan
}

var currentLineMatchIndexInfo = matchIndexInfo{matchIndexes: make([]matchIndexSpan, 0, 20)}

var outputFileBaseName string

var contextColumnTruncationMark = stringToIntArray("...")
var contextColumnBeginIndex int
var contextColumnEndIndex int

var fileNameIncludeFilters []string
var fileNameExcludeFilters []string
var dirNameIncludeFilters []string
var dirNameExcludeFilters []string
var searchStringRegexesToUse []*regexp.Regexp
var searchStringRegexesToExclude []*regexp.Regexp

var currentMatchesPhrase string
var outputFormatString = outputFormatDefault
var outputFormatFuncArray []func()

/**************************************************************************/

// Do pre-search actions if any.

func performArgumentActions() {
	needExit := false
	if optionHelp.value {
		printHelp()
		needExit = true
	}
	if optionListOptions.value {
		printListOfOptions()
		needExit = true
	}
	if optionInfo.value {
		printInfo()
		needExit = true
	}
	if optionMarkDown.value {
		printMarkDownInfo()
		needExit = true
	}
	if optionVersion.value {
		printVersion()
		needExit = true
	}
	if optionSetConfig.value != "" {
		setConfigOptions(optionSetConfig.value)
		needExit = true
	}
	if optionUnsetConfig.value != "" {
		unsetConfigOptions(optionUnsetConfig.value)
		needExit = true
	}
	if optionListConfig.value {
		// Already handled during argument parsing stage.
		needExit = true
	}
	if needExit {
		exit(0)
	}
}

/**************************************************************************/

// Search preparation.

func performSearch() {
	setupNoisyOutput()
	prepareSearchString()
	prepareStartingDir()
	setupOutputFile()
	setupContextLines()
	setupContextLineTempBuffer()
	prepareNameIncludeExcludeFilters()
	prepareRegexMatching()
	prepareOutputFormat()
	startTiming()
	startSearching()
	printTiming()
	finalizeOutputFile()
}

func setupNoisyOutput() {
	if !optionQuiet.value {
		writeNoisyOutput = putln
	}
}

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

	// This string is for printing to output only.
	if len(searchStringArgs) == 1 {
		readableSearchString = "\"" + searchStringArgs[0] + "\""
	} else {
		readableSearchString = "\"" + strings.Join(searchStringArgs, "\" + \"") + "\""
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

func prepareStartingDir() {
	// Check that dir exists.
	if optionDir.value != "." {
		exists, err := pathExists(optionDir.value)
		if !exists {
			putln("Given starting dir \"%v\" does not exists: %v", optionDir.value, err)
			exit(1)
		}
	}

	// Convert to absolute path if needed.
	if optionAbsolutePath.value {
		optionDir.value = tryGetAbsolutePath(optionDir.value)
	}
}

func setupOutputFile() {
	// If no need to write output file, do nothing.
	if !optionWriteToFile.value {
		return
	}

	// If output file name is given, check that it is usable.
	if optionOutputFile.value != "" {
		dir := filepath.Dir(optionOutputFile.value)
		if (dir != "") && (dir != ".") {
			exists, _ := pathExists(dir)
			if !exists {
				err := os.MkdirAll(dir, 0755)
				if err != nil {
					putln("Cannot create output dir \"%v\": %v", dir, err)
					exit(1)
				}
			}
		}
	} else {
		// Use default temp dir when output file name is not given.
		dir := tempDir
		optionOutputFile.value = tryGetAbsolutePath(filepath.Join(dir, defaultOutputFileName))
	}

	// Open output file for writing.
	file, err := os.Create(optionOutputFile.value)
	if err != nil {
		putln("Warning: Could not write output file \"%v\": %v", optionOutputFile.value, err)
		return
	}

	// Get the absolute path for the output file.
	absolutePath := tryGetAbsolutePath(optionOutputFile.value)
	if optionAbsolutePath.value {
		optionOutputFile.value = absolutePath
	}

	outputFileHandle = file
	outputFileBaseName = filepath.Base(absolutePath)

	// Add output file to buffered output writer.
	outputFileWriter = bufio.NewWriterSize(file, ioBufferSize)
	addOutputWriter(outputFileWriter)

	// Get the file info to avoid searching it (infinite loop).
	// Ignore errors because we will do anything with it.
	fileInfo, _ := file.Stat()
	outputFileInfo = fileInfo
}

func setupContextLineTempBuffer() {
	if optionContextLines.value == 0 {
		return
	}

	contextLineIntArrayTempBuffer = make([]int, 0, 1000)
	matchingLineIntArrayTempBuffer = make([]int, 0, 1000)
}

func finalizeOutputFile() {
	if outputFileHandle != nil {
		flush()
		removeOutputWriter(outputFileWriter)
		outputFileWriter = nil
		outputFileHandle.Close()
		outputFileHandle = nil

		writeNoisyOutput("Wrote output to file: %v", optionOutputFile.value)
	}

	// Open in external editor if needed.
	if optionSpawn.value {
		cmd := exec.Command(optionEditor.value, optionOutputFile.value)
		err := cmd.Start()
		if err != nil {
			putln("Cannot execute external editor \"%v\": %v", optionEditor.value, err)
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
	if !hasAnyIncludeExcludeFilters() {
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

func prepareOutputFormat() {
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

func writeFormattedOutputLine() {
	for _, fn := range outputFormatFuncArray {
		fn()
	}

	// Flush every search result so that the user can see it immediately.
	flush()
}

func startTiming() {
	if optionMeasureStats.value {
		searchStartTime = time.Now()
	}
}

func printTiming() {
	if optionMeasureStats.value {
		elapsed := time.Since(searchStartTime)
		putln("[time=%v, dirs=%v, files=%v, bytesRead=%v]",
			elapsed,
			numDirsRead,
			numFilesRead,
			addCommasToInt(numBytesRead))
	}
}

/**************************************************************************/

// Search directory tree traversal.

func startSearching() {
	searchType := selectString(optionInvertMatch.value, "non-matches of", "matches of")

	if optionListAll.value {
		searchDir(optionDir.value)
		return
	}

	writeNoisyOutput("%v=== Searching for %v %v in dir: %v ===",
		osNewLine, searchType, readableSearchString, optionDir.value)

	// Make up for one missing newline.
	if !strings.HasPrefix(outputFormatString, "%n") || showFileNamesOnly {
		putBlankLine()
	}

	searchDir(optionDir.value)

	if currentMatchCount == 1 {
		searchType = "match of"
		if optionInvertMatch.value {
			searchType = "non-match of"
		}
	}

	writeNoisyOutput("%v=== Found %v %v %v in dir: %v ===",
		osNewLine, currentMatchCount, searchType, readableSearchString, optionDir.value)
}

type fileOrDirWithDepth struct {
	fileInfo os.FileInfo
	depth    int
	path     string
}

// There is no difference in speed between this algorithm and filepath.Walk().
// filepath.Walk() uses recursion, so this algorithm should be better.
func searchDir(dir string) {
	// Depth-first search using list as stack.
	traverseList := list.New()
	startingDirInfo, err := os.Stat(dir)
	if err != nil {
		putln("Cannot read starting dir \"%v\": %v", dir, err)
		exit(1)
	}
	traverseList.PushBack(fileOrDirWithDepth{fileInfo: startingDirInfo, depth: -1, path: dir})

	for traverseList.Len() > 0 {
		element := traverseList.Back()
		traverseList.Remove(element)
		node := element.Value.(fileOrDirWithDepth)
		if node.depth >= 0 {
			visitFileOrDir(node.path, node.fileInfo)
		}

		// Check max dir depth.
		if (optionMaxLevels.value >= 0) && (node.depth >= optionMaxLevels.value) {
			continue
		}

		// Get list of subdirs and apply inclusion/exclusion filters.
		if !node.fileInfo.IsDir() {
			continue
		}

		fileInfoArray, err := ioutil.ReadDir(node.path)
		if err != nil {
			continue
		}

		newDepth := node.depth + 1
		// Maintain correct order for DFS.
		for i := len(fileInfoArray) - 1; i >= 0; i-- {
			fileInfo := fileInfoArray[i]
			newPath := filepath.Join(node.path, fileInfo.Name())

			// Don't follow symbolic links.
			if isLink(newPath) {
				continue
			}

			if fileInfo.IsDir() {
				if !shouldIncludeDir(fileInfo.Name()) {
					continue
				}
			} else {
				// Always skip the output file.
				if (outputFileInfo != nil) && os.SameFile(outputFileInfo, fileInfo) {
					continue
				}

				if !shouldIncludeFile(fileInfo.Name()) {
					continue
				}
			}
			traverseList.PushBack(fileOrDirWithDepth{fileInfo: fileInfo, depth: newDepth, path: newPath})
		}
	}
}

func hasAnyIncludeExcludeFilters() bool {
	return optionIncludeFiles.value != "" ||
		optionIncludeDirs.value != "" ||
		optionExcludeFiles.value != "" ||
		optionExcludeDirs.value != ""
}

func isLink(path string) bool {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return true
	}
	return (fileInfo.Mode() & os.ModeSymlink) == os.ModeSymlink
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

func shouldIncludeFile(fileBaseName string) bool {
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

func shouldIncludeDir(dirBaseName string) bool {
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

func visitFileOrDir(path string, fileInfo os.FileInfo) {
	currentFilePath = path
	isDir := fileInfo.IsDir()

	if isDir {
		numDirsRead++
	} else {
		numFilesRead++
	}

	if optionListAll.value {
		// Prevent any unwanted string escapes.
		puts(path)
		putBlankLine()
		return
	}

	if !optionSearchContentsOnly.value {
		searchPathName(isDir)
	}

	if !isDir && !optionSearchNamesOnly.value {
		searchFileContents()
	}
}

/**************************************************************************/

// Searching file and dir names.

func searchPathName(isDir bool) {
	baseName := filepath.Base(currentFilePath)

	// Check for match.
	checkLineMatchFullInfo(baseName)

	if currentLineMatchIndexInfo.matched == optionInvertMatch.value {
		return
	}

	currentMatchCount++

	// Format output.
	fileOrDir := selectString(isDir, "dir", "file")

	currentLineNumber = 0
	currentLineMatchIndexInfo.minIndex = -1
	currentLineText = fmt.Sprintf("%v - %v name %v", baseName, fileOrDir, currentMatchesPhrase)
	currentLineIntArray = insertMatchDecorations(currentLineIntArray[:0], currentLineText)

	writeFormattedOutputLine()
}

/**************************************************************************/

// Searching within files.

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

func searchFileContents() {
	// Open file for reading.
	fileHandle, err := os.Open(currentFilePath)
	if err != nil {
		// This output is usually unnecessary, but cannot be sure.
		writeNoisyOutput("Cannot read file %v: %v", currentFilePath, err)
		return
	}
	defer fileHandle.Close()

	setFileForScanning(fileHandle)

	// If file is empty, return.
	if !hasNextLineInFileOrCache() {
		return
	}

	// Only check first line for binary characters.
	currentLineText = getNextLineFromFileOrCache()
	if !optionSearchBinaryFiles.value && hasControlCharacters(currentLineText) {
		return
	}

	// If we want to show filenames only, then we do something special.
	if showFileNamesOnly {
		searchFileContentsForFileNameOnly(currentLineText)
		return
	}

	// Normal search through each line in the file.
	for currentLineNumber = 1; ; currentLineNumber++ {
		checkPrintLineMatch()
		if !hasNextLineInFileOrCache() {
			return
		}
		pushToPreContextLines(currentLineText)
		currentLineText = getNextLineFromFileOrCache()
	}
}

// We need the first line since it was read for determining whether the file is binary or not.
func searchFileContentsForFileNameOnly(firstLine string) {
	line := firstLine
	numMatches := 0

	for {
		if isLineMatching(line) {
			// We could have just stopped after the first match, but printing the
			// total number of matches provides a better user experience.
			numMatches++
		}

		if !hasNextLineInFileOrCache() {
			break
		}

		line = getNextLineFromFileOrCache()
	}

	// Return if not matching.
	hadMatch := (numMatches > 0)
	if hadMatch == optionInvertMatch.value {
		return
	}
	currentMatchCount += numMatches

	// Print result.
	if optionFormat3ShowFileNamesOnly.value {
		puts(currentFilePath)
		putBlankLine()
	} else if optionFormat2ShowFileNamesAndCounts.value {
		putln("%10d : %v", numMatches, currentFilePath)
	} else {
		panic("Unknown output format for filename only")
	}
}

/**************************************************************************/

// Matching logic.

func checkLineMatchFullInfo(line string) {
	if line == "" {
		return
	}

	if optionIgnoreCase.value {
		line = strings.ToLower(line)
	}

	if optionRegex.value {
		getMatchIndexesByRegexMatch(line)
	} else {
		currentLineIntArray = appendStringToIntArray(currentLineIntArray[:0], line)
		getMatchIndexesByExactMatch(currentLineIntArray)
	}
}

func isLineMatching(line string) bool {
	if line == "" {
		return false
	}

	if optionIgnoreCase.value {
		line = strings.ToLower(line)
	}

	if optionRegex.value {
		return isLineMatchingRegex(line)
	}

	currentLineIntArray = appendStringToIntArray(currentLineIntArray[:0], line)
	return isLineMatchingIntArray(currentLineIntArray)
}

func checkPrintLineMatch() {
	if currentLineText == "" {
		return
	}

	// This might be initialized as an intended side effect of checkLineMatchFullInfo.
	currentLineIntArray = currentLineIntArray[:0]

	checkLineMatchFullInfo(currentLineText)

	if currentLineMatchIndexInfo.matched == optionInvertMatch.value {
		return
	}

	currentMatchCount++

	// This optimization if-statement is to avoid initializing currentLineIntArray twice.
	// A bit hard to understand, but cannot think of a better approach right now.
	if needMatchDecorations() {
		currentLineIntArray = insertMatchDecorations(currentLineIntArray[:0], currentLineText)
	} else if len(currentLineIntArray) == 0 {
		currentLineIntArray = appendStringToIntArray(currentLineIntArray, currentLineText)
	}

	// Format matching line.
	transformOutputLine()

	// Output matching line.
	writeFormattedOutputLine()
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

// Output coloring.

func insertMatchDecorations(array []int, line string) []int {
	if !needMatchDecorations() {
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
	if e < numMatches {
		array = appendMatchDecorationsEnd(array)
	}

	return array
}

func needColoring() bool {
	return (optionNoColor.value == false) && isTerminal
}

func needMatchDecorations() bool {
	return (needColoring() || optionShowBrackets.value || optionContextColumns.value != 0) && currentLineMatchIndexInfo.matched
}

func appendMatchDecorationsBegin(array []int) []int {
	if needColoring() {
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
	if needColoring() {
		array = append(array, colorRuneEnd)
	}
	return array
}

/**************************************************************************/

// Format output lines.

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
	return line
}

/**************************************************************************/

// Format context columns.

// Returns a new array with the mapping, and the actual length of the string.
func createLineIndexMapping(line []int) ([]int, int) {
	// Maps the index in line to the actual index of the printed string.
	actualIndexes := make([]int, len(line))
	actualLength := 0
	for pos, char := range line {
		if char >= 0 {
			actualIndexes[pos] = actualLength
			actualLength++
		} else if char == color1RuneBegin {
			// Attach color begin to the next character.
			actualIndexes[pos] = actualLength
		} else if char == colorRuneEnd {
			// Attach color end to the previous character.
			actualIndexes[pos] = actualLength - 1
		}
	}
	return actualIndexes, actualLength
}

func calculateContextColumns(line []int) {
	if optionContextColumns.value == 0 {
		return
	}

	// If the expanded string is already shorter, then return.
	// "line" is an expanded string because it could contain color escape codes,
	// which does not take up any space when printed.
	if len(line) <= optionContextColumns.value {
		contextColumnBeginIndex = 0
		contextColumnEndIndex = optionContextColumns.value
		return
	}

	// Find the actual indexes and length.
	actualIndexes, actualLength := createLineIndexMapping(line)

	// If the printed string is really shorter, then return.
	if actualLength <= optionContextColumns.value {
		contextColumnBeginIndex = 0
		contextColumnEndIndex = optionContextColumns.value
		return
	}

	// Find the first and last actual indexes of the matching subsequence.
	minActualIndex := -1
	for i := 0; i < len(line); i++ {
		if line[i] == color1RuneBegin {
			minActualIndex = actualIndexes[i]
			break
		}
	}
	if minActualIndex == -1 {
		panic(fmt.Sprintf("Begin color missing: line=%v", intArrayToDebugString(line)))
	}

	maxActualIndex := -1
	for i := len(line) - 1; i >= 0; i-- {
		if line[i] == colorRuneEnd {
			maxActualIndex = actualIndexes[i] + 1
			break
		}
	}
	if maxActualIndex == -1 {
		panic(fmt.Sprintf("End color missing: line=%v", intArrayToDebugString(line)))
	}

	contextActualLength := maxActualIndex - minActualIndex

	//putln("calculateContextColumns1: min %v max %v len %v actualLength %v", minActualIndex, maxActualIndex, contextActualLength, actualLength)

	// Expand the context if the matching subsequence is too short.
	if contextActualLength < optionContextColumns.value {
		adjust := (optionContextColumns.value - contextActualLength) / 2
		minActualIndex -= adjust
		maxActualIndex = minActualIndex + optionContextColumns.value
		//putln("adjust %v min %v max %v contextActualLength %v", adjust, minActualIndex, maxActualIndex, contextActualLength)

		if minActualIndex < 0 {
			minActualIndex = 0
			maxActualIndex = optionContextColumns.value
		} else if maxActualIndex > actualLength {
			maxActualIndex = actualLength
			minActualIndex = maxActualIndex - optionContextColumns.value
			if minActualIndex < 0 {
				minActualIndex = 0
			}
		}
	}

	/*
	   putln("calculateContextColumns2: min %v max %v len %v", minActualIndex, maxActualIndex, contextActualLength)
	   for pos, char := range line {
	       puts(fmt.Sprintf("[%c,%v,%v]", char, pos, actualIndexes[pos]))
	   }
	   putBlankLine()
	   flush()
	*/

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
	actualIndexes, actualLength := createLineIndexMapping(line)

	// If really too short, just return.
	if actualLength <= contextColumnBeginIndex {
		return append(line[:0], contextColumnTruncationMark...)
	}

	// Get end index.
	needEndDots := true
	actualEndIndex := contextColumnEndIndex
	if actualEndIndex >= actualLength {
		actualEndIndex = actualLength
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
	for i := 0; i < len(actualIndexes); i++ {
		if contextColumnBeginIndex == actualIndexes[i] {
			beginIndex = i
			break
		}
	}
	if beginIndex == -1 {
		panic(fmt.Sprintf(
			"First fake index missing: contextColumnBeginIndex=%v, actualIndexes=%v, line=%v",
			contextColumnBeginIndex,
			actualIndexes,
			intArrayToDebugString(line)))
	}

	// Convert actual end index into fake end index.
	endIndex := -1
	if actualEndIndex >= actualLength {
		// The endIndex is always one past the last position.
		endIndex = len(line)
	} else {
		for i := len(actualIndexes) - 1; i >= 0; i-- {
			if actualEndIndex == actualIndexes[i] {
				endIndex = i
				break
			}
		}
	}
	if endIndex == -1 {
		panic(fmt.Sprintf(
			"Last fake index missing: actualEndIndex=%v, actualIndexes=%v, line=%v",
			actualEndIndex,
			actualIndexes,
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

// Trap Ctrl+C.

func trapControlC() {
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)
	signal.Notify(channel, syscall.SIGTERM)
	go func() {
		<-channel
		putln("Goodbye")
		cleanUpAndExit(-1)
	}()
}

func cleanUpAndExit(exitCode int) {
	resetColoring()
	exit(exitCode) // implicit flush - make sure everything goes to output before exiting
}

/**************************************************************************/

// Go!

// Main driver program for findfile.
func main() {
	trapControlC()
	loadArguments()
	validateArguments()
	performArgumentActions()
	performSearch()
	cleanUpAndExit(0)
}

/**************************************************************************/
