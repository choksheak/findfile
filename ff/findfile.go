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
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

/**************************************************************************/

// Global variables used during search.

var writeNoisyOutput = func(format string, a ...interface{}) {}
var readableSearchString string
var searchStartTime time.Time
var numDirsRead int64
var numFilesRead int64
var numBytesRead int64
var outputFileHandle *os.File
var outputFileWriter *bufio.Writer
var outputFileInfo os.FileInfo
var currentFilePath string
var currentLineText string
var contextLineIntArrayTempBuffer []int
var matchingLineIntArrayTempBuffer []int
var outputFileBaseName string
var currentMatchesPhrase string

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
	prepareReadableSearchString()
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

func prepareReadableSearchString() {
	if optionListAll.value {
		return
	}

	// This string is for printing to output only.
	if len(searchStringArgs) == 1 {
		readableSearchString = "\"" + searchStringArgs[0] + "\""
	} else {
		readableSearchString = "\"" + strings.Join(searchStringArgs, "\" + \"") + "\""
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
	if !isOutputFormatStringBeginWithNewLine() || showFileNamesOnly {
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
				if !shouldIncludeDirByNameFilters(fileInfo.Name()) {
					continue
				}
			} else {
				// Always skip the output file.
				if (outputFileInfo != nil) && os.SameFile(outputFileInfo, fileInfo) {
					continue
				}

				if !shouldIncludeFileByNameFilters(fileInfo.Name()) {
					continue
				}
			}
			traverseList.PushBack(fileOrDirWithDepth{fileInfo: fileInfo, depth: newDepth, path: newPath})
		}
	}
}

func isLink(path string) bool {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return true
	}
	return (fileInfo.Mode() & os.ModeSymlink) == os.ModeSymlink
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
		matched := searchPathName(isDir)

		// Don't double-print the same filename.
		// The side effect of this is that once the dir or file name matches,
		// we will not show the match count within the file content:
		// e.g. ff -2 txt
		if showFileNamesOnly && matched {
			return
		}
	}

	if !isDir && !optionSearchNamesOnly.value {
		searchFileContents()
	}
}

/**************************************************************************/

// Searching file and dir names.

func searchPathName(isDir bool) bool {
	baseName := filepath.Base(currentFilePath)

	// Check for match.
	checkLineMatchFullInfo(baseName, &currentLineIntArray)

	if currentLineMatchIndexInfo.matched == optionInvertMatch.value {
		return false
	}

	currentMatchCount++

	// Print result.
	writePathNameOutputLine(baseName, "(skip content)", isDir)
	return true
}

/**************************************************************************/

// Searching within files.

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
		if isLineMatchingWithFullInfo(currentLineText, &currentLineIntArray) {
			currentMatchCount++

			// This optimization if-statement is to avoid initializing currentLineIntArray twice.
			// A bit hard to understand, but cannot think of a better approach right now.
			if needMatchDecorations {
				currentLineIntArray = insertMatchDecorations(currentLineIntArray[:0], currentLineText)
			} else if len(currentLineIntArray) == 0 {
				currentLineIntArray = appendStringToIntArray(currentLineIntArray, currentLineText)
			}

			// Format matching line.
			transformOutputLine()

			// Output matching line.
			writeFormattedOutputLine()
		}
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
		if isLineMatching(line, &currentLineIntArray) {
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
	writePathNameOutputLine("", strconv.Itoa(numMatches), false)
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
