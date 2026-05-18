// Package plugin provides the remote plugin service contract and client helpers.
//
// A plugin is an independently deployed service that registers its HTTP routes
// with the Keiyaku-Go host. This package intentionally contains no references
// to internal application packages so generated or external plugin services can
// depend on it safely.
package plugin
