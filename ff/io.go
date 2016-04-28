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
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"path/filepath"
)

/**************************************************************************/

// I/O-related constants.

const (
	// Standard buffer size for our I/O operations.
	ioBufferSize = 4096

	// The largest unicode code takes 6 bytes. We give it 7 just to be safe.
	maxUnicodeCharBytes = 7

	// Common indent to use when printing output text.
	printIndent = "    "
)

/**************************************************************************/

// I/O-related variables.

var (
	// Check if windows or other OS.
	isWindows = (os.PathSeparator == '\\') && (os.PathListSeparator == ';')

	// Temp dir to use.
	tempDir = selectString(isWindows, `c:\temp`, `/tmp`)

	// User's home directory.
	userHomeDir = selectString(isWindows,
		filepath.Join(os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH")),
		os.Getenv("HOME"))

	// Why is this not a system constant in Go??
	osNewLine = selectString(isWindows, "\r\n", "\n")

	// Need to know whether we are in a terminal or not.
	isTerminal = terminal.IsTerminal(int(os.Stdout.Fd()))

	// Writer interface for stdout.
	stdoutWriter = bufio.NewWriterSize(os.Stdout, ioBufferSize)

	// List of buffered I/O wrapper writers.
	outputWriters = []*bufio.Writer{
		stdoutWriter,
	}
)

/**************************************************************************/

// Buffered output.

func addOutputWriter(writer *bufio.Writer) {
	outputWriters = append(outputWriters, writer)
}

func removeOutputWriter(writer *bufio.Writer) {
	for pos, w := range outputWriters {
		if w == writer {
			outputWriters = append(outputWriters[:pos], outputWriters[pos+1:]...)
			return
		}
	}
}

func writeString(writer *bufio.Writer, s string) {
	if len(s) <= writer.Available() {
		writer.WriteString(s)
	} else if len(s) <= ioBufferSize {
		flush()
		writer.WriteString(s)
	} else {
		start := 0
		end := ioBufferSize
		for {
			writer.WriteString(s[start:end])
			flush()

			if end == len(s) {
				break
			}
			start = end
			end += ioBufferSize
			if end > len(s) {
				end = len(s)
			}
		}
	}
}

func puts(s string) {
	for _, writer := range outputWriters {
		writeString(writer, s)
	}
}

func putBlankLine() {
	puts(osNewLine)
}

func putln(format string, a ...interface{}) {
	puts(fmt.Sprintf(format, a...))
	puts(osNewLine)
}

func putc(char rune) {
	for _, writer := range outputWriters {
		writer.WriteRune(char)
	}
}

func flush() {
	for _, writer := range outputWriters {
		writer.Flush()
	}
}

func exit(code int) {
	flush()
	os.Exit(code)
}

/**************************************************************************/

// File utilities.

func tryGetAbsolutePath(path string) string {
	absPath, err := filepath.Abs(path)
	if err == nil {
		return filepath.Clean(absPath)
	}
	return filepath.Clean(path)
}

// Checks whether the given file or directory exists or not.
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func tryCreateDirs(dir string) error {
	return os.MkdirAll(dir, 0755)
}

/**************************************************************************/
