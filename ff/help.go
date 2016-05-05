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

const (
	oneLinerUsage   = programName + " [option]... <search-string>..."
	color1String    = string(rune(-color1RuneBegin))
	color2String    = string(rune(-color2RuneBegin))
	colorEndString  = string(rune(-colorRuneEnd))
	singleLineBreak = "  "
)

/**************************************************************************/

// Print short help.

func printDefaultMessage() {
	putln("USAGE: %v", oneLinerUsage)
	putln("(Use option \"%v\" for short help, or \"%v\" for long help)", optionHelp.flags, optionInfo.flags)
}

func printHelp() {
	commonlyUsedOptions := []anyOption{
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

	putln("Use \"%v %v\" to see the list of all options.", programName, optionListOptions.flags)
	putln("Use \"%v %v\" to see all options and detailed help text.", programName, optionInfo.flags)
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

	putln("Use \"%v %v\" to see a list of commonly-used options only.", programName, optionHelp.flags)
	putln("Use \"%v %v\" to see all options and detailed help text.", programName, optionInfo.flags)
	putBlankLine()
}

func optionsToFlagsArray(options []anyOption) [][]string {
	flagsArray := make([][]string, len(options))
	for pos, option := range options {
		base := option.getBaseOption()
		flags := base.flags
		flags = strings.Replace(flags, "=", "|=", -1)
		flagsArray[pos] = strings.Split(flags, "|")
	}
	return flagsArray
}

/**************************************************************************/

// Print long help.

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

func getInfoText(markdown bool) string {
	var buffer bytes.Buffer

	// Init markdown tags.
	dlOpen, dlClose, dtOpen, dtClose, ddOpen, ddClose := "", "", "", "", "", ""
	ddIndent, mdLineBreak := " ", ""

	if markdown {
		dlOpen, dlClose = "\n<dl>", "</dl>\n"
		dtOpen, dtClose = "<dt>", "</dt>"
		ddOpen, ddClose = "<dd>", "</dd>"
		ddIndent, mdLineBreak = "", singleLineBreak
	}

	// Top section.
	buffer.WriteString(`
The MIT License (MIT)` + mdLineBreak + `
Copyright (c) 2016 Lau, Chok Sheak (for software "findfile")` + mdLineBreak + `
(Online: https://github.com/choksheak/findfile/blob/master/LICENSE.txt)

Synopsis:

 ` + oneLinerUsage + mdLineBreak + `
 ` + programName + ` ` + getFirstOptionFlag(optionHelp) + mdLineBreak + `
 ` + programName + ` some text` + mdLineBreak + `
 ` + programName + ` ` + getFirstOptionFlag(optionShowTabs) + ` ` + getFirstOptionFlag(optionWholeWord) + ` findme` + mdLineBreak + `

` + ddIndent + longProgramName + ` is a cross-platform portable, standalone command line utility for searching through files using non-indexed search.

Option rules:

` + ddIndent + `1. Alternate option specifiers
` + ddIndent + ` '-' and '--' can be replaced by '/' in any option (Windows mode).

` + ddIndent + `2. Toggling boolean options
` + ddIndent + ` For options that do not require a value, each time it appears will toggle its value (true/false).

` + ddIndent + `3. Specifying option values
` + ddIndent + ` For options that require a value (uppercase-letter options), the value must be specified using either '=' or ':', without spaces, e.g. "-X=123", "/X:123"

` + ddIndent + `4. Spaces in option values
` + ddIndent + ` For option values with spaces, use double-quotes to enclose the option value, e.g. -X="hello world", "--xyz:hello world"
`)

	// Options.
	for _, optionCategory := range optionCategoriesList {
		buffer.WriteString(`
` + optionCategory.name + `:
`)

		buffer.WriteString(dlOpen)

		for _, option := range optionsList {
			base := option.getBaseOption()

			if base.category != optionCategory {
				continue
			}

			flags := strings.Split(base.flags, "|")
			buffer.WriteString("\n ")
			buffer.WriteString(dtOpen)

			for pos, flag := range flags {
				if pos > 0 {
					buffer.WriteRune(' ')
				}
				if !markdown && isTerminal {
					buffer.WriteString(color2String + flag + colorEndString)
				} else {
					buffer.WriteString(flag)
				}
			}

			buffer.WriteString(dtClose)
			buffer.WriteString("\n ")
			buffer.WriteString(ddIndent)
			buffer.WriteString(ddOpen)
			buffer.WriteString(base.description)
			buffer.WriteString(ddClose)
			buffer.WriteRune('\n')
		}

		buffer.WriteString(dlClose)

		if optionCategory.additionalInfo != "" {
			lines := strings.Split(optionCategory.additionalInfo, "\n")
			for _, line := range lines {
				buffer.WriteRune('\n')
				buffer.WriteString(ddIndent)
				buffer.WriteString(line)
				buffer.WriteRune('\n')
			}
		}
	}

	buffer.WriteString(`
Output format string:

` + ddIndent + `%i :  result number, 1-indexed` + mdLineBreak + `
` + ddIndent + `%p :  file path` + mdLineBreak + `
` + ddIndent + `%l :  line number, 1-indexed` + mdLineBreak + `
` + ddIndent + `%c :  column number, 1-indexed` + mdLineBreak + `
` + ddIndent + `%s :  full line` + mdLineBreak + `
` + ddIndent + `%% :  percent sign` + mdLineBreak + `
` + ddIndent + `%n :  newline` + mdLineBreak + `

Environment variables:
` + dlOpen + `
 ` + dtOpen + configEnvVar + dtClose + `
 ` + ddIndent + ddOpen + `list of options to use, can be overridden from command line` + ddClose + `

 ` + dtOpen + editorEnvVar + dtClose + `
 ` + ddIndent + ddOpen + `used as default editor when ` + optionEditor.flags + ` is not specified` + ddClose + `
` + dlClose + `
Config file:

` + ddIndent + longProgramName + ` may optionally store a config file in your home directory which could contain a list of options to set before it reads from the command line. When the same option value is specified in both the config file and the environment variable, the option value from the environment variable will take higher priority. The option values from the command line will always take the highest priority. Note that boolean options will have their values toggled each time they appear, whether from the config file, the environment variable, or the command line. The config file is stored in the following location:
` + dlOpen + `
 ` + dtOpen + `WINDOWS` + dtClose + ` 
 ` + ddIndent + ddOpen + `%HOMEDRIVE%%HOMEPATH%\` + configSubDir + `\` + configFileName + ddClose + `

 ` + dtOpen + `NON-WINDOWS` + dtClose + ` 
 ` + ddIndent + ddOpen + `$HOME/` + configSubDir + `/` + configFileName + ddClose + `
` + dlClose + `
Examples:

` + ddIndent + `1. Search for all case-insensitive "World" within files only:
` + ddIndent + ` ` + programName + ` ` + getFirstOptionFlag(optionIgnoreCase) + ` ` + getFirstOptionFlag(optionSearchContentsOnly) + ` world

` + ddIndent + `2. Search for all filenames containing ".txt":
` + ddIndent + ` ` + programName + ` ` + getFirstOptionFlag(optionSearchNamesOnly) + ` ` + getFirstOptionFlag(optionExcludeFiles) + `=* .txt

` + ddIndent + `3. Search for all lines containing both "-abc" and "-xyz":
` + ddIndent + ` ` + programName + ` ` + getFirstOptionFlag(optionSearchContentsOnly) + ` ` + getFirstOptionFlag(optionEndOfOptions) + ` -abc -xyz

` + ddIndent + `4. Search for exact phrase "hello world" and open result in notepad:
` + ddIndent + ` ` + programName + ` ` + getFirstOptionFlag(optionSpawn) + ` ` + getFirstOptionFlag(optionEditor) + `=notepad "hello world"

` + ddIndent + `5. Some possibly useful flags to put in your config file:
` + ddIndent + ` ` + programName + ` ` + getFirstOptionFlag(optionSetConfig) + `="` + getFirstOptionFlag(optionShowTabs) + ` ` + getFirstOptionFlag(optionIgnoreCase) + ` ` + getFirstOptionFlag(optionWriteToFile) + ` ` + getFirstOptionFlag(optionEditor) + `=notepad++ ` + getFirstOptionFlag(optionSpawn) + ` ` + getFirstOptionFlag(optionExcludeDirs) + `=.git;.svn"

Feedback:

` + ddIndent + `We would love to hear from you! Please email all comments and suggestions for improvements to ` + contactEmail + `!

` + ddIndent + `Have fun searching through your files!

- The FindFile Team

email:   ` + contactEmail + mdLineBreak + `
website: ` + websiteURL + `

(Help for ` + longProgramName + ` version ` + version + `)
`)

	text := buffer.String()

	// Configure as needed.
	maxColumn := 80
	indentSize := 4

	// Make sure we work with the correct newline.
	text = strings.Replace(text, "\r\n", "\n", -1)

	// Transform indents.
	indentSpacesArray := make([]string, 3)
	buffer.Reset()
	for i := 1; i < len(indentSpacesArray); i++ {
		for j := 0; j < indentSize; j++ {
			buffer.WriteRune(' ')
		}
		indentSpacesArray[i] = buffer.String()
	}

	regex := regexp.MustCompile(`(?m)^( +)(\S.+)`)
	text = regex.ReplaceAllStringFunc(text, func(s string) string {
		parts := regex.FindStringSubmatch(s)
		spaces, text := parts[1], parts[2]
		indentSpaces := indentSpacesArray[len(spaces)]
		return indentSpaces + text
	})

	// Wrap any long lines.
	regex = regexp.MustCompile(`(?m)^( *)(\S.+)`)
	text = regex.ReplaceAllStringFunc(text, func(s string) string {
		if len(s) <= maxColumn {
			return s
		}
		parts := regex.FindStringSubmatch(s)
		spaces, text := parts[1], parts[2]
		thisMaxColumn := maxColumn - len(spaces)
		return spaces + lineWrap(text, "\n"+spaces, thisMaxColumn)
	})

	return text
}

func transformSectionHeader(text string, transform func(sectionHeader string) string) string {
	regex := regexp.MustCompile(`(?m)^([A-Z][^,:@!]+)(:)\n`)
	return regex.ReplaceAllStringFunc(text, func(s string) string {
		parts := regex.FindStringSubmatch(s)
		sectionHeader := parts[1]
		return transform(sectionHeader) + "\n"
	})
}

func transformNumberings(text string, transform func(numbering, gap, text string) string) string {
	regex := regexp.MustCompile(`(?m)^( *)(\d+\.)( )([^\n]+)`)
	return regex.ReplaceAllStringFunc(text, func(s string) string {
		parts := regex.FindStringSubmatch(s)
		indentSpaces, numbering, gap, rest := parts[1], parts[2], parts[3], parts[4]
		return indentSpaces + transform(numbering, gap, rest)
	})
}

func transformOutputFormats(text string, transform func(indentSpaces, format, gap, description string) string) string {
	regex := regexp.MustCompile(`(?m)^( *)(%\S)( +: +)(\S.+)`)
	return regex.ReplaceAllStringFunc(text, func(s string) string {
		parts := regex.FindStringSubmatch(s)
		indentSpaces, format, gap, description := parts[1], parts[2], parts[3], parts[4]
		return transform(indentSpaces, format, gap, description)
	})
}

func transformSampleCommands(text string, transform func(indentSpaces, command, spaces string) string) string {
	regex := regexp.MustCompile(`(?m)^( +)(` + programName + ` .+\S)( *)`)
	return regex.ReplaceAllStringFunc(text, func(s string) string {
		parts := regex.FindStringSubmatch(s)
		indentSpaces, command, spaces := parts[1], parts[2], parts[3]
		return transform(indentSpaces, command, spaces)
	})
}

func transformSmallHeaders(text string, transform func(header, gap, text string) string) string {
	regex := regexp.MustCompile(`(?m)^(\S.+:)( +)(\S.*)\n`)
	return regex.ReplaceAllStringFunc(text, func(s string) string {
		parts := regex.FindStringSubmatch(s)
		header, gap, text := parts[1], parts[2], parts[3]
		return transform(header, gap, text) + "\n"
	})
}

func transformUppercaseWords(text string, transform func(indentSpaces, word string) string) string {
	regex := regexp.MustCompile(`(?m)^(\s+)([A-Z].+[A-Z])$`)
	return regex.ReplaceAllStringFunc(text, func(s string) string {
		parts := regex.FindStringSubmatch(s)
		indentSpaces, word := parts[1], parts[2]
		return transform(indentSpaces, word)
	})
}

func printInfo() {
	text := getInfoText(false)

	// Transform section header text.
	beginSectionHeader := selectString(isTerminal, color1String, "[")
	endSectionHeader := selectString(isTerminal, colorEndString, "]")

	text = transformSectionHeader(text, func(sectionHeader string) string {
		return beginSectionHeader + strings.ToUpper(sectionHeader) + endSectionHeader
	})

	if isTerminal {
		// Color all numberings.
		text = transformNumberings(text, func(numbering, gap, text string) string {
			return color2String + numbering + colorEndString + gap + text
		})

		// Color all output formats.
		text = transformOutputFormats(text, func(indentSpaces, format, gap, description string) string {
			return indentSpaces + color2String + format + colorEndString + gap + description
		})

		// Color all sample commands.
		text = transformSampleCommands(text, func(indentSpaces, command, spaces string) string {
			return indentSpaces + color2String + command + colorEndString
		})

		// Color all small headers.
		text = transformSmallHeaders(text, func(header, gap, text string) string {
			return header + gap + color2String + text + colorEndString
		})

		// Color all uppercase words.
		text = transformUppercaseWords(text, func(indentSpaces, word string) string {
			return indentSpaces + color2String + word + colorEndString
		})
	}

	// Convert to windows line-endings if needed.
	if osNewLine != "\n" {
		text = strings.Replace(text, "\n", osNewLine, -1)
	}

	// Replace colors with proper codes.
	helpIntArray := stringToIntArray(text)
	for pos, char := range helpIntArray {
		if char <= -colorRuneMinValue {
			helpIntArray[pos] = -char
		}
	}

	putIntArrayWithColors(helpIntArray)
	putBlankLine()
}

func printMarkDownInfo() {
	text := getInfoText(true)

	// Transform section header text.
	text = transformSectionHeader(text, func(sectionHeader string) string {
		return "### " + strings.ToUpper(sectionHeader)
	})

	// Transform all numberings.
	text = transformNumberings(text, func(numbering, gap, text string) string {
		return numbering + gap + "**" + text + "**" + singleLineBreak
	})

	// Transform all output formats.
	text = transformOutputFormats(text, func(indentSpaces, format, gap, description string) string {
		return indentSpaces + "`" + format + "`" + gap + description
	})

	// Transform all sample commands.
	text = transformSampleCommands(text, func(indentSpaces, command, spaces string) string {
		if spaces == "" {
			return indentSpaces + "`" + command + "`" + spaces
		}
		return indentSpaces + command
	})

	// Transform all small headers.
	text = transformSmallHeaders(text, func(header, gap, text string) string {
		return header + gap + text
	})

	// Transform all uppercase words.
	text = transformUppercaseWords(text, func(indentSpaces, word string) string {
		return "</dd>\n" +
			"</dl>\n" +
			"<dl>\n" +
			indentSpaces + "<dt>" + word + "</dt>\n" +
			indentSpaces + "<dd>"
	})

	// Add page title.
	title := "Info for " + longProgramName + " for Windows, Linux, and Mac"
	var buffer bytes.Buffer
	for i := len(title); i > 0; i-- {
		buffer.WriteRune('=')
	}
	text = title + "\n" + buffer.String() + "\n" + text

	puts(text)
	putBlankLine()
}

/**************************************************************************/

// Print version.

func printVersion() {
	putln("%v version %v", longProgramName, version)
}

/**************************************************************************/
