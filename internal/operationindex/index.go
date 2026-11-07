// Package operationindex projects saved effect-view JSON without loading source
// or changing recognition. Records stay untyped to retain unknown evidence fields.
package operationindex

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

type object = map[string]any
type set = map[string]bool

type Report struct {
	data                                    object
	sites, nodes, edges, reverse, contracts map[string]object
	categories                              map[string]map[string]set
	unresolved, uncategorized               set
	SHA256                                  string
}

type validation struct{ err error }

func (v *validation) require(ok bool, message string) {
	if !ok && v.err == nil {
		v.err = fmt.Errorf("%s", message)
	}
}
func obj(a any) object   { x, _ := a.(map[string]any); return x }
func arr(a any) []any    { x, _ := a.([]any); return x }
func str(a any) string   { x, _ := a.(string); return x }
func boolean(a any) bool { x, _ := a.(bool); return x }
func integer(a any) (int, bool) {
	n, ok := a.(json.Number)
	if !ok {
		return 0, false
	}
	i, e := strconv.Atoi(string(n))
	return i, e == nil
}
func (v *validation) strings(a any, label string) set {
	x, ok := a.([]any)
	v.require(ok, label+": expected string array")
	out := set{}
	for _, item := range x {
		s, ok := item.(string)
		v.require(ok && s != "" && !out[s], label+": invalid or duplicate string")
		out[s] = true
	}
	return out
}
func (v *validation) objects(a any, key, label string) map[string]object {
	x, ok := a.([]any)
	v.require(ok, label+": expected object array")
	out := map[string]object{}
	for _, item := range x {
		m := obj(item)
		id := str(m[key])
		v.require(m != nil && id != "" && out[id] == nil, label+": invalid or duplicate "+key)
		out[id] = m
	}
	return out
}
func (v *validation) optionalBooleans(m object, label string, fields ...string) {
	for _, field := range fields {
		if value, exists := m[field]; exists {
			_, ok := value.(bool)
			v.require(ok, label+": invalid "+field)
		}
	}
}
func (v *validation) optionalStrings(m object, label string, fields ...string) {
	for _, field := range fields {
		if value, exists := m[field]; exists {
			_, ok := value.(string)
			v.require(ok, label+": invalid "+field)
		}
	}
}
func (v *validation) symbol(m object, label string) {
	for _, field := range []string{"package", "function"} {
		_, ok := m[field].(string)
		// Synthetic symbols can be empty, but malformed/missing identity types
		// must not silently become valid unknown-origin evidence.
		v.require(ok, label+": invalid "+field)
	}
}
func (v *validation) truncation(a any) {
	m := obj(a)
	for _, k := range []string{"sites", "edges", "depth"} {
		_, ok := m[k].(bool)
		v.require(ok, "invalid graph truncation")
	}
}
func keys[T any](m map[string]T) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
func union(a, b set) set {
	c := set{}
	for k := range a {
		c[k] = true
	}
	for k := range b {
		c[k] = true
	}
	return c
}
func subset(a, b set) bool {
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}
func same(a, b set) bool { return len(a) == len(b) && subset(a, b) }
func objectIDs(m map[string]object) set {
	out := set{}
	for k := range m {
		out[k] = true
	}
	return out
}

// Decode accepts exactly one JSON value; numbers and unknown fields are retained.
func Decode(raw []byte) (any, error) {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var value any
	if err := d.Decode(&value); err != nil {
		return nil, err
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("expected one JSON value")
	}
	return value, nil
}

func Load(raw []byte) (*Report, error) {
	value, err := Decode(raw)
	if err != nil {
		return nil, err
	}
	r := &Report{data: obj(value), SHA256: fmt.Sprintf("%x", sha256.Sum256(raw)), categories: map[string]map[string]set{}, uncategorized: set{}}
	v := &validation{}
	version, ok := integer(r.data["schema_version"])
	v.require(ok && version == 1, "expected effect-view schema_version 1")
	r.sites = v.objects(r.data["sites"], "id", "sites")
	r.nodes = v.objects(r.data["nodes"], "id", "nodes")
	r.edges = v.objects(r.data["edges"], "id", "edges")
	r.reverse = v.objects(r.data["reverse"], "site", "reverse")
	v.require(same(objectIDs(r.reverse), objectIDs(r.sites)), "reverse/site sets differ")
	expectedUnresolved := set{}
	for sid, s := range r.sites {
		for _, k := range []string{"package", "function", "kind", "recognition"} {
			v.require(str(s[k]) != "", sid+": invalid "+k)
		}
		_, ok := s["uncertain"].(bool)
		v.require(ok, sid+": invalid uncertain")
		v.strings(s["apis"], sid+".apis")
		v.strings(s["uncertainty_reasons"], sid+".uncertainty_reasons")
		loc := obj(s["location"])
		_, ok = loc["file"].(string)
		v.require(ok, sid+": invalid location")
		for _, k := range []string{"line", "column"} {
			n, ok := integer(loc[k])
			v.require(ok && n >= 0, sid+": invalid location")
		}
		ops, ok := s["operations"].([]any)
		v.require(ok, sid+": invalid operations")
		for _, op := range ops {
			v.require(obj(op) != nil, sid+": invalid operation")
		}
		e := v.strings(s["effects"], sid+".effects")
		p := set{}
		if x, exists := s["possible_effects"]; exists {
			p = v.strings(x, sid+".possible_effects")
		}
		for label := range e {
			v.require(!p[label], sid+": established/possible overlap")
		}
		for label := range union(e, p) {
			if r.categories[label] == nil {
				r.categories[label] = map[string]set{"established": {}, "possible_only": {}}
			}
			field := "possible_only"
			if e[label] {
				field = "established"
			}
			r.categories[label][field][sid] = true
		}
		if len(e) == 0 && boolean(s["uncertain"]) {
			expectedUnresolved[sid] = true
			if len(p) == 0 {
				r.uncategorized[sid] = true
			}
		}
	}
	for _, n := range r.nodes {
		v.symbol(n, "node")
		_, ok := n["selected"].(bool)
		v.require(ok, "node: invalid selected")
		v.optionalStrings(n, "node", "variant")
		v.optionalBooleans(n, "node", "synthetic", "uncertain", "direct_uncertainty", "inherited_uncertainty")
	}
	for _, e := range r.edges {
		v.require(r.nodes[str(e["caller"])] != nil && r.nodes[str(e["callee"])] != nil, "edge: dangling node")
		v.optionalBooleans(e, "edge", "synthetic", "candidate")
		if source, exists := e["source"]; exists && source != nil {
			s := obj(source)
			v.require(s != nil, "edge: invalid source")
			v.optionalStrings(s, "edge source", "id", "kind", "resolution")
			for _, field := range []string{"caller", "callee"} {
				if symbol, exists := s[field]; exists {
					v.symbol(obj(symbol), "edge source "+field)
				}
			}
			if location, exists := s["location"]; exists {
				loc := obj(location)
				_, ok := loc["file"].(string)
				v.require(ok, "edge source: invalid location file")
				for _, field := range []string{"line", "column"} {
					n, ok := integer(loc[field])
					v.require(ok && n >= 0, "edge source: invalid location "+field)
				}
			}
		}
	}
	for sid, rev := range r.reverse {
		origins := v.strings(rev["origins"], sid+".origins")
		refs := v.objects(rev["nodes"], "id", sid+".nodes")
		for _, ref := range refs {
			depth, ok := integer(ref["depth"])
			v.require(ok && depth >= 0, sid+": invalid reverse node depth")
		}
		v.optionalBooleans(rev, sid, "revisit", "cycle")
		v.require(subset(origins, objectIDs(refs)) && subset(objectIDs(refs), objectIDs(r.nodes)), sid+": dangling reverse nodes")
		for eid := range v.strings(rev["edges"], sid+".edges") {
			edge := r.edges[eid]
			v.require(edge != nil, sid+": dangling reverse edge")
			v.require(refs[str(edge["caller"])] != nil && refs[str(edge["callee"])] != nil, sid+": reverse edge endpoint missing from node refs")
		}
		v.truncation(rev["truncated"])
	}
	for category, g := range v.objects(r.data["groups"], "effect", "groups") {
		for key, field := range map[string]string{"sites": "established", "possible_sites": "possible_only"} {
			ids := set{}
			if x, exists := g[key]; exists {
				ids = v.strings(x, "group "+category+"."+key)
			}
			v.require(subset(ids, r.categories[category][field]), "group "+category+": invalid "+key+" reference")
		}
	}
	r.unresolved = v.strings(r.data["unresolved_sites"], "unresolved_sites")
	v.require(same(r.unresolved, expectedUnresolved), "unresolved_sites differs from schema's effect-free uncertain sites")
	v.truncation(r.data["truncated"])
	r.contracts = map[string]object{}
	if x, exists := r.data["interface_contracts"]; exists {
		r.contracts = v.objects(x, "site", "interface_contracts")
	}
	v.require(subset(objectIDs(r.contracts), objectIDs(r.sites)), "contract: dangling site")
	c := obj(r.data["coverage"])
	v.require(c != nil, "missing coverage")
	v.strings(c["packages"], "coverage.packages")
	for _, k := range []string{"sites_emitted", "sites_matching_filters", "operation_sites_identified", "selected_operation_sites", "dependency_operation_sites", "unresolved_sites_identified"} {
		n, ok := integer(c[k])
		v.require(ok && n >= 0, "invalid coverage."+k)
	}
	emitted, _ := integer(c["sites_emitted"])
	matching, _ := integer(c["sites_matching_filters"])
	v.require(emitted == len(r.sites), "coverage sites_emitted mismatch")
	v.require(matching >= len(r.sites), "invalid matching site count")
	v.require(boolean(obj(r.data["truncated"])["sites"]) == (matching > len(r.sites)), "site truncation/count mismatch")
	if v.err != nil {
		return nil, v.err
	}
	return r, nil
}

func (r *Report) origin(sid string) string {
	s := r.sites[sid]
	found := false
	for _, id := range arr(r.reverse[sid]["origins"]) {
		n := r.nodes[str(id)]
		if n["package"] == s["package"] && n["function"] == s["function"] {
			found = true
			if boolean(n["selected"]) {
				return "selected_origin"
			}
		}
	}
	if found {
		return "dependency_origin"
	}
	return "unknown_origin"
}

func (r *Report) bounds(provenance any) (object, error) {
	query, reasons, err := DeclaredQuery(provenance, r.SHA256)
	if err != nil {
		return nil, err
	}
	noLoss := !boolean(obj(r.data["truncated"])["sites"])
	complete := "not_established"
	if query != nil && noLoss {
		complete = "complete"
	}
	if !noLoss {
		reasons = append(reasons, "input site output is truncated")
	}
	c := obj(r.data["coverage"])
	dep := c["dependency_vta"]
	if dep == nil {
		dep = "not exposed; default discovery remains selected-source bounded"
	}
	filters := "unknown; coverage alone cannot disclose filters"
	limits := "unavailable; truncation flags do not state numeric limits"
	if provenance != nil {
		filters = "see acquisition.command; completeness eligibility in orientation_limits"
		limits = "see acquisition.command; unspecified producer defaults unavailable"
	}
	other := 0
	for _, s := range r.sites {
		if len(arr(s["effects"])) == 0 && len(arr(s["possible_effects"])) == 0 && !boolean(s["uncertain"]) {
			other++
		}
	}
	return object{"emitted_records": "complete", "declared_query_categories": complete, "orientation_limits": reasons, "query": query, "acquisition": provenance,
		"analysis_coverage": "recognized vocabulary/current build only; not runtime or all mutations", "dependency_analysis_bounds": dep,
		"excluded_scope": "other builds, unselected roots and unmodeled operations; direct scope also excludes dependency-origin sites and candidate-only helper reachability",
		"registry":       "unavailable; absent labels are not recognized-zero categories", "source_filters": filters, "output_limits": limits, "truncated": r.data["truncated"], "coverage": c,
		"uncategorized_unresolved_records": len(r.uncategorized), "unresolved_view_records": len(r.unresolved), "uncategorized_other_records": other, "possible_only_unresolved_overlap": len(r.unresolved) - len(r.uncategorized), "unresolved_inventory": "not established; identified count is not an emitted inventory"}, nil
}

func (r *Report) Index(provenance any) (object, error) {
	bounds, err := r.bounds(provenance)
	if err != nil {
		return nil, err
	}
	rows := []any{}
	categorized := set{}
	memberships, multi := 0, 0
	for _, category := range keys(r.categories) {
		g := r.categories[category]
		ids := union(g["established"], g["possible_only"])
		categorized = union(categorized, ids)
		memberships += len(ids)
		counts := map[string]int{"selected_origin": 0, "dependency_origin": 0, "unknown_origin": 0}
		uncertain, unresolved := 0, 0
		for sid := range ids {
			counts[r.origin(sid)]++
			if boolean(r.sites[sid]["uncertain"]) {
				uncertain++
			}
			if r.unresolved[sid] {
				unresolved++
			}
		}
		rows = append(rows, object{"category": category, "sites": len(ids), "established": len(g["established"]), "possible_only": len(g["possible_only"]), "uncertain": uncertain, "unresolved": unresolved, "selected_origin": counts["selected_origin"], "dependency_origin": counts["dependency_origin"], "unknown_origin": counts["unknown_origin"]})
	}
	for _, s := range r.sites {
		if len(arr(s["effects"]))+len(arr(s["possible_effects"])) > 1 {
			multi++
		}
	}
	return object{"projection": "sidefx-operation-index-v1", "input_sha256": r.SHA256, "bounds": bounds, "categories": rows, "category_count": len(rows), "categorized_distinct_sites": len(categorized), "category_site_memberships": memberships, "multi_category_sites": multi,
		"count_definitions": "Distinct canonical site IDs per category; established/possible_only are origin labels, not caller-path certainty. possible_only is per category. uncertain overlaps both. Origin counts partition sites by matching reverse origin nodes; not direct/inherited function marks. Category counts overlap across multi-effect sites; do not sum as distinct sites."}, nil
}

// Selector is exact; it never interprets substrings or invents categories.
type Selector struct {
	Category      string `json:"category"`
	Site          string `json:"site"`
	Uncategorized bool   `json:"uncategorized_unresolved"`
}
type cursor struct {
	SHA256   string   `json:"sha256"`
	Selector Selector `json:"selector"`
	Offset   int      `json:"offset"`
}

func (r *Report) Page(sel Selector, offset, limit int, token string, provenance any) (object, error) {
	n := 0
	if sel.Category != "" {
		n++
	}
	if sel.Site != "" {
		n++
	}
	if sel.Uncategorized {
		n++
	}
	if n != 1 {
		return nil, fmt.Errorf("retrieval requires exactly one selector")
	}
	if token != "" {
		raw, err := base64.RawURLEncoding.DecodeString(token)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor")
		}
		var c cursor
		d := json.NewDecoder(bytes.NewReader(raw))
		d.DisallowUnknownFields()
		if d.Decode(&c) != nil || c.SHA256 != r.SHA256 || c.Selector != sel || c.Offset < 0 || offset != 0 {
			return nil, fmt.Errorf("stale/invalid cursor or selector mismatch; do not combine cursor and offset")
		}
		var extra any
		if d.Decode(&extra) != io.EOF {
			return nil, fmt.Errorf("invalid cursor")
		}
		offset = c.Offset
	}
	ids := r.uncategorized
	if sel.Category != "" {
		g, ok := r.categories[sel.Category]
		if !ok {
			return nil, fmt.Errorf("category not present: %s", sel.Category)
		}
		ids = union(g["established"], g["possible_only"])
	}
	if sel.Site != "" {
		if r.sites[sel.Site] == nil {
			return nil, fmt.Errorf("site not present: %s", sel.Site)
		}
		ids = set{sel.Site: true}
	}
	if offset < 0 || offset > len(ids) || limit < 1 || limit > 100 {
		return nil, fmt.Errorf("offset outside selected set or limit not 1..100")
	}
	bounds, err := r.bounds(provenance)
	if err != nil {
		return nil, err
	}
	end := min(offset+limit, len(ids))
	selected := keys(ids)[offset:end]
	sites, reverse, nodes, edges, contracts := []any{}, []any{}, []any{}, []any{}, []any{}
	nids, eids := set{}, set{}
	for _, sid := range selected {
		sites = append(sites, r.sites[sid])
		rev := r.reverse[sid]
		reverse = append(reverse, rev)
		for _, ref := range arr(rev["nodes"]) {
			nids[str(obj(ref)["id"])] = true
		}
		for _, id := range arr(rev["edges"]) {
			eids[str(id)] = true
		}
	}
	for _, id := range keys(nids) {
		nodes = append(nodes, r.nodes[id])
	}
	for _, id := range keys(eids) {
		edges = append(edges, r.edges[id])
	}
	// Preserve contract array order, as in the saved report/reference projection.
	selectedIDs := set{}
	for _, id := range selected {
		selectedIDs[id] = true
	}
	for _, c := range arr(r.data["interface_contracts"]) {
		if selectedIDs[str(obj(c)["site"])] {
			contracts = append(contracts, c)
		}
	}
	var next, nextCursor any
	if end < len(ids) {
		next = end
		raw, _ := json.Marshal(cursor{r.SHA256, sel, end})
		nextCursor = base64.RawURLEncoding.EncodeToString(raw)
	}
	var category, site any
	if sel.Category != "" {
		category = sel.Category
	}
	if sel.Site != "" {
		site = sel.Site
	}
	return object{"projection": "sidefx-operation-sites-v1", "input_sha256": r.SHA256, "selector": object{"category": category, "site": site, "uncategorized_unresolved": sel.Uncategorized}, "total": len(ids), "offset": offset, "limit": limit, "next_offset": next, "next_cursor": nextCursor, "bounds": bounds, "sites": sites, "reverse": reverse, "nodes": nodes, "edges": edges, "interface_contracts": contracts}, nil
}

// DeclaredQuery validates hash-bound caller-supplied provenance independently of
// its eligibility for unfiltered orientation. It does not authenticate execution.
func DeclaredQuery(provenance any, digest string) (any, []string, error) {
	if provenance == nil {
		return nil, []string{"producer command/build/scope provenance unavailable"}, nil
	}
	p := obj(provenance)
	v := &validation{}
	v.require(p != nil, "provenance: expected object")
	v.require(p["input_sha256"] == digest, "provenance: input SHA-256 mismatch")
	for _, key := range []string{"producer", "build"} {
		v.require(len(obj(p[key])) > 0, "provenance: missing/invalid "+key+" object")
	}
	hash := str(obj(p["producer"])["sha256"])
	validHash := len(hash) == 64
	for _, c := range hash {
		validHash = validHash && strings.ContainsRune("0123456789abcdef", c)
	}
	v.require(validHash, "provenance: invalid producer SHA-256")
	v.require(str(p["cwd"]) != "", "provenance: missing/invalid cwd")
	exit, ok := integer(p["exit"])
	v.require(ok && exit == 0, "provenance: acquisition did not succeed")
	argv, ok := p["command"].([]any)
	v.require(ok && len(argv) > 1, "provenance: command must be argv array")
	for _, arg := range argv {
		v.require(str(arg) != "", "provenance: invalid argv")
	}
	if v.err != nil {
		return nil, nil, v.err
	}
	booleans := set{"json": true, "direct": true, "tests": true, "exact": true, "stats": true, "all": true}
	valued := set{"view": true, "max-sites": true, "max-edges": true, "max-depth": true, "dependency-vta-functions": true, "dependency-vta-edges": true}
	values := object{}
	patterns := []string{}
	for i := 1; i < len(argv); i++ {
		arg := str(argv[i])
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			for _, p := range argv[i:] {
				patterns = append(patterns, str(p))
			}
			break
		}
		name, value, equals := strings.Cut(strings.TrimLeft(arg, "-"), "=")
		_, duplicate := values[name]
		if (!booleans[name] && !valued[name]) || duplicate {
			return nil, []string{"filtered/unsupported or duplicate acquisition flag: " + name}, nil
		}
		if booleans[name] {
			v.require(!equals || value == "true" || value == "false", "provenance: invalid -"+name)
			values[name] = value != "false"
		} else {
			if !equals {
				i++
				if i < len(argv) {
					value = str(argv[i])
				}
			}
			v.require(value != "", "provenance: missing -"+name+" value")
			if name != "view" {
				num, err := strconv.Atoi(value)
				digits := value != ""
				for _, c := range value {
					digits = digits && c >= '0' && c <= '9'
				}
				minimum := 1
				if name == "max-depth" {
					minimum = 0
				}
				v.require(digits && err == nil && num >= minimum, "provenance: invalid -"+name)
			}
			values[name] = value
		}
	}
	v.require(len(patterns) > 0, "provenance: explicit package scope required; flags must precede packages")
	for _, p := range patterns {
		v.require(!strings.HasPrefix(p, "-"), "provenance: flags must precede packages")
	}
	v.require(values["view"] == "effects" && values["json"] == true, "provenance: require -view effects -json")
	v.require(values["max-sites"] != nil, "provenance: record explicit positive -max-sites")
	v.require((values["dependency-vta-functions"] != nil) == (values["dependency-vta-edges"] != nil), "provenance: dependency VTA budgets must be paired")
	if v.err != nil {
		return nil, nil, v.err
	}
	scope := "reachable from selected roots (including candidate paths)"
	if values["direct"] == true {
		scope = "direct selected roots"
	}
	return object{"scope": scope, "patterns": patterns, "flags": values, "authority": "caller-supplied acquisition record, hash-bound but not authenticated"}, []string{}, nil
}
