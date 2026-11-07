package main

import (
	"slices"
	"testing"

	"github.com/Version2beta/sidefx/internal/analyzer"
)

func TestOperationCoverageContextTimerConditions(t *testing.T) {
	dir := operationCoverageFixture(t)
	r, _ := effectResult(t, dir, "-all", "-max-depth", "2")
	inv := inventoryByFunction(r)
	for _, fn := range []string{"ContextTimeout", "ContextTimeoutCause", "ContextDeadline", "ContextDeadlineCause"} {
		rec := inv[fn]
		if !rec.Uncertain || !slices.EqualFunc(rec.Effects, []analyzer.Effect{{Identifier: "concurrency.timer.schedule", Direct: true}}, func(a, b analyzer.Effect) bool { return a == b }) || len(rec.PossibleEffects) > 0 {
			t.Errorf("conditional timer %s: %+v", fn, rec)
		}
	}
	for _, fn := range []string{"ContextExpired", "ContextNil", "ContextCancelOnly"} {
		if slices.ContainsFunc(inv[fn].Effects, func(e analyzer.Effect) bool { return e.Identifier == "concurrency.timer.schedule" }) {
			t.Errorf("timer invented for %s: %+v", fn, inv[fn])
		}
	}
	if !inv["ContextExpired"].Uncertain {
		t.Error("expired context lost custom-parent opacity")
	}
	if !slices.ContainsFunc(inv["ContextCancelOnly"].Effects, func(e analyzer.Effect) bool { return e.Identifier == "context.cancel" }) {
		t.Error("cancel provenance lost")
	}
	if !slices.ContainsFunc(inv["ContextTimerCaller"].Effects, func(e analyzer.Effect) bool { return e.Identifier == "concurrency.timer.schedule" && !e.Direct }) {
		t.Error("timer inheritance lost")
	}
	q, _ := effectResult(t, dir, "-api", "context.WithTimeout", "-reason", "state_condition", "-function", "=ContextTimeout", "-max-depth", "1")
	if len(q.Sites) != 1 || !q.Sites[0].Uncertain || len(q.Sites[0].PossibleEffects) > 0 || q.Sites[0].Location.Line < 1 || q.Sites[0].Location.Column < 1 {
		t.Fatalf("conditional API/source retrieval: %+v", q)
	}
	no, _ := effectResult(t, dir, "-unresolved", "-api", "context.WithTimeout", "-function", "=ContextTimeout")
	if len(no.Sites) != 0 {
		t.Error("known conditional operation is not unresolved-only", no.Sites)
	}
	glob, _ := effectResult(t, dir, "-api", "io/fs.(GlobFS).Glob")
	if len(glob.Sites) != 1 || !glob.Sites[0].Uncertain || !slices.Contains(glob.Sites[0].Effects, "fs.read") {
		t.Error("public declaring Glob contract missing", glob.Sites)
	}
}
