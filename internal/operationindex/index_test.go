package operationindex

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func fixture(t *testing.T) object {
	t.Helper()
	raw := `{"schema_version":1,"sites":[],"groups":[],"reverse":[],"nodes":[],"edges":[],"unresolved_sites":["site:3","site:4"],"coverage":{"packages":["p"],"sites_emitted":5,"sites_matching_filters":5,"operation_sites_identified":2,"selected_operation_sites":1,"dependency_operation_sites":1,"unresolved_sites_identified":2,"future":{"big":123456789012345678901}},"truncated":{"sites":false,"edges":false,"depth":true},"interface_contracts":[{"site":"site:0","open":true,"future_contract":[1,2]}]}`
	value, err := Decode([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	d := obj(value)
	for i := 0; i < 5; i++ {
		sid := []string{"site:0", "site:1", "site:2", "site:3", "site:4"}[i]
		nid := "node:" + sid[5:]
		effects, possible := []any{}, []any{}
		if i == 0 {
			effects = []any{"a", "b"}
		}
		if i == 1 {
			effects = []any{"a"}
		}
		if i == 2 || i == 3 {
			possible = []any{"c"}
		}
		d["sites"] = append(arr(d["sites"]), object{"id": sid, "package": "p", "function": "F", "kind": "call", "apis": []any{"API"}, "operations": []any{object{"future_operation": json.Number("123456789012345678901")}}, "location": object{"file": "f.go", "line": json.Number("1"), "column": json.Number("1")}, "effects": effects, "possible_effects": possible, "uncertain": i == 0 || i >= 3, "uncertainty_reasons": []any{}, "recognition": "summary", "future_site": object{"x": "retained"}})
		d["nodes"] = append(arr(d["nodes"]), object{"id": nid, "package": "p", "function": "F", "selected": i != 1, "future_node": "retained"})
		origins := []any{nid}
		if i == 2 {
			origins = []any{}
		}
		d["reverse"] = append(arr(d["reverse"]), object{"site": sid, "origins": origins, "nodes": []any{object{"id": nid, "depth": json.Number("0")}}, "edges": []any{}, "truncated": d["truncated"], "future_reverse": "retained"})
	}
	return d
}
func loadFixture(t *testing.T, d object) *Report {
	t.Helper()
	raw, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	r, err := Load(raw)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func provenance(t *testing.T, hash string, flags ...string) any {
	t.Helper()
	args := []string{"/bin/sidefx", "-view", "effects", "-json", "-max-sites", "10000"}
	args = append(args, flags...)
	args = append(args, "./...")
	raw, _ := json.Marshal(object{"input_sha256": hash, "producer": object{"sha256": strings.Repeat("a", 64)}, "build": object{"GOOS": "local"}, "cwd": "/fixture", "exit": 0, "command": args})
	p, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func TestSavedReportCountsAndFidelity(t *testing.T) {
	d := fixture(t)
	r := loadFixture(t, d)
	index, err := r.Index(nil)
	if err != nil {
		t.Fatal(err)
	}
	if index["category_count"] != 3 || index["categorized_distinct_sites"] != 4 || index["category_site_memberships"] != 5 || index["multi_category_sites"] != 1 {
		t.Fatal(index)
	}
	rows := index["categories"].([]any)
	a, c := obj(rows[0]), obj(rows[2])
	if a["established"] != 2 || a["selected_origin"] != 1 || a["dependency_origin"] != 1 || c["unknown_origin"] != 1 || c["possible_only"] != 2 || c["unresolved"] != 1 {
		t.Fatal(rows)
	}
	bounds := obj(index["bounds"])
	if bounds["uncategorized_unresolved_records"] != 1 || bounds["possible_only_unresolved_overlap"] != 1 || bounds["declared_query_categories"] != "not_established" {
		t.Fatal(bounds)
	}
	first, err := r.Page(Selector{Category: "a"}, 0, 1, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := r.Page(Selector{Category: "a"}, 0, 1, first["next_cursor"].(string), nil)
	if err != nil {
		t.Fatal(err)
	}
	if second["next_cursor"] != nil || second["next_offset"] != nil {
		t.Fatal(second)
	}
	got := append(first["sites"].([]any), second["sites"].([]any)...)
	if !reflect.DeepEqual(got, arr(d["sites"])[:2]) || !reflect.DeepEqual(first["reverse"], arr(d["reverse"])[:1]) || !reflect.DeepEqual(first["interface_contracts"], arr(d["interface_contracts"])) {
		t.Fatal("records changed")
	}
	if !reflect.DeepEqual(obj(first["bounds"])["coverage"], d["coverage"]) {
		t.Fatal("unknown coverage lost")
	}
	for _, sel := range []Selector{{Site: "site:3"}, {Uncategorized: true}} {
		p, err := r.Page(sel, 0, 10, "", nil)
		if err != nil || p["total"] != 1 {
			t.Fatal(p, err)
		}
	}
}
func TestSavedReportMalformedInputs(t *testing.T) {
	cases := map[string]func(object){
		"version":         func(d object) { d["schema_version"] = 2 },
		"duplicate":       func(d object) { d["sites"] = append(arr(d["sites"]), arr(d["sites"])[0]) },
		"effects type":    func(d object) { obj(arr(d["sites"])[0])["effects"] = nil },
		"duplicate label": func(d object) { obj(arr(d["sites"])[0])["effects"] = []any{"a", "a"} },
		"overlap":         func(d object) { obj(arr(d["sites"])[0])["possible_effects"] = []any{"a"} },
		"unresolved":      func(d object) { d["unresolved_sites"] = []any{} },
		"coverage":        func(d object) { obj(d["coverage"])["sites_emitted"] = 6 },
		"truncation":      func(d object) { obj(d["truncated"])["sites"] = true },
		"dangling global": func(d object) { d["edges"] = []any{object{"id": "e", "caller": "absent", "callee": "node:0"}} },
		"reverse endpoint": func(d object) {
			d["edges"] = []any{object{"id": "e", "caller": "node:1", "callee": "node:0"}}
			obj(arr(d["reverse"])[0])["edges"] = []any{"e"}
		},
		"dangling origin": func(d object) { obj(arr(d["reverse"])[0])["origins"] = []any{"missing"} },
		"group":           func(d object) { d["groups"] = []any{object{"effect": "a", "sites": []any{"site:4"}}} },
		"contract":        func(d object) { d["interface_contracts"] = []any{object{"site": "missing"}} },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			d := fixture(t)
			change(d)
			raw, _ := json.Marshal(d)
			if _, err := Load(raw); err == nil {
				t.Fatal("accepted malformed input")
			}
		})
	}
	for _, raw := range []string{"null", "[]", "{}", "{", "{} {}"} {
		if _, err := Load([]byte(raw)); err == nil {
			t.Fatal(raw)
		}
	}
}
func TestSavedReportBoundsProvenanceAndPaginationGuards(t *testing.T) {
	r := loadFixture(t, fixture(t))
	p := provenance(t, r.SHA256)
	index, err := r.Index(p)
	if err != nil || obj(index["bounds"])["declared_query_categories"] != "complete" {
		t.Fatal(index, err)
	}
	for _, flags := range [][]string{{"-effect", "a"}, {"-contracts"}, {"-future-flag"}, {"-json"}} {
		p := provenance(t, r.SHA256, flags...)
		i, err := r.Index(p)
		if err != nil || obj(i["bounds"])["declared_query_categories"] != "not_established" || !reflect.DeepEqual(obj(i["bounds"])["acquisition"], p) {
			t.Fatal(i, err)
		}
	}
	for _, p := range []any{provenance(t, "bad"), object{}, "bad", provenance(t, r.SHA256, "-max-depth", "-1")} {
		if _, err := r.Index(p); err == nil {
			t.Fatal("invalid provenance accepted")
		}
	}
	d := fixture(t)
	obj(d["coverage"])["sites_matching_filters"] = 6
	obj(d["truncated"])["sites"] = true
	capped := loadFixture(t, d)
	i, err := capped.Index(provenance(t, capped.SHA256))
	if err != nil || obj(i["bounds"])["declared_query_categories"] != "not_established" {
		t.Fatal(i, err)
	}
	page, _ := r.Page(Selector{Category: "a"}, 0, 1, "", nil)
	token := page["next_cursor"].(string)
	for _, test := range []struct {
		sel           Selector
		offset, limit int
		token         string
	}{{Selector{Category: "missing"}, 0, 1, ""}, {Selector{Category: "a"}, -1, 1, ""}, {Selector{Category: "a"}, 3, 1, ""}, {Selector{Category: "a"}, 0, 101, ""}, {Selector{}, 0, 1, ""}, {Selector{Category: "a", Site: "site:0"}, 0, 1, ""}, {Selector{Category: "b"}, 0, 1, token}, {Selector{Category: "a"}, 1, 1, token}, {Selector{Category: "a"}, 0, 1, "broken"}} {
		if _, err := r.Page(test.sel, test.offset, test.limit, test.token, nil); err == nil {
			t.Fatal("bad page accepted", test)
		}
	}
	if _, err := capped.Page(Selector{Category: "a"}, 0, 1, token, nil); err == nil {
		t.Fatal("stale cursor accepted")
	}
}
func TestSavedReportOwnReverseGraphsClosed(t *testing.T) {
	d := fixture(t)
	d["edges"] = []any{object{"id": "e", "caller": "node:1", "callee": "node:0", "candidate": true, "future_edge": []any{"retained"}}}
	rev := obj(arr(d["reverse"])[0])
	rev["edges"] = []any{"e"}
	rev["nodes"] = append(arr(rev["nodes"]), object{"id": "node:1", "depth": json.Number("1")})
	r := loadFixture(t, d)
	p, err := r.Page(Selector{Site: "site:0"}, 0, 1, "", nil)
	if err != nil || len(p["nodes"].([]any)) != 2 || !reflect.DeepEqual(p["edges"], d["edges"]) {
		t.Fatal(p, err)
	}
}
