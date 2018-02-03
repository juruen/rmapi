# rmapi

ReMarkable Cloud Go API

# What is this?

This is a shell to interact with ReMarkable Cloud API.

# Install it

# Commands

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
