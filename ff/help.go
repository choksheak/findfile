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
	commonlyUsedOptions := []interface{}{
		optionHelp,
		optionDir,
		optionSearchNamesOnly,
		optionSearchContentsOnly,
		optionIncludeFiles,
		optionExcludeFiles,
		optionIgnoreCase,
		optionWholeWord,
		optionRegex,
		optionExcludeStrings,
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

	commonlyUsedFlags := optionsToFlagsArray(commonlyUsedOptions)
	printNeatColumns(commonlyUsedFlags, 3, 2)
	putBlankLine()

	putln("Use \"%v %v\" to see the list of all options.", programName, optionListOptions.getDefinition().flags)
	putln("Use \"%v %v\" to see all options and detailed help text.", programName, optionInfo.getDefinition().flags)
	putBlankLine()
}

func printListOfOptions() {
	putBlankLine()
	putln("USAGE: %v", oneLinerUsage)
	putBlankLine()

	putln("List of all options: ('-' and '--' can be replaced by '/')")
	putBlankLine()

	allFlags := optionsToFlagsArray(optionsList)
	printNeatColumns(allFlags, 3, 2)
	putBlankLine()

	putln("Use \"%v %v\" to see a list of commonly-used options only.", programName, optionHelp.getDefinition().flags)
	putln("Use \"%v %v\" to see all options and detailed help text.", programName, optionInfo.getDefinition().flags)
	putBlankLine()
}

func optionsToFlagsArray(options []interface{}) [][]string {
	flagsArray := make([][]string, len(options))
	for pos, opt := range options {
		option := asOption(opt)
		flags := option.getDefinition().flags
		flags = strings.Replace(flags, "=", "|=", -1)
		flagsArray[pos] = strings.Split(flags, "|")
	}
	return flagsArray
}

/**************************************************************************/

// Print long help.

func printInfo() {
	var helpBuffer bytes.Buffer

	// Top section.
	helpBuffer.WriteString(`
The MIT License (MIT)
Copyright (c) 2016 Lau, Chok Sheak (for software "findfile")
(Online: https://github.com/choksheak/findfile/blob/master/LICENSE.txt)

Synopsis:

 ` + oneLinerUsage + `
 ` + programName + ` ` + getFirstOptionFlag(optionHelp) + `
 ` + programName + ` some text
 ` + programName + ` ` + getFirstOptionFlag(optionShowTabs) + ` ` + getFirstOptionFlag(optionWholeWord) + ` findme

 ` + longProgramName + ` is a cross-platform portable, standalone command line utility for searching through files using non-indexed search.

Option rules:

 1. Alternate option specifiers
  '-' and '--' can be replaced by '/' in any option (Windows mode).

 2. Toggling boolean options
  For options that do not require a value, each time it appears will toggle its value (true/false).

 3. Specifying option values
  For options that require a value (uppercase-letter options), the value must be specified using either '=' or ':', without spaces, e.g. "-X=123", "/X:123"

 4. Spaces in option values
  For option values with spaces, use double-quotes to enclose the option value, e.g. -X="hello world", "--xyz:hello world"
`)

	// Options.
	for _, optionCategory := range optionCategoriesList {
		helpBuffer.WriteString(`
` + optionCategory.name + `:
`)

		for _, opt := range optionsList {
			option := asOption(opt)
			def := option.getDefinition()

			if def.category != optionCategory {
				continue
			}

			flags := strings.Split(def.flags, "|")
			helpBuffer.WriteString("\n ")
			for pos, flag := range flags {
				if pos > 0 {
					helpBuffer.WriteRune(' ')
				}
				helpBuffer.WriteString(string(rune(-color2RuneBegin)) + flag +
					string(rune(-colorRuneEnd)))
			}
			helpBuffer.WriteString("\n  " + def.description + "\n")
		}

		if optionCategory.additionalInfo != "" {
			lines := strings.Split(optionCategory.additionalInfo, "\n")
			for _, line := range lines {
				helpBuffer.WriteString("\n " + line + "\n")
			}
		}
	}

	helpBuffer.WriteString(`
Output format string:

 %i :  result number, 1-indexed
 %p :  file path
 %l :  line number, 1-indexed
 %c :  column number, 1-indexed
 %s :  full line
 %% :  percent sign
 %n :  newline

Environment variables:

 ` + configEnvVar + `
  list of options to use, can be overridden from command line

 ` + editorEnvVar + `
  used as default editor when ` + optionEditor.getDefinition().flags + ` is not specified

Config file:

 Stores default command line options in a config file also:

 WINDOWS
  %HOMEDRIVE%%HOMEPATH%\` + configSubDir + `\` + configFileName + `

 NON-WINDOWS
  $HOME/` + configSubDir + `/` + configFileName + `

Examples:

 1. Search for all case-insensitive "World" within files only:
  ` + programName + ` ` + getFirstOptionFlag(optionIgnoreCase) + ` ` + getFirstOptionFlag(optionSearchContentsOnly) + ` world

 2. Search for all filenames containing ".txt":
  ` + programName + ` ` + getFirstOptionFlag(optionSearchNamesOnly) + ` ` + getFirstOptionFlag(optionExcludeFiles) + `=* .txt

 3. Search for all lines containing both "-abc" and "-xyz":
  ` + programName + ` ` + getFirstOptionFlag(optionSearchContentsOnly) + ` ` + getFirstOptionFlag(optionEndOfOptions) + ` -abc -xyz

 4. Search for exact phrase "hello world" and open result in notepad:
  ` + programName + ` ` + getFirstOptionFlag(optionSpawn) + ` ` + getFirstOptionFlag(optionEditor) + `=notepad "hello world"

 5. Some possibly useful flags to put in your config file:
  ` + programName + ` ` + getFirstOptionFlag(optionSetConfig) + `="` + getFirstOptionFlag(optionShowTabs) + ` ` + getFirstOptionFlag(optionIgnoreCase) + ` ` + getFirstOptionFlag(optionWriteToFile) + ` ` + getFirstOptionFlag(optionEditor) + `=notepad++ ` + getFirstOptionFlag(optionSpawn) + ` ` + getFirstOptionFlag(optionExcludeDirs) + `=.git;.svn"

Feedback:

 We would love to hear from you! Please email all comments and suggestions for improvements to ` + contactEmail + `!

 Have fun searching through your files!

- The FindFile Team

email:   ` + contactEmail + `
website: ` + websiteURL + `

(Help for ` + longProgramName + ` version ` + version + `)
`)

	helpText := helpBuffer.String()

	// Configure as needed.
	maxColumn := 80
	indentSize := 4

	// Make sure we work with the correct newline.
	helpText = strings.Replace(helpText, "\r\n", "\n", -1)

	// Transform indents.
	indentSpacesArray := make([]string, 3)
	var buffer bytes.Buffer
	for i := 1; i < len(indentSpacesArray); i++ {
		for j := 0; j < indentSize; j++ {
			buffer.WriteRune(' ')
		}
		indentSpacesArray[i] = buffer.String()
	}

	regex := regexp.MustCompile(`(?m)^( +)(\S.+)$`)
	helpText = regex.ReplaceAllStringFunc(helpText, func(s string) string {
		parts := regex.FindStringSubmatch(s)
		spaces, text := parts[1], parts[2]
		indentSpaces := indentSpacesArray[len(spaces)]
		return indentSpaces + text
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

	// Transform section header text.
	regex = regexp.MustCompile(`(?m)^([A-Z]\S.+):(.*)$`)
	helpText = regex.ReplaceAllStringFunc(helpText, func(s string) string {
		parts := regex.FindStringSubmatch(s)
		// Add coloring.
		beginHeader := selectString(isTerminal, string(rune(-colorRuneBegin)), "[")
		endHeader := selectString(isTerminal, string(rune(-colorRuneEnd)), "]")
		return beginHeader + strings.ToUpper(parts[1]) + endHeader + parts[2]
	})

	if isTerminal {
		// Color all numberings.
		regex = regexp.MustCompile(`(?m)^(\s+)(\d+\.)( .+)$`)
		helpText = regex.ReplaceAllStringFunc(helpText, func(s string) string {
			parts := regex.FindStringSubmatch(s)
			spaces, numbering, rest := parts[1], parts[2], parts[3]
			return spaces + string(rune(-color2RuneBegin)) + numbering + string(rune(-colorRuneEnd)) + rest
		})

		// Color all output formats.
		regex = regexp.MustCompile(`(?m)^(\s+)(%\S)(\s+:.+)$`)
		helpText = regex.ReplaceAllStringFunc(helpText, func(s string) string {
			parts := regex.FindStringSubmatch(s)
			spaces, format, rest := parts[1], parts[2], parts[3]
			return spaces + string(rune(-color2RuneBegin)) + format + string(rune(-colorRuneEnd)) + rest
		})

		// Color all sample commands.
		regex = regexp.MustCompile(`(?m)^(\s+)(` + programName + ` .+)$`)
		helpText = regex.ReplaceAllStringFunc(helpText, func(s string) string {
			parts := regex.FindStringSubmatch(s)
			spaces, rest := parts[1], parts[2]
			return spaces + string(rune(-color2RuneBegin)) + rest + string(rune(-colorRuneEnd))
		})

		// Color all small headers.
		regex = regexp.MustCompile(`(?m)^(\S.+:\s+)(\S.*)$`)
		helpText = regex.ReplaceAllStringFunc(helpText, func(s string) string {
			parts := regex.FindStringSubmatch(s)
			header, rest := parts[1], parts[2]
			return header + string(rune(-color2RuneBegin)) + rest + string(rune(-colorRuneEnd))
		})

		// Color all uppercase words.
		regex = regexp.MustCompile(`(?m)^(\s+)([A-Z].+[A-Z])$`)
		helpText = regex.ReplaceAllStringFunc(helpText, func(s string) string {
			parts := regex.FindStringSubmatch(s)
			space, word := parts[1], parts[2]
			return space + string(rune(-color2RuneBegin)) + word + string(rune(-colorRuneEnd))
		})
	}

	// Convert to windows line-endings if needed.
	if osNewLine != "\n" {
		helpText = strings.Replace(helpText, "\n", osNewLine, -1)
	}

	// Replace colors with proper codes.
	helpIntArray := stringToIntArray(helpText)
	for pos, char := range helpIntArray {
		if char <= -colorRuneMinValue {
			helpIntArray[pos] = -char
		}
	}

	putIntArrayWithColors(helpIntArray)
	putBlankLine()
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

// Print version.

func printVersion() {
	putln("%v version %v", longProgramName, version)
}

/**************************************************************************/
