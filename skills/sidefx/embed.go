// Package sidefxskill exposes the canonical embedded agent skill.
package sidefxskill

import _ "embed"

// Content is the exact SKILL.md content installed by sidefx skill install.
//
//go:embed SKILL.md
var Content []byte
