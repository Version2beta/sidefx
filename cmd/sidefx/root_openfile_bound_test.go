package main

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/Version2beta/sidefx/internal/analyzer"
)

func TestRootOpenFileBoundArgumentShapes(t *testing.T) {
	dir := operationCoverageFixture(t)
	r, _ := effectResult(t, dir, "-all", "-max-depth", "2")
	inv := inventoryByFunction(r)
	read := []string{"fs.read", "resource.acquire"}
	write := []string{"fs.create", "fs.write", "resource.acquire"}
	dynamic := []string{"fs.create", "fs.read", "fs.write", "resource.acquire"}
	for _, c := range []struct {
		fn   string
		want []string
	}{
		{"RootBoundWrite", write}, {"RootBoundDynamic", dynamic}, {"RootBoundReadMode", read},
		{"RootExpressionWrite", write}, {"RootExpressionDynamic", dynamic}, {"RootExpressionReadMode", read},
		{"RootBoundCaller", write}, {"RootRead", read}, {"RootWrite", write}, {"RootDynamic", dynamic},
	} {
		t.Run(c.fn, func(t *testing.T) {
			var got []string
			for _, e := range inv[c.fn].Effects {
				got = append(got, e.Identifier)
			}
			if !slices.Equal(got, c.want) || len(inv[c.fn].PossibleEffects) > 0 || inv[c.fn].Uncertain {
				t.Errorf("%s actual=%+v want=%v", c.fn, inv[c.fn], c.want)
			}
		})
	}
	candidate := inv["RootBoundAlternatives"]
	if len(candidate.Effects) != 0 || !slices.Equal(candidate.PossibleEffects, write) || candidate.Uncertain {
		t.Errorf("bound alternatives must stay possible-only with actual flags: %+v", candidate)
	}
	q, _ := effectResult(t, dir, "-api", "os.(*Root).OpenFile", "-function", "=RootBoundWrite")
	if len(q.Sites) != 1 || !slices.Equal(q.Sites[0].Effects, write) {
		t.Fatalf("bound API retrieval: %+v", q.Sites)
	}
	site := q.Sites[0]
	src, err := os.ReadFile(site.Location.File)
	if err != nil {
		t.Fatal(err)
	}
	line := strings.Split(string(src), "\n")[site.Location.Line-1]
	if line[site.Location.Column-1] != '(' || !strings.Contains(line, "f(") {
		t.Fatalf("not bound invocation location: %+v", site)
	}
	for _, args := range [][]string{{"-effect", "fs.write", "-function", "=RootBoundWrite"}, {"-site", site.ID}, {"-effect", "fs.create", "-function", "=RootBoundCaller", "-max-depth", "1"}} {
		selected, _ := effectResult(t, dir, args...)
		found := false
		for _, s := range selected.Sites {
			found = found || s.ID == site.ID
		}
		if !found {
			t.Errorf("bound site lost by filter %v", args)
		}
	}
	if !slices.ContainsFunc(inv["RootBoundCaller"].Effects, func(e analyzer.Effect) bool { return e.Identifier == "fs.write" && !e.Direct }) {
		t.Error("wrapper lost inherited write", inv["RootBoundCaller"])
	}
}
