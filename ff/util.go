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
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
)

/**************************************************************************/

// Utilities that don't fit anywhere else.

func selectString(which bool, trueString, falseString string) string {
	if which {
		return trueString
	}
	return falseString
}

func intArrayToString(array []int) string {
	var buffer bytes.Buffer
	for _, char := range array {
		buffer.WriteRune(rune(char))
	}
	return buffer.String()
}

func intArrayToDebugString(array []int) string {
	var buffer bytes.Buffer
	for pos, char := range array {
		buffer.WriteString(fmt.Sprintf("[%c,%v,%v]", char, char, pos))
	}
	return buffer.String()
}

func stringToIntArray(str string) []int {
	return appendStringToIntArray(make([]int, 0, len(str)), str)
}

func appendStringToIntArray(array []int, str string) []int {
	for _, char := range str {
		array = append(array, int(char))
	}
	return array
}

func execCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}

func addCommasToInt(i int64) string {
	s := strconv.FormatInt(i, 10)
	if len(s) <= 3 {
		return s
	}
	var b bytes.Buffer
	mod := len(s) % 3
	for _, char := range s {
		if mod == 0 {
			mod = 2
			if b.Len() > 0 {
				b.WriteRune(',')
			}
		} else {
			mod--
		}
		b.WriteRune(char)
	}
	return b.String()
}

/**************************************************************************/
