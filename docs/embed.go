// Package references provides the authoritative local reference bytes for
// installed skills. The markdown remains maintained at its existing docs path.
package references

import _ "embed"

//go:embed operation-coverage-audit.md
var OperationCoverage []byte

//go:embed operation-recognition-decisions.md
var RecognitionDecisions []byte

//go:embed operation-index.md
var OperationIndex []byte
