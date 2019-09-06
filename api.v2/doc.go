// Package api is the second version of the client implemented
// for interacting with the Remarkable Cloud API. It has been initiated because
// the first version was not decoupled enough and could not be easily used
// by external packages. The first version of the package is still available for
// backward compatility purposes.
//
// The aim of this package is to provide simple bindings to the Remarkable Cloud API.
// The design has been mostly discussed here: https://github.com/juruen/rmapi/issues/54.
// It has to be high level in order to let a user easily upload, download or interact
// with the storage of a Remarkable device.
//
// The SplitBrain reference has helped a lot to explore the Cloud API has there is no
// official API from Remarkable. See: https://github.com/splitbrain/ReMarkableAPI/wiki.
//
// For interacting with the API, we decoupled the process of authentication
// from the actual storage operations. The authentication is not handled in this package.
// See the auth package from this project.
//
// We tend to follow the good practices from this article to write a modular http client:
// https://medium.com/@marcus.olsson/writing-a-go-client-for-your-restful-api-c193a2f4998c.
package api
