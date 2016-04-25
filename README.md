FindFile for Windows, Linux, and Mac
====================================

FindFile is a handy, cross-platform portable, standalone command line file searching utility using non-indexed search only. It works the same way across Windows, Linux, and Mac, and any other platform or operating system supported by the Go language. FindFile is similar to a combination of the find and grep commands in Linux. This is needed because both find and grep are not specifically designed for searching through files.

FindFile is not meant to replace the Linux find and grep commands, or any other comparable commands, because each program excels in their own niche.

FindFile was developed out of a need to search for local files quickly from the command line. It seems like existing tools cannot do what I needed to do. I first started out using the Windows FIND command, then realized that it cleverly skipped over binary files, including XML files with a unicode header, and hence did not return the search results I needed. This could be a problem if you rely on its results for some business. Therefore I set out to develop a simple program to search for files, starting from about 50 lines of code, and then it became what it is today.

### Getting Started

FindFile is designed to be extremely easy to install and use. Follow these steps to get yourself started using it:

1. Download FindFile (from below).
2. Unzip it to extract the executable file.
3. Put the executable file in your PATH, or add its location to your PATH environment variable.
4. Open a terminal / command prompt / powershell prompt.
5. Run `ff -h` or `ff --help`.
6. Run `ff -?` or `ff --info` (or [view INFO.md online] (https://github.com/choksheak/findfile/blob/master/INFO.md)). Use `ff -? | less` if you have `less` on your machine (Linux, Mac, or [Less for Windows] (http://gnuwin32.sourceforge.net/packages/less.htm)).
7. Run `ff hello world` to search for the strings `hello` and `world` from your current directory.

### Download Latest Release

The releases are in the form of a single, standalone executable file only. This means that you do not need to install any supporting libraries in order to run the program. But note that the flip side of this is that these executables tend to be larger than what you would find in other native executables because they have to bake in all the supporting libraries into one file.

Don't worry, this is not a piece of malware or spyware! If you are skeptical, just download the full source code from GitHub (download the [zip file] (https://github.com/choksheak/findfile/archive/master.zip) from the website or run `git clone https://github.com/choksheak/findfile`), inspect the source code thoroughly, and compile it for yourself. FindFile does not ever make or require any network connections, because there is no reason for FindFile to need it.

##### Download Links

Latest release version: 0.4.20160420

- [Windows X64/AMD64 -- findfile-0.4.20160420.windows-amd64.zip] (https://github.com/choksheak/findfile/blob/master/distribution/findfile-0.4.20160420.windows-amd64.zip?raw=true)
- [Ubuntu Linux X64/AMD64 -- findfile-0.4.20160420-ubuntu-linux-amd64.zip] (https://github.com/choksheak/findfile/blob/master/distribution/findfile-0.4.20160420-ubuntu-linux-amd64.zip?raw=true)

If you need the software for a different OS/Architecture, please feel free to build it yourself. You will need to install [Go] (https://golang.org/dl/), but will not need to install Git, to do the local build:

- [FindFile source tree -- master.zip] (https://github.com/choksheak/findfile/archive/master.zip)

### How to use FindFile

In the simplest sense, using FindFile is as simple as running the `ff` command with the strings you want to find given as arguments. For example, let's say you are currently in a command prompt at `C:\Temp`. You want to search for the string `hello` that appears in any file under `C:\Temp`. So all you need is to run:

```
ff hello
```

That's the simplest and most common way to use FindFile. There are many command-line options that you can use to control how the search operates. Please see the [INFO.md] (https://github.com/choksheak/findfile/blob/master/INFO.md) file for all the details:

- [Detailed help text in INFO.md] (https://github.com/choksheak/findfile/blob/master/INFO.md)

### Why is it written in the Go language?

I developed the first, simple version of FindFile in Python. But then, it soon became clear that performance was an issue. The performance comes down to mainly two factors: (1) disk I/O, and (2) strings manipulation. Maybe I did something wrong in Python but it was not as fast as I hoped. I know Python but it is not my primary language. So I wanted to solve the performance problem once and for all, which basically means that I need to choose a natively-compiled language. It is not that Java or C# could not deliver on this performance, but that VM languages require users to install a huge support runtime framework before users can even run the program. Therefore I did not want to impose such kind of constraint on the end-user. I am also happy to say that FindFile does not have any dependency on Python, or any other software that you can think of, which might not be already installed on your machine.

Go was chosen as the language because:
- I did not know Go and wanted to learn it just for fun.
- Code execution performance is excellent.
- Writing code in Go is much shorter and simpler than writing code in C or C++ (but much harder than in Java or C#).
- Excellent support for Unicode (not that I have any real use for it now, because I work in English only, but someone might need it).
- Cross-platform portable. Don't have to worry about supporting each platform, including those that I have never heard of.
- Reasonably easy to maintain. The Go build comes out of the box and works just fine.
- Easy download and installation.
- Good supporting online documentation available.
- Large supporting community.
- Availability of supporting tools and IDEs. Go Fmt and Go Lint are simply indispensable. I develop it in Visual Studio Code and it offers full support for the Go language, which makes development much easier than using a plain text editor (I started out coding it in Notepad++).
- Go seems to be currently (April 2016) the top language in this domain (natively-compiled, cross-platform portable). I also considered using [Rust] (https://www.rust-lang.org/) but I found its syntax to be less appealing than Go's syntax (even though the Go syntax also feels odd to me in certain areas).

### Feedback

We would love to hear from you! Please email all comments and suggestions for
improvements to [findfile.go@gmail.com] (mailto:findfile.go@gmail.com).

Have fun searching through your files and let me know how it goes!

**The FindFile Team**
- email: [findfile.go@gmail.com] (mailto:findfile.go@gmail.com)
- website: [https://github.com/choksheak/findfile] (https://github.com/choksheak/findfile)
