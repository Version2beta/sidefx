package operationindex

import (
	"encoding/json"
	"testing"
)

func TestSavedReportRejectsMalformedKnownGraphMetadata(t *testing.T) {
	cases := map[string]func(object){
		"node package number":    func(d object) { obj(arr(d["nodes"])[0])["package"] = 17 },
		"node function array":    func(d object) { obj(arr(d["nodes"])[0])["function"] = []any{"F"} },
		"node missing package":   func(d object) { delete(obj(arr(d["nodes"])[0]), "package") },
		"node missing function":  func(d object) { delete(obj(arr(d["nodes"])[0]), "function") },
		"node synthetic string":  func(d object) { obj(arr(d["nodes"])[0])["synthetic"] = "false" },
		"node uncertain array":   func(d object) { obj(arr(d["nodes"])[0])["uncertain"] = []any{} },
		"node variant number":    func(d object) { obj(arr(d["nodes"])[0])["variant"] = 17 },
		"edge candidate string":  func(d object) { obj(arr(d["edges"])[0])["candidate"] = "false" },
		"edge candidate number":  func(d object) { obj(arr(d["edges"])[0])["candidate"] = 1 },
		"edge synthetic null":    func(d object) { obj(arr(d["edges"])[0])["synthetic"] = nil },
		"edge source string":     func(d object) { obj(arr(d["edges"])[0])["source"] = "not an object" },
		"reverse revisit number": func(d object) { obj(arr(d["reverse"])[0])["revisit"] = 1 },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			d := fixture(t)
			d["edges"] = []any{object{"id": "edge:0", "caller": "node:0", "callee": "node:0", "candidate": false}}
			change(d)
			raw, _ := json.Marshal(d)
			if _, err := Load(raw); err == nil {
				t.Fatal("malformed known metadata accepted")
			}
		})
	}
	for name, value := range map[string]any{"negative": -1, "fraction": 0.5, "string": "0", "array": []any{}, "null": nil, "bool": false} {
		t.Run("reverse depth "+name, func(t *testing.T) {
			d := fixture(t)
			obj(arr(obj(arr(d["reverse"])[0])["nodes"])[0])["depth"] = value
			raw, _ := json.Marshal(d)
			if _, err := Load(raw); err == nil {
				t.Fatal("malformed reverse depth accepted")
			}
		})
	}
	t.Run("missing reverse depth", func(t *testing.T) {
		d := fixture(t)
		delete(obj(arr(obj(arr(d["reverse"])[0])["nodes"])[0]), "depth")
		raw, _ := json.Marshal(d)
		if _, err := Load(raw); err == nil {
			t.Fatal("missing reverse depth accepted")
		}
	})
}

func TestSavedReportValidSyntheticAndUnknownOriginMetadata(t *testing.T) {
	d := fixture(t)
	node := obj(arr(d["nodes"])[0])
	node["package"] = ""
	node["function"] = ""
	node["synthetic"] = true
	// Empty symbols are valid synthetic metadata, not malformed types. A
	// nonmatching or missing origin remains unknown; optional legacy metadata
	// may be absent. Unknown numeric fields must still round-trip precisely.
	node["future_number"] = json.Number("123456789012345678901")
	d["edges"] = []any{object{"id": "edge:0", "caller": "node:0", "callee": "node:0"}}
	r := loadFixture(t, d)
	index, err := r.Index(nil)
	if err != nil {
		t.Fatal(err)
	}
	if obj(index["categories"].([]any)[0])["unknown_origin"] != 1 {
		t.Fatal("valid synthetic origin was not unknown")
	}
	page, err := r.Page(Selector{Site: "site:0"}, 0, 1, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if obj(page["nodes"].([]any)[0])["future_number"] != json.Number("123456789012345678901") {
		t.Fatal("unknown numeric precision lost")
	}
}
