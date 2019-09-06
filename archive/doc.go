// Package archive is used to parse a .zip file retrieved
// by the API.
//
// Here is the content of an archive retried on the tablet as example:
// 384327f5-133e-49c8-82ff-30aa19f3cfa4.content
// 384327f5-133e-49c8-82ff-30aa19f3cfa4//0-metadata.json
// 384327f5-133e-49c8-82ff-30aa19f3cfa4//0.rm
// 384327f5-133e-49c8-82ff-30aa19f3cfa4.pagedata
// 384327f5-133e-49c8-82ff-30aa19f3cfa4.thumbnails/0.jpg
//
// As the .zip file from remarkable is simply a normal .zip file
// containing specific file formats, this package is a helper to
// read and write zip files with the correct format expected by
// the tablet.
//
// At the core of this archive package, we have the Zip struct
// that is defined and that represents a Remarkable zip file.
// Then it provides a Zip.Read() method to unmarshal data
// from an io.Reader into a Zip struct and a Zip.Write() method
// to marshal a Zip struct into a io.Writer.
//
// In order to correctly use this package, you will have to understand
// the format of a Remarkable zip file, and the format of the files
// that it contains.
//
// You can find some help about the format at the following URL:
// https://remarkablewiki.com/tech/filesystem
//
// You can also display the go documentation of public structs of this package
// to have more information. This will be completed in the future hopefully
// to have a precise overall documentation directly held in this Golang package.
//
// Note that the binary format ".rm" holding the drawing contained in a zip has
// a dedicated golang package and is not decoded/encoded from the archive package.
// See encoding/rm in this repository.
//
// To have a more concrete example, see the test files of this package.
package archive
