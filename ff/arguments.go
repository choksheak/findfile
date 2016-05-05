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
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

/**************************************************************************/

// Option categories like an enum.

type optionCategory struct {
	name           string
	additionalInfo string
}

// Add the option categories in order.
// Name the categories like how they would appear in the help text.
// Structs cannot take a default initial value, so using structs
// will force us to write each field name twice.
var (
	optionCategoryGeneral = newOptionCategory("General options", "")
	optionCategoryWhere   = newOptionCategory("Where to search", "")
	optionCategoryWhat    = newOptionCategory("What to search",
		`For multiple glob patterns, use ';' as the delimiter, e.g. "*.cfg; *.txt; *.go". Leading and trailing spaces in each sub-expression will be ignored.
Note that when you use the '*' and '?' pattern strings from the command line, they may be escaped by the command shell before `+longProgramName+` is invoked. Therefore it is best to always enclose these patterns with double-quotes, e.g. -I="*.txt", or "-I=*.txt".
For detailed syntax of glob patterns, please see: https://golang.org/pkg/path/filepath/#Match`)
	optionCategoryMatching = newOptionCategory("How to match search string",
		`For detailed syntax of regex patterns, please see: https://golang.org/pkg/regexp/syntax/`)
	optionCategoryOutputDisplay = newOptionCategory("How to display output", "")
	optionCategoryOutputFile    = newOptionCategory("Output file options", "")
	optionCategoryConfigFile    = newOptionCategory("Config file options", "")
)

// List of all option categories.
var optionCategoriesList = make([]optionCategory, 0, 10)

func newOptionCategory(categoryName, additionalInfo string) optionCategory {
	newCategory := optionCategory{categoryName, additionalInfo}
	optionCategoriesList = append(optionCategoriesList, newCategory)
	return newCategory
}

/**************************************************************************/

// Option definitions.

type baseOption struct {
	category    optionCategory
	name        string
	flags       string
	description string
	isGiven     bool
}

type boolOption struct {
	baseOption
	defaultValue bool
	value        bool
}

type intOption struct {
	baseOption
	defaultValue int
	value        int
	minValue     int
	maxValue     int
}

type stringOption struct {
	baseOption
	isGiven      bool
	defaultValue string
	value        string
}

type anyOption interface {
	getBaseOption() *baseOption
}

func (o *boolOption) getBaseOption() *baseOption {
	return &o.baseOption
}

func (o *intOption) getBaseOption() *baseOption {
	return &o.baseOption
}

func (o *stringOption) getBaseOption() *baseOption {
	return &o.baseOption
}

/**************************************************************************/

// List of options.

var (
	// General.
	optionHelp = newBoolOption(optionCategoryGeneral,
		"help", "-h|--help",
		"print short help with commonly-used options", false)
	optionListOptions = newBoolOption(optionCategoryGeneral,
		"list-options", "-o|--list-options",
		"print the list of all options", false)
	optionInfo = newBoolOption(optionCategoryGeneral,
		"info", "-?|--info",
		"print full help information", false)
	optionVersion = newBoolOption(optionCategoryGeneral,
		"version", "-vs|--version",
		"print version and exit", false)
	optionListAll = newBoolOption(optionCategoryGeneral,
		"list", "-l|--list-all",
		"list all the dir and file names without searching", false)
	optionShowArgs = newBoolOption(optionCategoryGeneral,
		"show-args", "-a|--show-args",
		"print passed-in arguments that are specified after this argument", false)
	optionEndOfOptions = newBoolOption(optionCategoryGeneral,
		"end-of-options", "--|--end-of-options",
		"indicate end of options, so that you can write search strings starting with dash '-' and slash '/'", false)
	optionMarkDown = newBoolOption(optionCategoryGeneral,
		"markdown", "-md|--markdown",
		"print help information in markdown format", false)

	// Where.
	optionDir = newStringOption(optionCategoryWhere,
		"dir", "-D|--dir=[starting-dir]",
		"starting dir to search, defaults to the current dir \".\"", ".")
	optionMaxLevels = newIntOption(optionCategoryWhere,
		"max-levels", "-M|--max-levels=[-1:"+strconv.Itoa(math.MaxInt32)+"]",
		"search up to given dir depth, 0 to search starting dir only; default of -1 means no limit", -1)

	// What.
	optionSearchNamesOnly = newBoolOption(optionCategoryWhat,
		"search-names-only", "-n|--search-names-only",
		"search dir and file names only, ignoring file contents", false)
	optionSearchContentsOnly = newBoolOption(optionCategoryWhat,
		"search-contents-only", "-c|--search-contents-only",
		"search file contents only, ignoring dir and file names", false)
	optionIncludeFiles = newStringOption(optionCategoryWhat,
		"include", "-IF|--include-files=[glob-pattern]",
		"include glob pattern for files", "")
	optionIncludeDirs = newStringOption(optionCategoryWhat,
		"include-dirs", "-ID|--include-dirs=[glob-pattern]",
		"include glob pattern for dirs", "")
	optionExcludeFiles = newStringOption(optionCategoryWhat,
		"exclude", "-XF|--exclude-files=[glob-pattern]",
		"exclude glob pattern for files", "")
	optionExcludeDirs = newStringOption(optionCategoryWhat,
		"exclude-dirs", "-XD|--exclude-dirs=[glob-pattern]",
		"exclude glob pattern for dirs", "")

	// Matching.
	optionIgnoreCase = newBoolOption(optionCategoryMatching,
		"ignore-case", "-i|--ignore-case",
		"use case-insensitive matching", false)
	optionWholeWord = newBoolOption(optionCategoryMatching,
		"whole-word", "-w|--whole-word",
		"match whole words only", false)
	optionRegex = newBoolOption(optionCategoryMatching,
		"regex", "-r|--regex",
		"treat search strings as regular expressions", false)
	optionExcludeStrings = newStringOption(optionCategoryMatching,
		"exclude", "-EX|--exclude-strings=[strings-to-exclude]",
		"exclude lines containing given strings, delimited by ';'", "")

	// Output display.
	optionMeasureStats = newBoolOption(optionCategoryOutputDisplay,
		"measure-stats", "-m|--measure-stats",
		"measure time taken and number of bytes read", false)
	optionNoColor = newBoolOption(optionCategoryOutputDisplay,
		"no-color", "-nc|--no-color",
		"turn off coloring for matching strings; coloring only applies for terminal windows", false)
	optionShowBrackets = newBoolOption(optionCategoryOutputDisplay,
		"show-brackets", "-b|--show-brackets",
		"show brackets around the matching substrings", false)
	optionShowTabs = newBoolOption(optionCategoryOutputDisplay,
		"show-tabs", "-t|--show-tabs",
		"make tabs visible", false)
	optionTabSpacing = newIntOption(optionCategoryOutputDisplay,
		"tab-spacing", "-T|--tab-spacing=[0:30]",
		"number of spaces per tab, defaults to 4", 4)
	optionSearchBinaryFiles = newBoolOption(optionCategoryOutputDisplay,
		"search-binary-files", "-bin|--search-binary-files",
		"include binary files in the search; by default they will be skipped", false)
	optionShowControlChars = newBoolOption(optionCategoryOutputDisplay,
		"show-control-chars", "-cc|--show-control-chars",
		"show all control characters as-is; control characters here are defined as ASCII characters 0-8, 11-12, 14-31, 127",
		false)
	optionContextLines = newIntOption(optionCategoryOutputDisplay,
		"context-lines", "-L|--context-lines=[0:"+strconv.Itoa(math.MaxInt32)+"]",
		"print number of lines before and after match; default is 0 to show matching line only", 0)
	optionContextColumns = newIntOption(optionCategoryOutputDisplay,
		"context-columns", "-C|--context-columns=[0:"+strconv.Itoa(math.MaxInt32)+"]",
		"print number of characters around and including matching substring; default is 200; use 0 to show the entire line", 200)
	optionAbsolutePath = newBoolOption(optionCategoryOutputDisplay,
		"absolute-path", "-abs|--absolute-path",
		"print absolute file paths", false)
	optionInvertMatch = newBoolOption(optionCategoryOutputDisplay,
		"invert-match", "-v|--invert-match",
		"print non-matching lines or file/dir names only", false)
	optionQuiet = newBoolOption(optionCategoryOutputDisplay,
		"quiet", "-q|--quiet",
		"turn off supporting messages", false)
	optionFirstResult = newIntOption(optionCategoryOutputDisplay,
		"first-result", "-FR|--first-result=[1:"+strconv.Itoa(math.MaxInt32)+"]",
		"start returning results from the Nth result; default is 1 to start from the first result", 1)
	optionMaxResults = newIntOption(optionCategoryOutputDisplay,
		"max-results", "-MR|--max-results=[0:"+strconv.Itoa(math.MaxInt32)+"]",
		"return up to the given maximum number of results; default is 0 to show all results", 0)
	optionFormat0ShowLinesOnly = newBoolOption(optionCategoryOutputDisplay,
		"format0", "-0|--format0|--show-lines-only",
		"print matching lines only, without any other decoration: \""+outputFormat0+"\"", false)
	optionFormat1ShowFileNamesAndLines = newBoolOption(optionCategoryOutputDisplay,
		"format1", "-1|--format1|--show-filenames-and-lines",
		"use one-liner compact format per match: \""+outputFormat1+"\"", false)
	optionFormat2ShowFileNamesAndCounts = newBoolOption(optionCategoryOutputDisplay,
		"format2", "-2|--format2|--show-filenames-and-counts",
		"print matching filenames with match counts", false)
	optionFormat3ShowFileNamesOnly = newBoolOption(optionCategoryOutputDisplay,
		"format3", "-3|--format3|--show-filenames-only",
		"print matching filenames only", false)
	optionFormat = newStringOption(optionCategoryOutputDisplay,
		"format", "-F|--format=[format-string]",
		"specify output format string; default is \""+outputFormatDefault+"\"", outputFormatDefault)

	// Output file.
	optionWriteToFile = newBoolOption(optionCategoryOutputFile,
		"write-to-file", "-wf|--write-to-file",
		"write output to file", false)
	optionOutputFile = newStringOption(optionCategoryOutputFile,
		"output-file", "-O|--output-file=[filename]",
		"write output to given file name, defaults to \"\\temp\\"+defaultOutputFileName+" in Windows, and \"/tmp/"+defaultOutputFileName+"\" in other operating systems", "")
	optionSpawn = newBoolOption(optionCategoryOutputFile,
		"spawn", "-s|--spawn",
		"spawn editor program to open output file", false)
	optionEditor = newStringOption(optionCategoryOutputFile,
		"editor", "-E|--editor=[editor-path]",
		"specify editor program to use to open output file", "")

	// Config file.
	optionSetConfig = newStringOption(optionCategoryConfigFile,
		"set-config", "-S|--set-config=[options-to-set]",
		"save the given options in a config file and exit", "")
	optionUnsetConfig = newStringOption(optionCategoryConfigFile,
		"unset-config", "-U|--unset-config=[options-to-unset]",
		"unset the given options in a config file and exit", "")
	optionListConfig = newBoolOption(optionCategoryConfigFile,
		"list-config", "-lc|--list-config",
		"print the current config file values and exit", false)
)

// These options cannot be specified within config data.
var disallowedConfigOptions = map[string]bool{
	optionHelp.name:        true,
	optionListOptions.name: true,
	optionInfo.name:        true,
	optionMarkDown.name:    true,
	optionVersion.name:     true,
	optionSetConfig.name:   true,
	optionUnsetConfig.name: true,
	optionListConfig.name:  true,
}

// List of all options.
var optionsList = make([]anyOption, 0, 50)

/**************************************************************************/

// Construct options.

func newBoolOption(category optionCategory, name, flags, description string, defaultValue bool) *boolOption {
	newOption := boolOption{
		baseOption: baseOption{
			category:    category,
			name:        name,
			flags:       flags,
			description: description,
			isGiven:     false,
		},
		defaultValue: defaultValue,
		value:        defaultValue,
	}
	optionsList = append(optionsList, &newOption)
	return &newOption
}

func newIntOption(category optionCategory, name, flags, description string, defaultValue int) *intOption {
	// Get range.
	_, value := splitOptionFlagAndValue(flags)
	minMax := strings.Split(value[1:len(value)-1], ":")
	min, err1 := strconv.Atoi(minMax[0])
	max, err2 := strconv.Atoi(minMax[1])

	if err1 != nil || err2 != nil {
		panic("Value range not integer: " + flags)
	}

	if min > max {
		panic("Bad min max range: " + flags)
	}

	newOption := intOption{
		baseOption: baseOption{
			category:    category,
			name:        name,
			flags:       flags,
			description: description,
			isGiven:     false,
		},
		defaultValue: defaultValue,
		value:        defaultValue,
		minValue:     min,
		maxValue:     max,
	}

	optionsList = append(optionsList, &newOption)
	return &newOption
}

func newStringOption(category optionCategory, name, flags, description, defaultValue string) *stringOption {
	if !strings.Contains(flags, "=") {
		panic("Missing value name in string option " + flags)
	}

	newOption := stringOption{
		baseOption: baseOption{
			category:    category,
			name:        name,
			flags:       flags,
			description: description,
			isGiven:     false,
		},
		defaultValue: defaultValue,
		value:        defaultValue,
	}

	optionsList = append(optionsList, &newOption)
	return &newOption
}

/**************************************************************************/

// Option utilities.

var flagToOptionMap = initFlagToOptionMap()

func initFlagToOptionMap() map[string]anyOption {
	m := make(map[string]anyOption)
	for _, option := range optionsList {
		// Process flags string to drop sample value and support slashes.
		base := option.getBaseOption()
		flagsArray := addAndGetWithSlashOptionFlags(base.flags)
		for _, flag := range flagsArray {
			if m[flag] != nil {
				panic("Option flag doubly-specific: " + flag)
			}
			m[flag] = option
		}
	}
	return m
}

func addAndGetWithSlashOptionFlags(flags string) []string {
	flags, _ = splitOptionFlagAndValue(flags)
	flagsArray := strings.Split(flags, "|")
	length := len(flagsArray)
	for i := 0; i < length; i++ {
		flag := flagsArray[i]
		if strings.HasPrefix(flag, "--") {
			newFlag := "/" + flag[2:]
			flagsArray = append(flagsArray, newFlag)
		} else {
			newFlag := "/" + flag[1:]
			flagsArray = append(flagsArray, newFlag)
		}
	}
	return flagsArray
}

func splitOptionFlagAndValue(option string) (string, string) {
	equalIndex := strings.Index(option, "=")
	colonIndex := strings.Index(option, ":")

	if (equalIndex >= 0) && (colonIndex >= 0) {
		lesser := equalIndex
		if colonIndex < lesser {
			lesser = colonIndex
		}
		return option[0:lesser], option[lesser+1:]
	} else if equalIndex > 0 {
		return option[0:equalIndex], option[equalIndex+1:]
	} else if colonIndex > 0 {
		return option[0:colonIndex], option[colonIndex+1:]
	}

	return option, ""
}

func tryGetOptionByFlag(flag string) anyOption {
	flag, _ = splitOptionFlagAndValue(flag)

	// Will return nil when not found.
	return flagToOptionMap[flag]
}

func getRuneByEscapeChar(char rune) rune {
	switch char {
	case 'a':
		return '\a'
	case 'b':
		return '\b'
	case 'f':
		return '\f'
	case 'n':
		return '\n'
	case 'r':
		return '\r'
	case 't':
		return '\t'
	case 'v':
		return '\v'
	default:
		return char
	}
}

func getEscapeStringByRune(char rune) string {
	switch char {
	case '\a':
		return "\\a"
	case '\b':
		return "\\b"
	case '\f':
		return "\\f"
	case '\n':
		return "\\n"
	case '\r':
		return "\\r"
	case '\t':
		return "\\t"
	case '\v':
		return "\\v"
	default:
		return string(char)
	}
}

func escapeString(s string) string {
	var buffer bytes.Buffer
	for _, char := range s {
		buffer.WriteString(getEscapeStringByRune(char))
	}
	return buffer.String()
}

func getFirstOptionFlag(option anyOption) string {
	return strings.Split(option.getBaseOption().flags, "|")[0]
}

func printNeatColumns(arrayOfArrays [][]string, initialIndent, numSpacesBetween int) {
	// Get max size of all arrays.
	maxArraySize := 0
	for _, array := range arrayOfArrays {
		size := len(array)
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

/**************************************************************************/

// Parse arguments.

var configFileDir = filepath.Join(userHomeDir, configSubDir)
var configFilePath = filepath.Join(userHomeDir, configSubDir, configFileName)

var endOfOptionsReached = false
var optionArguments = make([]string, 0, 10)
var nonOptionArguments = make([]string, 0, 10)

func hasOptionPrefix(optionString string) bool {
	return (optionString != "") && ((optionString[0] == '/') || (optionString[0] == '-'))
}

func appendNonOptionArgument(argumentString, sourceName string) {
	// Ignore empty non-option arguments because they don't mean anything.
	if argumentString == "" {
		return
	}

	if (optionSetConfig.value != "") || (optionUnsetConfig.value != "") {
		putln("Cannot specify search string when setting or unsetting config.")
		exit(1)
	}

	nonOptionArguments = append(nonOptionArguments, argumentString)
}

func parseAndSetArgument(argumentString, sourceName string, allowNonOptions bool) {

	if endOfOptionsReached {
		appendNonOptionArgument(argumentString, sourceName)
		return
	}

	if !hasOptionPrefix(argumentString) {
		if !allowNonOptions {
			putln("Argument \"%v\" read from the %v is not an option.", argumentString, sourceName)
			exit(1)
		}
		endOfOptionsReached = true
		appendNonOptionArgument(argumentString, sourceName)
		return
	}

	flag, value := splitOptionFlagAndValue(argumentString)

	// Get and check option.
	option := tryGetOptionByFlag(flag)
	if option == nil {
		putln("Unrecognized option \"%v\" given in the %v.", flag, sourceName)
		exit(1)
	}

	// Check for end of options indicator.
	if option == optionEndOfOptions {
		endOfOptionsReached = true
		return
	}

	// Set value if any.
	flags := option.getBaseOption().flags

	switch typedOption := option.(type) {
	case *boolOption:
		if value != "" {
			putln("Option %v read from the %v cannot take a value", flags, sourceName)
			exit(1)
		}
		typedOption.isGiven = true

		// Specifying boolean options each time will toggle its value.
		typedOption.value = !typedOption.value
		value = strconv.FormatBool(typedOption.value)

	case *intOption:
		if value == "" {
			putln("Option %v read from the %v requires a value.", flags, sourceName)
			exit(1)
		}
		number, err := strconv.Atoi(value)
		if (err != nil) || (number < typedOption.minValue) || (number > typedOption.maxValue) {
			putln("Option %v read from the %v must be an integer between %v and %v inclusive.",
				argumentString, sourceName, typedOption.minValue, typedOption.maxValue)
			exit(1)
		}
		typedOption.isGiven = true
		typedOption.value = number
		value = strconv.Itoa(number)

	case *stringOption:
		if value == "" {
			putln("Option %v read from the %v requires a value.", flags, sourceName)
			exit(1)
		}
		typedOption.isGiven = true
		typedOption.value = value

	default:
		panic(fmt.Sprintf("Bad option type: %#v", typedOption))
	}

	optionArguments = append(optionArguments, argumentString)
}

func splitArgumentsString(arguments, sourceName string) []string {
	var buffer bytes.Buffer
	insideQuote := false
	insideString := false
	isEscapeCode := false
	array := make([]string, 0, 10)

	for _, char := range arguments {
		if isEscapeCode {
			buffer.WriteRune(getRuneByEscapeChar(char))
			isEscapeCode = false
		} else if unicode.IsSpace(char) {
			if insideString {
				insideString = false
				array = append(array, buffer.String())
				buffer.Reset()
			} else if insideQuote {
				buffer.WriteRune(char)
			}
		} else if char == '"' {
			if insideQuote {
				insideQuote = false
				insideString = true
			} else {
				insideString = false
				insideQuote = true
			}
		} else if char == '\\' {
			if insideQuote {
				isEscapeCode = true
			} else {
				buffer.WriteRune(char)
			}
		} else {
			if !insideQuote {
				insideString = true
			}
			buffer.WriteRune(char)
		}
	}

	if buffer.Len() > 0 {
		if isEscapeCode {
			putln("Unterminated arguments string read from the %v ending with '\\': %v", sourceName, arguments)
			exit(1)
		}
		if insideQuote {
			putln("Unterminated arguments string read from the %v ending with '\"': %v", sourceName, arguments)
			exit(1)
		}
		array = append(array, buffer.String())
	}

	return array
}

/**************************************************************************/

// Load and read arguments.

func loadArguments() {
	// Load all arguments first because we might need to print them all out.
	configFileContents := strings.TrimSpace(readConfigFile())
	configFileSourceName := "config file " + configFilePath
	configFileArguments := splitArgumentsString(configFileContents, configFileSourceName)

	envVarContents := strings.TrimSpace(os.Getenv(configEnvVar))
	envVarSourceName := "environment variable " + configEnvVar
	envVarArguments := splitArgumentsString(envVarContents, envVarSourceName)

	commandLineArguments := os.Args[1:]

	// Check if we need to print out the arguments.
	setBooleanOptionValue(optionShowArgs, configFileArguments, false)
	setBooleanOptionValue(optionShowArgs, envVarArguments, false)
	setBooleanOptionValue(optionShowArgs, commandLineArguments, true)

	setBooleanOptionValue(optionListConfig, commandLineArguments, true)

	// Print args if needed.
	if optionShowArgs.value || optionListConfig.value {
		printArguments(
			configFileArguments,
			envVarArguments,
			commandLineArguments,
			optionListConfig.value,
			optionListConfig.value,
			optionShowArgs.value)

		// Reset option values.
		optionShowArgs.value = false
		optionListConfig.value = false
	}

	// Do the actual load now.
	loadArgumentsFromConfigString(configFileArguments, configFileSourceName)
	loadArgumentsFromConfigString(envVarArguments, envVarSourceName)
	loadArgumentsFromCommandLine()
}

func setBooleanOptionValue(optionToSet *boolOption, arguments []string, allowNonOptions bool) bool {
	for _, arg := range arguments {
		if allowNonOptions {
			if !hasOptionPrefix(arg) {
				break
			}
		}
		option := tryGetOptionByFlag(arg)
		if option == nil {
			// Ignore invalid options for now.
			continue
		}
		if option == optionToSet {
			// Use boolean toggle.
			optionToSet.value = !optionToSet.value
		}
		if allowNonOptions {
			if option == optionEndOfOptions {
				break
			}
		}
	}
	return false
}

func printArguments(
	configFileArguments []string,
	envVarArguments []string,
	commandLineArguments []string,
	showEmptyArguments bool,
	showEditorEnvVar bool,
	showCommandLineArguments bool) {

	isPrinted := false

	if showEmptyArguments || len(configFileArguments) > 0 {
		isPrinted = true
		putBlankLine()
		putln("Config file location:")
		putln("%v%v", printIndent, configFilePath)

		putBlankLine()
		putln("Config file options:")

		exists, _ := pathExists(configFilePath)
		if !exists {
			putln("%vConfig file does not exist", printIndent)
		} else {
			if len(configFileArguments) == 0 {
				putln("%v<empty>", printIndent)
			} else {
				printArgumentsAsList(configFileArguments, false)
			}
		}
	}

	if showEmptyArguments || len(envVarArguments) > 0 {
		isPrinted = true
		putBlankLine()
		putln("Environment variable %v options:", configEnvVar)
		if len(envVarArguments) == 0 {
			putln("%v<not set>", printIndent)
		} else {
			printArgumentsAsList(envVarArguments, false)
		}
	}

	if showEditorEnvVar {
		isPrinted = true
		putBlankLine()
		putln("Environment variable %v value:", editorEnvVar)
		editor := os.Getenv(editorEnvVar)
		putln("%v%v", printIndent, selectString(editor == "", "<not set>", editor))
	}

	// When listing config only, don't show command line arguments.
	if showCommandLineArguments {
		isPrinted = true
		putBlankLine()
		putln("Command line arguments:")
		if len(commandLineArguments) == 0 {
			putln("%v<empty>", printIndent)
		} else {
			printArgumentsAsList(commandLineArguments, true)
		}
	}

	if isPrinted {
		putBlankLine()
	}
}

func printArgumentsAsList(arguments []string, allowNonOptions bool) {
	arrayOfArrays := make([][]string, len(arguments))
	endOfOptions := false

	for pos, argument := range arguments {
		if allowNonOptions {
			if endOfOptions {
				arrayOfArrays[pos] = []string{argument, ":", "Search string"}
				continue
			} else if !hasOptionPrefix(argument) {
				endOfOptions = true
				arrayOfArrays[pos] = []string{argument, ":", "Search string"}
				continue
			}
		}
		flag, _ := splitOptionFlagAndValue(argument)
		option := tryGetOptionByFlag(flag)
		if option == nil {
			arrayOfArrays[pos] = []string{argument, ":", "Unknown option"}
		} else {
			if allowNonOptions {
				if option == optionEndOfOptions {
					endOfOptions = true
				}
			}
			arrayOfArrays[pos] = []string{argument, ":", option.getBaseOption().flags}
		}
	}

	printNeatColumns(arrayOfArrays, len(printIndent), 2)
}

func loadArgumentsFromCommandLine() {
	sourceName := "command line"
	for i := 1; i < len(os.Args); i++ {
		parseAndSetArgument(os.Args[i], sourceName, true)
	}
}

func readConfigFile() string {
	// If config file does not exist, then return.
	exists, _ := pathExists(configFilePath)
	if !exists {
		return ""
	}

	// Try to read the file.
	bytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func loadArgumentsFromConfigString(arguments []string, sourceName string) {
	checkForUnrecognizedOptions(arguments, sourceName)
	checkForDisallowedConfigOptions(arguments, sourceName)

	for _, argumentString := range arguments {
		parseAndSetArgument(argumentString, sourceName, false)
	}
}

func checkForDisallowedConfigOptions(arguments []string, sourceName string) {
	for _, argument := range arguments {
		flag, _ := splitOptionFlagAndValue(argument)
		option := tryGetOptionByFlag(flag)
		base := option.getBaseOption()
		if disallowedConfigOptions[base.name] {
			putln("Please remove option %v which is not allowed in the %v.", base.flags, sourceName)
			exit(1)
		}
	}
}

func checkForUnrecognizedOptions(arguments []string, sourceName string) {
	for _, argument := range arguments {
		flag, _ := splitOptionFlagAndValue(argument)
		option := tryGetOptionByFlag(flag)
		if option == nil {
			putln("Please remove unknown option %v given in the %v.", flag, sourceName)
			exit(1)
		}
	}
}

/**************************************************************************/

// Validate arguments.

func validateArguments() {
	// Search where.
	if optionSearchNamesOnly.value && optionSearchContentsOnly.value {
		putln("Only one of %v and %v can be given at a time.",
			optionSearchNamesOnly.flags,
			optionSearchContentsOnly.flags)
		exit(1)
	}

	// Turn off context lines when showing filenames only.
	if optionFormat2ShowFileNamesAndCounts.value || optionFormat3ShowFileNamesOnly.value {
		optionContextLines.value = 0
	}

	// Output format.
	count := 0
	if optionFormat0ShowLinesOnly.isGiven {
		count++
	}
	if optionFormat1ShowFileNamesAndLines.isGiven {
		count++
	}
	if optionFormat2ShowFileNamesAndCounts.isGiven {
		count++
	}
	if optionFormat3ShowFileNamesOnly.isGiven {
		count++
	}
	if optionFormat.isGiven {
		count++
	}
	if count > 1 {
		putln("Only one of %v, %v, %v, %v and %v can be given at a time.",
			optionFormat0ShowLinesOnly.flags,
			optionFormat1ShowFileNamesAndLines.flags,
			optionFormat2ShowFileNamesAndCounts.flags,
			optionFormat3ShowFileNamesOnly.flags,
			optionFormat.flags)
	}

	// Enable output file automatically when spawn is specified.
	if !optionWriteToFile.value && optionSpawn.value {
		optionWriteToFile.value = true
	}

	// When spawn is specified, editor must be available.
	if optionSpawn.value && (optionEditor.value == "") {
		optionEditor.value = os.Getenv(editorEnvVar)
		if optionEditor.value == "" {
			putln("Editor not specified for spawn. Please use the %v option or set the environment variable %v.",
				optionEditor.flags, editorEnvVar)
			exit(1)
		}
	}

	// Cannot set and unset config at the same time.
	if (optionSetConfig.value != "") && (optionUnsetConfig.value != "") {
		putln("Cannot specify %v and %v at the same time.",
			optionSetConfig.flags,
			optionUnsetConfig.flags)
		exit(1)
	}

	// Cannot specify any non-option arguments when specifying disallowed config options.
	if len(nonOptionArguments) > 0 {
		for _, argument := range optionArguments {
			flag, _ := splitOptionFlagAndValue(argument)
			option := tryGetOptionByFlag(flag)
			base := option.getBaseOption()
			if disallowedConfigOptions[base.name] {
				putln("Cannot specify search strings when specifying the %v option.", base.flags)
				exit(1)
			}
		}
	}
}
