// A module of its own, nested inside the repository's.
//
// `go get github.com/bryanster/blacklight/sdk/go` must not drag the server's
// dependency graph into a caller's build — DuckDB (which is cgo and a C++ static
// archive), chromedp, the SAML and OIDC stacks. A caller of a JSON API over HTTP
// should compile with a Go toolchain and nothing else installed, so the client's
// requirements are exactly the three the generator's output imports.
//
// The version here is the module's, not the server's: `GET /version` reports
// what a deployment is running, and this SDK talks to any server whose API
// document it was generated from.
module github.com/bryanster/blacklight/sdk/go

// One minor behind the repository's own go directive on purpose: an SDK that
// demanded the newest toolchain would be unusable by a caller who is one release
// behind, for no feature this client uses.
go 1.24.0

require (
	github.com/oapi-codegen/nullable v1.2.0
	github.com/oapi-codegen/runtime v1.7.0
	go.yaml.in/yaml/v3 v3.0.5
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
)
