package analyzer

import (
	"reflect"
	"sort"
	"testing"

	"github.com/Version2beta/sidefx/internal/effects"
	"golang.org/x/tools/go/ssa"
)

func TestMergeFunctionSummariesPreservesVariantEvidence(t *testing.T) {
	a, b, empty, dependency := &ssa.Function{}, &ssa.Function{}, &ssa.Function{}, &ssa.Function{}
	id := identity{"example.com/p", "Merged"}
	g := &graph{
		selected: map[*ssa.Function]identity{a: id, b: id, empty: {"example.com/p", "Empty"}},
		nodes: map[*ssa.Function]*node{
			a:          {direct: effects.FSRead, reachable: effects.FSRead, possible: effects.FSWrite, uncertain: true, directUncertain: true},
			b:          {direct: effects.FSWrite, reachable: effects.FSRead | effects.FSWrite, inherited: effects.FSRead, possible: effects.FSRemove, inheritedUncertain: true},
			dependency: {reachable: effects.ProcessExit},
		},
	}
	want := map[identity]functionSummary{
		id:                         {direct: effects.FSRead | effects.FSWrite, reachable: effects.FSRead | effects.FSWrite, inherited: effects.FSRead, possible: effects.FSWrite | effects.FSRemove, uncertain: true, directUncertain: true, inheritedUncertain: true},
		{"example.com/p", "Empty"}: {},
	}
	if got := g.mergeFunctionSummaries(); !reflect.DeepEqual(got, want) {
		t.Fatalf("merged source summaries: got %+v want %+v", got, want)
	}
	// Root-context refinement replaces global possibilities, including when a
	// selected variant has no context. It must not fall back per variant.
	g.contextualVTA = true
	g.contextualRoots = map[*ssa.Function]*candidateContextResult{a: {possible: effects.NetConnect}}
	merged := want[id]
	merged.possible = effects.NetConnect
	want[id] = merged
	if got := g.mergeFunctionSummaries(); !reflect.DeepEqual(got, want) {
		t.Fatalf("contextual source summaries: got %+v want %+v", got, want)
	}
	if len(g.nodes) != 3 || g.nodes[a] == g.nodes[b] || g.nodes[a].possible != effects.FSWrite || g.nodes[b].possible != effects.FSRemove {
		t.Fatal("source aggregation changed graph nodes")
	}
}

func TestProjectFunctionSummariesSeparatesInventoryCoverageAndFilters(t *testing.T) {
	merged := map[identity]functionSummary{
		{"p", "Mixed"}:    {direct: effects.FSRead, reachable: effects.FSRead | effects.FSWrite, inherited: effects.FSWrite, possible: effects.FSWrite | effects.FSRemove, uncertain: true},
		{"p", "Possible"}: {possible: effects.NetConnect},
		{"p", "Unknown"}:  {uncertain: true},
		{"p", "Empty"}:    {},
		{"q", "Other"}:    {direct: effects.EnvRead, reachable: effects.EnvRead},
	}
	baseStats := Stats{Packages: 2, DependencyVTA: &DependencyVTAStats{ContextBudget: 7}}
	fullStats := Stats{Packages: 2, SourceFunctions: 5, Effectful: 2, Direct: 2, Inherited: 1, Uncertain: 2, UnresolvedOnly: 1, Uncharacterized: 2, DependencyVTA: baseStats.DependencyVTA}
	wantInventory := []Record{
		{Package: "p", Function: "Empty", Effects: []Effect{}, PossibleEffects: []string{}},
		{Package: "p", Function: "Mixed", Effects: []Effect{{"fs.read", true}, {"fs.write", false}}, PossibleEffects: []string{"fs.remove"}, Uncertain: true},
		{Package: "p", Function: "Possible", Effects: []Effect{}, PossibleEffects: []string{"net.connect"}},
		{Package: "p", Function: "Unknown", Effects: []Effect{}, PossibleEffects: []string{}, Uncertain: true},
		{Package: "q", Function: "Other", Effects: []Effect{{"env.read", true}}, PossibleEffects: []string{}},
	}
	for _, tc := range []struct {
		name string
		opts Options
		want []Record
	}{
		{"default", Options{}, wantInventory[1:]},
		{"all", Options{All: true}, wantInventory},
		{"exact root", Options{Function: "=Mixed"}, wantInventory[1:2]},
		{"package", Options{Package: "q"}, wantInventory[4:]},
		{"candidate filter", Options{Effect: "remove"}, []Record{{Package: "p", Function: "Mixed", Effects: []Effect{}, PossibleEffects: []string{"fs.remove"}, Uncertain: true}}},
		{"direct", Options{Direct: true, Function: "=Mixed"}, []Record{{Package: "p", Function: "Mixed", Effects: []Effect{{"fs.read", true}}, PossibleEffects: []string{}, Uncertain: true}}},
		{"direct excludes candidates", Options{Direct: true, Effect: "remove", All: true}, []Record{}},
		{"missing effect", Options{Effect: "absent", All: true}, []Record{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			records, inventory, stats := projectFunctionSummaries(merged, tc.opts, baseStats)
			if !reflect.DeepEqual(records, tc.want) {
				t.Fatalf("records got %+v want %+v", records, tc.want)
			}
			// Inventory ordering remains the consuming projection's policy.
			sort.Slice(inventory, func(i, j int) bool {
				if inventory[i].Package != inventory[j].Package {
					return inventory[i].Package < inventory[j].Package
				}
				return inventory[i].Function < inventory[j].Function
			})
			if !reflect.DeepEqual(inventory, wantInventory) {
				t.Fatalf("filter changed full inventory: %+v", inventory)
			}
			wantStats := fullStats
			wantStats.Emitted, wantStats.Omitted = len(records), len(merged)-len(records)
			if stats != wantStats {
				t.Fatalf("coverage got %+v want %+v", stats, wantStats)
			}
			// Returned records must not alias the inventory's effect slices.
			for i := range records {
				if len(records[i].Effects) > 0 {
					records[i].Effects[0].Identifier = "changed"
				}
			}
			if !reflect.DeepEqual(inventory, wantInventory) {
				t.Fatal("filtered record mutation changed inventory")
			}
		})
	}
	if merged[identity{"p", "Mixed"}].possible != effects.FSWrite|effects.FSRemove || baseStats != (Stats{Packages: 2, DependencyVTA: fullStats.DependencyVTA}) {
		t.Fatal("projection mutated input summaries or stats")
	}
}

func TestProjectFunctionSummariesEmptyUsesArrays(t *testing.T) {
	records, inventory, stats := projectFunctionSummaries(nil, Options{}, Stats{})
	if records == nil || inventory == nil || len(records) != 0 || len(inventory) != 0 || stats != (Stats{}) {
		t.Fatalf("empty projection: %#v %#v %+v", records, inventory, stats)
	}
}
