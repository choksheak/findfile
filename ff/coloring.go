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
	"github.com/fatih/color"
)

/**************************************************************************/

// Add colors.

const colorRuneBegin = -1
const colorRuneEnd = -2

var hiColor = color.New(color.FgHiWhite).Add(color.BgHiBlue)
var colorNestLevel = 0

func pushColoring() {
	colorNestLevel++
	if colorNestLevel == 1 {
		flush()
		hiColor.Set()
	}
}

func popColoring() {
	if colorNestLevel == 0 {
		panic("No more coloring to pop")
	}
	colorNestLevel--
	if colorNestLevel == 0 {
		flush()
		color.Unset()
	}
}

func resetColoring() {
	colorNestLevel = 0
	flush()
	color.Unset()
}

func putIntArrayWithColors(array []int) {
	for _, char := range array {
		// Check for color escape codes.
		if char >= 0 {
			putc(rune(char))
		} else if char == colorRuneBegin {
			pushColoring()
		} else if char == colorRuneEnd {
			popColoring()
		}
	}
}

/**************************************************************************/
