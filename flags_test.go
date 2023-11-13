package msgpack

import "flag"

// some tests are skipped by default due to large memory requirements
// or slow execution time; these tests can be run by passing the -all flag
// to 'go test' (e.g. 'go test -all')
//
// add '-all' to VS code settings (go.testFlags": ["-all"]) to run all tests
var allTests = flag.Bool("all", false, "include this flag to run all tests")
