package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Version2beta/sidefx/internal/analyzer"
)

func TestWriteEffectsCategoryCountsIncludeCandidateSites(t *testing.T) {
	for _, tc := range []struct {
		name        string
		established []string
		possible    []string
		want        string
	}{
		{"established", []string{"site:established"}, nil, "fs.read (1 sites)\n"},
		{"candidate only", nil, []string{"site:candidate"}, "fs.read (1 sites; 0 established, 1 candidate-only)\n"},
		{"mixed", []string{"site:established"}, []string{"site:candidate"}, "fs.read (2 sites; 1 established, 1 candidate-only)\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &analyzer.EffectReport{
				Groups: []analyzer.EffectGroup{{Effect: "fs.read", Sites: tc.established, PossibleSites: tc.possible}},
				Sites: []analyzer.Site{
					{ID: "site:established", APIs: []string{"os.ReadFile"}},
					{ID: "site:candidate", APIs: []string{"os.ReadFile", "os.ReadDir"}},
				},
			}
			var out bytes.Buffer
			if err := WriteEffects(&out, r, false); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out.String(), tc.want) {
				t.Fatalf("category count loses candidate evidence; want %q:\n%s", tc.want, out.String())
			}
		})
	}
}

func TestWriteEffectsInventoryPreservesPossibleEffects(t *testing.T) {
	// Inventory is unfiltered even when no operation groups are emitted.
	r := &analyzer.EffectReport{Inventory: []analyzer.Record{
		{Package: "p", Function: "Candidate", Effects: []analyzer.Effect{}, PossibleEffects: []string{"fs.read"}},
		{Package: "p", Function: "Empty", Effects: []analyzer.Effect{}},
	}}
	var out bytes.Buffer
	if err := WriteEffects(&out, r, false); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"p.Candidate effects=[] possible=[fs.read] uncertain=false\n",
		"p.Empty effects=[] uncertain=false\n",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("inventory lost evidence; want %q:\n%s", want, out.String())
		}
	}
}

func TestCategoryReferencesIdentifyOperationsWithoutFollowingIDs(t *testing.T) {
	const id = "site:shared"
	r := &analyzer.EffectReport{
		Groups: []analyzer.EffectGroup{
			{Effect: "fs.read", Sites: []string{id}},
			{Effect: "resource.acquire", Sites: []string{id}},
		},
		Sites: []analyzer.Site{{
			ID: id, Symbol: analyzer.Symbol{Package: "example.com/store", Function: "Load"},
			APIs: []string{"os.Open"}, Location: analyzer.Location{File: "store.go", Line: 12, Column: 8},
		}},
	}
	var out bytes.Buffer
	if err := WriteEffects(&out, r, false); err != nil {
		t.Fatal(err)
	}
	header, _, found := strings.Cut(out.String(), "\n"+id+"\n")
	if !found {
		t.Fatal("missing canonical site definition")
	}
	for _, effect := range []string{"fs.read", "resource.acquire"} {
		want := effect + " (1 sites)\n  os.Open — store.go:12:8 in example.com/store.Load\n    " + id
		if !strings.Contains(header, want) {
			t.Fatalf("category lacks operation evidence %q:\n%s", want, header)
		}
	}
	if strings.Count(out.String(), "\n"+id+"\n") != 1 {
		t.Fatal("multi-category references duplicated the canonical definition")
	}
}
