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
	"io/ioutil"
	"os"
	"strings"
)

/**************************************************************************/

// Config files

func setConfigOptions(options string) {
	// If user's home directory does not exist, return.
	exists, _ := pathExists(userHomeDir)
	if !exists {
		putln("User's home directory %v does not exist. Cannot set config options.", userHomeDir)
		exit(1)
	}

	// Create findfile config dir.
	exists, _ = pathExists(configFileDir)
	if !exists {
		err := os.Mkdir(configFileDir, 0755)
		if err != nil {
			putln("Cannot create config dir \"%v\" to store config options. %v", configFileDir, err)
			exit(1)
		}
	}

	// Parse input string as options list.
	newOptionStrings := splitArgumentsString(options, "command line")

	if len(newOptionStrings) == 0 {
		putln("No options were specified to be set in the config file.")
		exit(1)
	}

	// Validate for disallowed options.
	checkForUnrecognizedOptions(newOptionStrings, "option value of "+optionSetConfig.getDefinition().flags)
	checkForDisallowedConfigOptions(newOptionStrings, "config file")

	// Read options from config file.
	mergedOptionsMap := readConfigFileOptionsMap()

	// Info.
	putln("Setting config options: %v", options)
	putBlankLine()

	// Merge old and new options.
	for _, newOptionString := range newOptionStrings {
		newOption := getOptionByFlag(newOptionString)
		newDef := newOption.getDefinition()
		oldOptionString := mergedOptionsMap[newDef.name]

		if oldOptionString == "" {
			putln("Adding new option: %v (%v)", newOptionString, newDef.flags)
			mergedOptionsMap[newDef.name] = newOptionString
		} else if oldOptionString == newOptionString {
			putln("Option already exists: %v (%v)",
				oldOptionString,
				newDef.flags)
		} else {
			putln("Replacing existing option %v with %v (%v)",
				oldOptionString,
				newOptionString,
				newDef.flags)
			mergedOptionsMap[newDef.name] = newOptionString
		}
	}

	// Write back list of options into config file.
	putBlankLine()
	writeOptionsToConfigFile(mergedOptionsMap, configFilePath)
}

func unsetConfigOptions(options string) {
	// If config file does not exist, return.
	exists, _ := pathExists(configFilePath)
	if !exists {
		putln("Config file \"%v\" does not exist. Nothing to unset.", configFilePath)
		exit(1)
	}

	// Read options from config file.
	mergedOptionsMap := readConfigFileOptionsMap()

	// Validate that config file args are present.
	if len(mergedOptionsMap) == 0 {
		putln("No arguments defined in the config file \"%v\". Nothing to unset.", configFilePath)
		exit(1)
	}

	// Parse input string as options list.
	newOptionStrings := splitArgumentsString(options, "command line")
	checkForUnrecognizedOptions(newOptionStrings, "option value of "+optionUnsetConfig.getDefinition().flags)
	checkForDisallowedConfigOptions(newOptionStrings, "config file")

	// Info.
	putln("Unsetting config options: %v", options)
	putBlankLine()

	// Remove new options from old options.
	for _, newOptionString := range newOptionStrings {
		newOption := getOptionByFlag(newOptionString)
		newDef := newOption.getDefinition()
		oldOptionString := mergedOptionsMap[newDef.name]

		if oldOptionString == "" {
			putln("Ignoring option: %v (%v)", newOptionString, newDef.flags)
		} else {
			putln("Removing existing option: %v (%v)", oldOptionString, newDef.flags)
			delete(mergedOptionsMap, newDef.name)
		}
	}

	// Write back list of options into config file.
	putBlankLine()
	writeOptionsToConfigFile(mergedOptionsMap, configFilePath)
}

func readConfigFileOptionsMap() map[string]string {
	// Read options from config file.
	fileContents := readConfigFile()
	rawOldOptionStrings := splitArgumentsString(fileContents, "config file "+configFilePath)

	// Add good old options, thus removing bad old options.
	mergedOptionsMap := make(map[string]string)

	for _, oldOptionString := range rawOldOptionStrings {
		oldOption := getOptionByFlag(oldOptionString)
		if oldOption == nil {
			putln("Removing unknown option %v in the config file %v.", oldOptionString, configFilePath)
			continue
		}
		mergedOptionsMap[oldOption.getDefinition().name] = oldOptionString
	}

	return mergedOptionsMap
}

func writeOptionsToConfigFile(optionsToWrite map[string]string, configFilePath string) {
	var optionsBuffer bytes.Buffer

	for _, optionString := range optionsToWrite {
		if optionsBuffer.Len() > 0 {
			optionsBuffer.WriteRune(' ')
		}

		hasSpace := strings.ContainsAny(optionString, " \a\b\f\n\r\t\v")
		if hasSpace {
			optionsBuffer.WriteRune('"')
			optionsBuffer.WriteString(escapeString(optionString))
			optionsBuffer.WriteRune('"')
		} else {
			optionsBuffer.WriteString(optionString)
		}
	}

	// Write back list of options into config file.
	putln("Writing options to config file \"%v\".", configFilePath)
	putln("  %v", optionsBuffer.String())
	err := ioutil.WriteFile(configFilePath, optionsBuffer.Bytes(), 0755)
	if err != nil {
		putln("Failed to write options to file: %v", err)
		exit(1)
	}
	putBlankLine()
}

func printConfigOptions() {
	putBlankLine()

	exists, _ := pathExists(configFilePath)
	if !exists {
		putln("Config file %v does not exist.", configFilePath)
	} else {
		// Read options from config file.
		fileContents := readConfigFile()
		rawOptionStrings := splitArgumentsString(fileContents, "config file "+configFilePath)

		if len(rawOptionStrings) == 0 {
			putln("Config file %v is empty.", configFilePath)
		} else {
			putln("Config file location:")
			putln("  %v", configFilePath)
			putBlankLine()

			putln("Options loaded from the config file:")
			putln("  %v", fileContents)
			putBlankLine()

			putln("Legend:")
			arrayOfArrays := make([][]string, len(rawOptionStrings))
			for pos, optionString := range rawOptionStrings {
				flag, _ := splitOptionFlagAndValue(optionString)
				option := getOptionByFlag(flag)
				if option == nil {
					arrayOfArrays[pos] = []string{optionString, ":", "Unknown option"}
				} else {
					arrayOfArrays[pos] = []string{flag, ":", option.getDefinition().flags}
				}
			}

			printNeatColumns(arrayOfArrays, 2, 2)
		}
	}
	putBlankLine()

	putln("Environment variables:")
	value, isSet := os.LookupEnv(configEnvVar)
	putln("  %v = %v", configEnvVar, selectString(isSet, value, "(not set)"))
	value, isSet = os.LookupEnv(editorEnvVar)
	putln("  %v = %v", editorEnvVar, selectString(isSet, value, "(not set)"))
	putBlankLine()
}

/**************************************************************************/