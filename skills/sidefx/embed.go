// Package sidefxskill exposes the canonical embedded agent skill.
package sidefxskill

import (
	_ "embed"
	documentation "github.com/Version2beta/sidefx/docs"
)

// Content is the exact SKILL.md content installed by sidefx skill install.
//
//go:embed SKILL.md
var Content []byte

// Resources are installed beside SKILL.md from the single authoritative docs
// sources. No development helper or project-conditioned retrospective is bundled.
var Resources = map[string][]byte{
	"references/operation-coverage-audit.md":        documentation.OperationCoverage,
	"references/operation-recognition-decisions.md": documentation.RecognitionDecisions,
	"references/operation-index.md":                 documentation.OperationIndex,
}
