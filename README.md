# rmapi

ReMarkable Cloud Go API

# What is this?

An attempt to access the ReMarkable Cloud API programmatically.
So far, we expose interactions through a shell. However, you can
run the shell commands non-interactively as way to create scripts
that work with your reMarkable data.

![Console Capture](docs/console.gif)

# Install it

Install and build the project:

`go get -u github.com/juruen/rmapi`

# API support

- [x] list files and directories
- [x] move around directories
- [x] download a specific file
- [x] download a directory and all its files and subdiretores recursively
- [ ] delete a file or a directory
- [ ] renme a file or a directory
- [ ] upload a specic file
- [ ] upload a directory and all its files and subdirectories recursively

# Commands

Start the shell by running `rmapi`

## List current directory

Use `ls` to list the contents of the current directory. Entries are listed with `[d]` if they
are directories, and `[f]` if they are files.

## Change current directory

Use `cd` to change the current directory to any other directory in the hiearchy.

## Download a file

Use `get path_to_file` to download a file from the cloud to your local computer.

## Recursively download directories and files

Use `mget path_to_dir` to recursively download all the files in that directory.

E.g: download all the files

```
mget /
```

# Run command non-interactively

Add the commands you want to execute to the arguments of the binary, and add
`exit` as the first argument.

E.g: simple script to download all files from the cloud to your local machine

```bash
$ rmapi exit mget .
```
