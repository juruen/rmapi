# rMAPI

rMAPI is a Go app that allows you to access the ReMarkable Cloud API programmatically.

You can interact with the different API end-points through a shell. However, you can also
run commands non-interactively. This may come in handy to script certian workflows such as
taking automatic backups or uploading documents programmatically.

*Experimental*: Apart from the interactive shell, there's some work happing to access the
documents through a FUSE system. You can read further at the bottom of this file.

![Console Capture](docs/console.gif)

# Some examples of use

[Tutorial on how to directly print to your reMarkable on Mac with rMAPI](docs/tutorial-print-macosx.md)

# Install

## From sources

Install and build the project:

`go get -u github.com/juruen/rmapi`

## Binary

You can download an already built version for either Linux or OSX from [releases](https://github.com/juruen/rmapi/releases).

# API support

- [x] list files and directories
- [x] move around directories
- [x] download a specific file
- [x] download a directory and all its files and subdiretores recursively
- [x] create a directory
- [x] delete a file or a directory
- [x] move/rename a file or a directory
- [x] upload a specific file
- [ ] live syncs

# Shell ergonomics

- [x] autocomplete
- [ ] globbing
- [x] upload a directory and all its files and subdirectories recursively

# Commands

Start the shell by running `rmapi`

## List current directory

Use `ls` to list the contents of the current directory. Entries are listed with `[d]` if they
are directories, and `[f]` if they are files.

## Change current directory

Use `cd` to change the current directory to any other directory in the hiearchy.

## Upload a file

Use `put path_to_local_file` to upload a file  to the current dirctory.

You can also specify the destination directory:

```
put book.pdf /books
```

## Recursively upload directories and files

Use `mput path_to_dir` to recursively upload all the local files to that directory.

E.g: upload all the files

```
mput /Papers
```

![Console Capture](docs/mput-console.png)

## Download a file

Use `get path_to_file` to download a file from the cloud to your local computer.

## Recursively download directories and files

Use `mget path_to_dir` to recursively download all the files in that directory.

E.g: download all the files

```
mget .
```

## Create a directoy

Use `mkdir path_to_new_dir` to create a new directory

##  Remove a directory or a file

Use `rm directory_or_file` to remove. If it's directory, it needs to be empty in order to be deleted.

You can remove multiple entries at the same time.

##  Move/rename a directory or a file

Use `mv source destination` to move or rename a file or directory.

## Stat a directory or file

Use `stat entry` to dump its metadata as reported by the Cloud API.

# Run command non-interactively

Add the commands you want to execute to the arguments of the binary.

E.g: simple script to download all files from the cloud to your local machine

```bash
$ rmapi mget .
```

rMAPI will set the exit code to `0` if the command succeedes, or `1` if it fails.

# Access documents through a mounted FUSE file system

We have started to add support to access your rM documents through a FUSE file system.

This support is very alpha and only a few operations are supported by now:

- [x] Mount/umount
- [x] Lookup entry
- [x] Rename
- [x] Rmdir
- [x] Unlink 
- [x] Read files
- [ ] Write files
- [ ] Live update

If you want to give it a go, you can run:

```bash
rmapi --fuse-mount mount_point
```
