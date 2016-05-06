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

/**************************************************************************/

// findfile constants.

const (
	version               = "0.7.20160506"
	programName           = "ff"
	longProgramName       = "FindFile"
	contactEmail          = "findfile.go@gmail.com"
	websiteURL            = "https://github.com/choksheak/findfile"
	defaultOutputFileName = "ff-output.txt"
	configSubDir          = ".findfile"
	configFileName        = "config.txt"
	configEnvVar          = "FINDFILE_OPTIONS"
	editorEnvVar          = "EDITOR"
	outputFormat0         = "%s%n"
	outputFormat1         = "%p:%l: %s%n"
	outputFormatDefault   = "%n%i. %p line %l col %c%n%s%n"
)

/**************************************************************************/
