Info for FindFile for Windows, Linux, and Mac
=============================================

The MIT License (MIT)  
Copyright (c) 2016 Lau, Chok Sheak (for software "findfile")  
(Online: https://github.com/choksheak/findfile/blob/master/LICENSE.txt)

### SYNOPSIS

    ff [option]... <search-string>...
    ff -h
    ff some text
    ff -t -w findme

FindFile is a cross-platform portable, standalone command line utility for
searching through files using non-indexed search.

### OPTION RULES

1. **Alternate option specifiers**  
    '-' and '--' can be replaced by '/' in any option (Windows mode).
2. **Toggling boolean options**  
    For options that do not require a value, each time it appears will
    toggle its value (true/false).
3. **Specifying option values**  
    For options that require a value (uppercase-letter options), the value
    must be specified using either '=' or ':', without spaces, e.g. "-X=123",
    "/X:123"
4. **Spaces in option values**  
    For option values with spaces, use double-quotes to enclose the option
    value, e.g. -X="hello world", "--xyz:hello world"

### GENERAL OPTIONS

<dl>
    <dt>-h --help</dt>
    <dd>print short help with commonly-used options</dd>

    <dt>-o --list-options</dt>
    <dd>print the list of all options</dd>

    <dt>-? --info</dt>
    <dd>print full help information</dd>

    <dt>-vs --version</dt>
    <dd>print version and exit

    <dt>-l --list-all</dt>
    <dd>list all the dir and file names without searching

    <dt>-a --show-args</dt>
    <dd>print passed-in arguments that are specified after this argument</dd>

    <dt>-- --end-of-options</dt>
    <dd>indicate end of options, so that you can write search strings starting
        with dash '-' and slash '/'</dd>
</dl>

### WHERE TO SEARCH

<dl>
    <dt>-D --dir=[starting-dir]</dt>
    <dd>starting dir to search, defaults to the current dir "."

    <dt>-M --max-levels=[-1:2147483647]</dt>
    <dd>search up to given dir depth, 0 to search starting dir only; default of
        -1 means no limit</dd>
</dl>

### WHAT TO SEARCH

<dl>
    <dt>-n --search-names-only</dt>
    <dd>search dir and file names only, ignoring file contents</dd>

    <dt>-c --search-contents-only</dt>
    <dd>search file contents only, ignoring dir and file names</dd>

    <dt>-IF --include-files=[glob-pattern]</dt>
    <dd>include glob pattern for files</dd>

    <dt>-ID --include-dirs=[glob-pattern]</dt>
    <dd>include glob pattern for dirs</dd>

    <dt>-XF --exclude-files=[glob-pattern]</dt>
    <dd>exclude glob pattern for files</dd>

    <dt>-XD --exclude-dirs=[glob-pattern]</dt>
    <dd>exclude glob pattern for dirs</dd>
</dl>

For multiple glob patterns, use ';' as the delimiter, e.g. "*.cfg; *.txt;
*.go". Leading and trailing spaces in each sub-expression will be ignored.

Note that when you use the '*' and '?' pattern strings from the command
line, they may be escaped by the command shell before FindFile is invoked.
Therefore it is best to always enclose these patterns with double-quotes,
e.g. -I="*.txt", or "-I=*.txt".

For detailed syntax of glob patterns, please see:
    https://golang.org/pkg/path/filepath/#Match 

### HOW TO MATCH SEARCH STRING

<dl>
    <dt>-i --ignore-case</dt>
    <dd>use case-insensitive matching</dd>

    <dt>-w --whole-word</dt>
    <dd>match whole words only</dd>

    <dt>-r --regex</dt>
    <dd>treat search strings as regular expressions</dd>

    <dt>-EX --exclude-strings=[strings-to-exclude]</dt>
    <dd>exclude lines containing given strings, delimited by ';'</dd>
</dl>

For detailed syntax of regex patterns, please see:
https://golang.org/pkg/regexp/syntax/

### HOW TO DISPLAY OUTPUT

<dl>
    <dt>-m --measure-stats</dt>
    <dd>measure time taken and number of bytes read</dd>

    <dt>-nc --no-color</dt>
    <dd>turn off coloring for matching strings; coloring only applies for
        terminal windows</dd>

    <dt>-b --show-brackets</dt>
    <dd>show brackets around the matching substrings</dd>

    <dt>-t --show-tabs</dt>
    <dd>make tabs visible</dd>

    <dt>-T --tab-spacing=[0:30]</dt>
    <dd>number of spaces per tab, defaults to 4</dd>

    <dt>-bin --search-binary-files</dt>
    <dd>include binary files in the search; by default they will be skipped</dd>

    <dt>-cc --show-control-chars</dt>
    <dd>show all control characters as-is; control characters here are defined
        as ASCII characters 0-8, 11-12, 14-31, 127</dd>

    <dt>-L --context-lines=[0:2147483647]</dt>
    <dd>print number of lines before and after match; default is 0 to show
        matching line only</dd>

    <dt>-C --context-columns=[0:2147483647]</dt>
    <dd>print number of characters around and including matching substring;
        default is 200; use 0 to show the entire line</dd>

    <dt>-abs --absolute-path</dt>
    <dd>print absolute file paths</dd>

    <dt>-v --invert-match</dt>
    <dd>print non-matching lines or file/dir names only</dd>

    <dt>-q --quiet</dt>
    <dd>turn off supporting messages</dd>

    <dt>-0 --format0 --show-lines-only</dt>
    <dd>print matching lines only, without any other decoration: "%s%n"</dd>

    <dt>-1 --format1 --show-filenames-and-lines</dt>
    <dd>use one-liner compact format per match: "%p:%l: %s%n"</dd>

    <dt>-2 --format2 --show-filenames-and-counts</dt>
    <dd>print matching filenames with match counts</dd>

    <dt>-3 --format3 --show-filenames-only</dt>
    <dd>print matching filenames only</dd>

    <dt>-F --format=[format-string]</dt>
    <dd>specify output format string; default is "%n%i. %p line %l col %c%n%s%n"</dd>
</dl>

### OUTPUT FILE OPTIONS

<dl>
    <dt>-wf --write-to-file</dt>
    <dd>write output to file</dd>

    <dt>-O --output-file=[filename]</dt>
    <dd>write output to given file name, defaults to "\temp\ff-output.txt in
        Windows, and "/tmp/ff-output.txt" in other operating systems</dd>

    <dt>-s --spawn</dt>
    <dd>spawn editor program to open output file</dd>

    <dt>-E --editor=[editor-path]</dt>
    <dd>specify editor program to use to open output file</dd>
</dl>

### CONFIG FILE OPTIONS

<dl>
    <dt>-S --set-config=[options-to-set]</dt>
    <dd>save the given options in a config file and exit</dd>

    <dt>-U --unset-config=[options-to-unset]</dt>
    <dd>unset the given options in a config file and exit</dd>

    <dt>-lc --list-config</dt>
    <dd>print the current config file values and exit</dd>
</dl>

### OUTPUT FORMAT STRING

    `%i` :  result number, 1-indexed
    `%p` :  file path
    `%l` :  line number, 1-indexed
    `%c` :  column number, 1-indexed
    `%s` :  full line
    `%%` :  percent sign
    `%n` :  newline

### ENVIRONMENT VARIABLES

<dl>
    <dt>FINDFILE_OPTIONS</dt>
    <dd>list of options to use, can be overridden from command line</dd>

    <dt>EDITOR</dt>
    <dd>used as default editor when -E|--editor=[editor-path] is not specified</dd>
</dl>

### CONFIG FILE

Stores default command line options in a config file also:

<dl>
    <dt>WINDOWS</dt>
    <dd>%HOMEDRIVE%%HOMEPATH%\.findfile\config.txt</dd>

    <dt>NON-WINDOWS</dt>
    <dd>$HOME/.findfile/config.txt</dd>
</dl>

### EXAMPLES

1. Search for all case-insensitive "World" within files only:  
    `ff -i -c world`

2. Search for all filenames containing ".txt":  
    `ff -n -XF=* .txt`

3. Search for all lines containing both "-abc" and "-xyz":  
    `ff -c -- -abc -xyz`

4. Search for exact phrase "hello world" and open result in notepad:  
    `ff -s -E=notepad "hello world"`

5. Some possibly useful flags to put in your config file:  
    `ff -S="-t -i -wf -E=notepad++ -s -XD=.git;.svn"`

### FEEDBACK

We would love to hear from you! Please email all comments and suggestions
for improvements to findfile.go@gmail.com!

Have fun searching through your files!

- The FindFile Team

email:   findfile.go@gmail.com  
website: https://github.com/choksheak/findfile

(Help for FindFile version 0.4.20160420)

