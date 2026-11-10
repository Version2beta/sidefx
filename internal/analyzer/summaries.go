package analyzer

import (
	"sort"

	"github.com/Version2beta/sidefx/internal/effects"
)

// functionSummary merges source-equivalent variants only at the report boundary;
// graph nodes and their established/candidate relationships remain distinct.
type functionSummary struct {
	direct, reachable, inherited, possible         effects.Set
	uncertain, directUncertain, inheritedUncertain bool
}

func (g *graph) mergeFunctionSummaries() map[identity]functionSummary {
	merged := make(map[identity]functionSummary)
	for _, id := range g.selected {
		merged[id] = functionSummary{}
	}
	for fn, n := range g.nodes {
		source := fn
		if origin := fn.Origin(); origin != nil {
			source = origin
		}
		id, ok := g.selected[source]
		if !ok {
			continue
		}
		s := merged[id]
		s.direct |= n.direct
		s.reachable |= n.reachable
		s.inherited |= n.inherited
		if g.contextualVTA {
			if context := g.contextualRoots[fn]; context != nil {
				s.possible |= context.possible
			}
		} else {
			s.possible |= n.possible
		}
		s.uncertain = s.uncertain || n.uncertain
		s.directUncertain = s.directUncertain || n.directUncertain
		s.inheritedUncertain = s.inheritedUncertain || n.inheritedUncertain
		merged[id] = s
	}
	return merged
}

// projectFunctionSummaries keeps filtered output separate from full inventory
// and coverage. It does not mutate summaries or require a loaded SSA program.
func projectFunctionSummaries(merged map[identity]functionSummary, opts Options, stats Stats) ([]Record, []Record, Stats) {
	stats.SourceFunctions = len(merged)
	records := make([]Record, 0, len(merged))
	inventory := make([]Record, 0, len(merged))
	for id, s := range merged {
		direct := make(map[string]bool)
		for _, name := range s.direct.Names() {
			direct[name] = true
		}
		r := s.record(id, direct, opts.Direct, opts.Effect)
		full := s.record(id, direct, false, "")
		inventory = append(inventory, full)
		if s.reachable != 0 {
			stats.Effectful++
		} else if s.uncertain {
			stats.UnresolvedOnly++
		} else {
			stats.Uncharacterized++
		}
		if s.direct != 0 {
			stats.Direct++
		}
		if s.inherited != 0 {
			stats.Inherited++
		}
		if s.uncertain {
			stats.Uncertain++
		}
		hasEffects := len(r.Effects) != 0 || len(r.PossibleEffects) != 0
		if recordMatches(r, opts) && (opts.Effect == "" || hasEffects) && (opts.All || hasEffects || r.Uncertain) {
			records = append(records, r)
		}
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].Package != records[j].Package {
			return records[i].Package < records[j].Package
		}
		return records[i].Function < records[j].Function
	})
	stats.Emitted = len(records)
	stats.Omitted = stats.SourceFunctions - stats.Emitted
	return records, inventory, stats
}

func (s functionSummary) record(id identity, directEffects map[string]bool, direct bool, effectFilter string) Record {
	set := s.reachable
	if direct {
		set = s.direct
	}
	r := Record{Package: id.pkg, Function: id.name, Effects: []Effect{}, PossibleEffects: []string{}, Uncertain: s.uncertain}
	for _, name := range set.Names() {
		if matches(name, effectFilter) {
			r.Effects = append(r.Effects, Effect{Identifier: name, Direct: directEffects[name]})
		}
	}
	if !direct {
		for _, name := range (s.possible &^ s.reachable).Names() {
			if matches(name, effectFilter) {
				r.PossibleEffects = append(r.PossibleEffects, name)
			}
		}
	}
	return r
}
