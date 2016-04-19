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

// Package findfile contains the implementation of the findfile program.
package findfile

import (
	"bytes"
	"regexp"
	"strings"
)

/**************************************************************************/

// Constants.

const oneLinerUsage = programName + " [option]... <search-string>..."

/**************************************************************************/

// Print short help.

func printDefaultMessage() {
	putln("USAGE: %v", oneLinerUsage)
	putln("(Use option \"%v\" for short help, or \"%v\" for long help)",
		optionHelp.getDefinition().flags,
		optionInfo.getDefinition().flags)
}

func printHelp() {
	commonlyUsedOptions := []option{
		optionHelp,
		optionDir,
		optionSearchNamesOnly,
		optionSearchContentsOnly,
		optionInclude,
		optionExclude,
		optionIgnoreCase,
		optionWholeWord,
		optionRegex,
		optionShowTabs,
		optionContextLines,
		optionContextColumns,
		optionInvertMatch,
		optionSetConfig,
		optionUnsetConfig,
		optionListConfig,
	}

	putBlankLine()
	putln("USAGE: %v", oneLinerUsage)
	putBlankLine()
	putln("Commonly used options: ('-' and '--' can be replaced by '/')")
	putBlankLine()

	commonlyUsedFlags := make([]string, len(commonlyUsedOptions))
	for pos, option := range commonlyUsedOptions {
		commonlyUsedFlags[pos] = option.getDefinition().flags
	}

	printNeatColumns(commonlyUsedFlags, 3, 2)

	putBlankLine()
	putln("Use \"%v %v\" to see all options and detailed help text.", programName, optionInfo.getDefinition().flags)
	putBlankLine()
}

func printNeatColumns(flagsArray []string, initialIndent, numSpacesBetween int) {
	// Break up into array of arrays
	arrayOfArrays := make([][]string, len(flagsArray))
	maxArraySize := 0
	for i, flags := range flagsArray {
		flags := strings.Replace(flags, "=", "|=", -1)
		arrayOfArrays[i] = strings.Split(flags, "|")
		size := len(arrayOfArrays[i])
		if size > maxArraySize {
			maxArraySize = size
		}
	}

	// Find max size of each column
	columnSizes := make([]int, maxArraySize)
	for _, array := range arrayOfArrays {
		for col, flag := range array {
			if len(flag) > columnSizes[col] {
				columnSizes[col] = len(flag)
			}
		}
	}

	// Print each column.
	for _, array := range arrayOfArrays {
		for i := initialIndent; i > 0; i-- {
			putc(' ')
		}
		for col, flag := range array {
			puts(flag)
			if col != len(array)-1 {
				for i := columnSizes[col] - len(flag) + numSpacesBetween; i > 0; i-- {
					putc(' ')
				}
			}
		}
		putBlankLine()
	}
}

func lineWrap(s, lineSeparator string, maxColumn int) string {
	var buffer bytes.Buffer
	column := 1
	words := strings.Split(s, " ")
	for _, word := range words {
		if column+len(word) > maxColumn {
			buffer.WriteString(lineSeparator)
			column = len(word)
		} else if column > 1 {
			buffer.WriteRune(' ')
			column += 1 + len(word)
		} else {
			column += len(word)
		}
		buffer.WriteString(word)
	}
	s = buffer.String()
	return s
}

/**************************************************************************/

// Print long help.

func printInfo() {
	maxColumn := 80
	var helpBuffer bytes.Buffer

	// Top section.
	helpBuffer.WriteString(`
Synopsis:

  ` + oneLinerUsage + `
  
  ` + longProgramName + ` is a cross-platform portable, standalone command line utility for searching through files using non-indexed search.

Option rules:

  1. '-' and '--' can be replaced with '/' in any option (Windows mode).

  2. For options that do not require a value, each time it appears will toggle its value (true/false).
  
  3. For options that require a value (uppercase-letter options), the value must be specified using either '=' or ':', without spaces, e.g. "-X=123"
`)

	// Options.
	for _, optionCategory := range optionCategoriesList {
		helpBuffer.WriteString(`
` + optionCategory.name + ":\n")

		for _, opt := range optionsList {
			option := asOption(opt)
			def := option.getDefinition()

			if def.category != optionCategory {
				continue
			}

			helpBuffer.WriteString("  " + def.flags + ` : ` + def.description + "\n")
		}

		if optionCategory.additionalInfo != "" {
			lines := strings.Split(optionCategory.additionalInfo, "\n")
			for _, line := range lines {
				helpBuffer.WriteString("\n  " + line + "\n")
			}
		}
	}

	helpBuffer.WriteString(`
Output format string:
  %i : result number, 1-indexed
  %p : file path
  %l : line number, 1-indexed
  %c : column number, 1-indexed
  %s : full line
  %% : percent sign
  %n : newline

Specifying option values:
  For options that require a value, you may specify the value using either '=' or ':', without spaces, as follows:
     -x=value
     --xyz:value
     /x:value
     /xyz=value

  For option values with spaces, use double-quotes to enclose the option value:
     -x="hello world"
     "--xyz:hello world"

Environment variables:
  ` + configEnvVar + `
     list of options to use, can be overridden from command line
  ` + editorEnvVar + `
     used as default editor when ` + optionEditor.getDefinition().flags + ` is not specified

Config file:
  Stores default command line options in a config file also:
     Windows - %HOMEDRIVE%%HOMEPATH%\` + configSubDir + `\` + configFileName + `
     Linux   - $HOME/` + configSubDir + `/` + configFileName + `

Examples:
  1. Search for all case-insensitive "World" within files only:
     ` + programName + ` ` + getFirstOptionFlag(optionIgnoreCase) + ` ` + getFirstOptionFlag(optionSearchContentsOnly) + ` world

  2. Search for all filenames containing ".txt":
     ` + programName + ` ` + getFirstOptionFlag(optionSearchNamesOnly) + ` ` + getFirstOptionFlag(optionExclude) + `=* .txt

  3. Search for all lines containing both "-abc" and "-xyz":
     ` + programName + ` ` + getFirstOptionFlag(optionSearchContentsOnly) + ` ` + getFirstOptionFlag(optionEndOfOptions) + ` -abc -xyz

  4. Search for exact phrase "hello world" and open result in notepad:
     ` + programName + ` ` + getFirstOptionFlag(optionSpawn) + ` ` + getFirstOptionFlag(optionEditor) + `=notepad "hello world"

  5. Some possibly useful flags to put in your config file:
     ` + programName + ` ` + getFirstOptionFlag(optionSetConfig) + `="` + getFirstOptionFlag(optionShowTabs) + ` ` + getFirstOptionFlag(optionIgnoreCase) + ` ` + getFirstOptionFlag(optionWriteToFile) + ` ` + getFirstOptionFlag(optionEditor) + `=notepad++ ` + getFirstOptionFlag(optionSpawn) + `"

Feedback:

We would love to hear from you! Please email all comments and suggestions for improvements to ` + contactEmail + `!

Have fun searching through your files!

- The FindFile Team
` + contactEmail + `
` + websiteURL + `

(Help for ` + longProgramName + ` version ` + version + `)
`)

	helpText := helpBuffer.String()

	// Make sure we work with the correct newline.
	helpText = strings.Replace(helpText, "\r\n", "\n", -1)

	// Transform option lines to make them more readable.
	regex := regexp.MustCompile(`(?m)^( +)(-\S+) : (\S.+)$`)
	indent := "\n     "
	indentSize := len(indent) - 1
	thisMaxColumn := maxColumn - indentSize
	helpText = regex.ReplaceAllStringFunc(helpText, func(s string) string {
		parts := regex.FindStringSubmatch(s)
		spaces, flags, text := parts[1], parts[2], parts[3]
		flags = strings.Replace(flags, "|", " ", -1)

		// Special treatment for very short flags (i.e. "--").
		if len(flags) < indentSize {
			return "\n" + spaces + flags + indent[len(spaces)+len(flags)+1:] +
				lineWrap(text, indent, thisMaxColumn)
		}

		return "\n" + spaces + flags + indent + text
	})

	// Add blank line after header text ending with colon if needed.
	regex = regexp.MustCompile(`(?m)^(\S.+:)(\n *\S)`)
	helpText = regex.ReplaceAllStringFunc(helpText, func(s string) string {
		parts := regex.FindStringSubmatch(s)
		header, text := parts[1], parts[2]
		return header + "\n" + text
	})

	// Use uppercase for header text.
	regex = regexp.MustCompile(`(?m)^(\S.+):(.*)$`)
	helpText = regex.ReplaceAllStringFunc(helpText, func(s string) string {
		parts := regex.FindStringSubmatch(s)
		// Add coloring.
		beginHeader := selectString(isTerminal, "\x01", "[")
		endHeader := selectString(isTerminal, "\x02", "]")
		return beginHeader + strings.ToUpper(parts[1]) + endHeader + parts[2]
	})

	// Wrap any long lines.
	regex = regexp.MustCompile(`(?m)^( *)(\S.+)$`)
	helpText = regex.ReplaceAllStringFunc(helpText, func(s string) string {
		if len(s) <= maxColumn {
			return s
		}
		parts := regex.FindStringSubmatch(s)
		spaces, text := parts[1], parts[2]
		thisMaxColumn := maxColumn - len(spaces)
		return spaces + lineWrap(text, "\n"+spaces, thisMaxColumn)
	})

	// Convert to windows line-endings if needed.
	if osNewLine != "\n" {
		helpText = strings.Replace(helpText, "\n", osNewLine, -1)
	}

	// Replace colors with proper codes.
	helpIntArray := stringToIntArray(helpText)
	for pos, char := range helpIntArray {
		if char == 1 {
			helpIntArray[pos] = colorRuneBegin
		} else if char == 2 {
			helpIntArray[pos] = colorRuneEnd
		}
	}

	putIntArrayWithColors(helpIntArray)
	putBlankLine()
}

/**************************************************************************/

// Print version.

func printVersion() {
	putln("%v version %v", longProgramName, version)
}

/**************************************************************************/
