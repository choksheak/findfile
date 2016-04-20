findfile
========

findfile is a handy, cross-platform portable, standalone command line file searching utility using non-indexed search only. It is similar to a combination of the find and grep commands in Linux. This is needed because both find and grep are not specifically designed for searching files.

findfile is not meant to replace the Linux find and grep commands, or any other comparable commands, because each program excels in their own niche.

findfile was developed out of a need to search for local files quickly from the command line. It seems like existing tools cannot do what I needed. I first started out using the Windows FIND command, then realized that it cleverly skipped over binary files, including XML files with a unicode header, and hence did not return the search results I needed. Therefore I set out to develop a simple program to search for files, starting from about 50 lines of code, and then it became what it is today.

### Getting Started

findfile is designed to be extremely easy to install and use. Follow these steps to get yourself started using it:

1. Download findfile (from below).
2. Unzip it to extract the executable file.
3. Put the executable file in your PATH, or add its location to your PATH environment variable.
4. Open a terminal / command prompt / powershell prompt.
5. Run `ff -h` or `ff --help`.
6. Run `ff -?` or `ff --info`. Use `ff -? | less` if you have `less` on your machine (Linux, or GnuWin32).
7. Run `ff hello world` to search for the strings `hello` and `world` from your current directory.

### Download Latest Release

Latest release version: 0.4.20160420

The releases are in the form of a single, standalone executable file only. This means that you do not need to install any supporting libraries in order to run the program. But note that the flip side of this is that these executables tend to be larger than what you would find in other native executables because they have to bake in all supporting libraries into one file.

Don't worry. This is not a piece of malware or spyware! If you are skeptical, just download the full source code from GitHub (download zip from website or run `git clone https://github.com/choksheak/findfile`), inspect the code thoroughly, and compile it for yourself. findfile does not ever make or require any network connections.

- [ff.exe - Windows AMD64 (native executable in a zip file)] (https://github.com/choksheak/findfile/blob/master/distribution/findfile-0.4.20160420.windows-amd64.zip?raw=true)
- [ff - Ubuntu Linux AMD64 (native executable in a zip file)] (https://github.com/choksheak/findfile/blob/master/distribution/findfile-0.4.20160420-ubuntu-linux-amd64.zip?raw=true)

### Why is it written in the Go language?

I developed the first, simple version of findfile in Python. But then, it soon became clear that performance was an issue. The performance comes down to mainly two factors: (1) disk I/O, and (2) strings manipulation. Maybe I did something wrong in Python but it was not as fast as I hoped. So I wanted to solve the performance problem once and for all, which basically means that it comes down to some natively-compiled language. It is not that Java or C# could not deliver on this performance, but that VM languages require users to install a huge support runtime framework before users can even run the program. Therefore I did not want to impose such kind of constraint on the end-user. I am also happy to say that findfile does not have any dependency on Python, or any other software that you can think of, which might not be already installed on your machine.

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
- Go seems to be currently (April 2016) the top language in this domain (natively-compiled cross-platform portable). I also considered using [Rust] (https://www.rust-lang.org/) but I found its syntax to be less appealing than Go's syntax (even though the Go syntax also feels odd to me in certain areas).

### Feedback

We would love to hear from you! Please email all comments and suggestions for
improvements to [findfile.go@gmail.com] (mailto:findfile.go@gmail.com).

Have fun searching through your files and let me know how it goes!

**The FindFile Team**
- email: [findfile.go@gmail.com] (mailto:findfile.go@gmail.com)
- website: [https://github.com/choksheak/findfile] (https://github.com/choksheak/findfile)
