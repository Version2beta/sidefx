package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/Version2beta/sidefx/internal/analyzer"
)

func operationCoverageFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.CopyFS(dir, os.DirFS("../../testdata/operation-coverage")); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestOperationCoverageExactFamilies(t *testing.T) {
	dir := operationCoverageFixture(t)
	before := snapshot(t, dir)
	r, _ := effectResult(t, dir, "-all", "-max-depth", "2")
	type want struct{ fn, api, effect string }
	cases := []want{
		{"CopyFile", "os.(*File).ReadFrom", "fs.write"}, {"CopyFile", "os.(*File).ReadFrom", "io.read"},
		{"CopyOut", "os.(*File).WriteTo", "fs.read"}, {"CopyOut", "os.(*File).WriteTo", "io.write"},
		{"CopyTree", "os.CopyFS", "fs.create"}, {"WalkTree", "path/filepath.Walk", "fs.read"}, {"WalkDirTree", "path/filepath.WalkDir", "fs.read"},
		{"FSFacade", "io/fs.ReadFile", "fs.read"}, {"FSFacade", "io/fs.ReadLink", "fs.read"}, {"FSOpen", "io/fs.(FS).Open", "resource.acquire"},
		{"FSInterfaces", "io/fs.(ReadFileFS).ReadFile", "fs.read"}, {"FSInterfaces", "io/fs.(ReadDirFS).ReadDir", "fs.read"}, {"FSInterfaces", "io/fs.(StatFS).Stat", "fs.read"}, {"FSInterfaces", "io/fs.(ReadLinkFS).ReadLink", "fs.read"}, {"FSInterfaces", "io/fs.(File).Read", "fs.read"}, {"FSInterfaces", "io/fs.(File).Close", "resource.release"},
		{"ZipOpen", "archive/zip.OpenReader", "resource.acquire"}, {"ZipClose", "archive/zip.(*ReadCloser).Close", "resource.release"}, {"FileSync", "os.(*File).Sync", "fs.sync"}, {"FileSync", "os.(*File).Sync", "fs.write"},
		{"ReadStream", "io.(Reader).Read", "io.read"}, {"WriteStream", "io.(Writer).Write", "io.write"},
		{"IOInterfaces", "io.(ReaderAt).ReadAt", "io.read"}, {"IOInterfaces", "io.(ByteReader).ReadByte", "io.read"}, {"IOInterfaces", "io.(RuneReader).ReadRune", "io.read"}, {"IOInterfaces", "io.(WriterAt).WriteAt", "io.write"}, {"IOInterfaces", "io.(StringWriter).WriteString", "io.write"}, {"IOInterfaces", "io.(ByteWriter).WriteByte", "io.write"}, {"IOInterfaces", "io.(ReaderFrom).ReadFrom", "io.read"}, {"IOInterfaces", "io.(WriterTo).WriteTo", "io.write"},
		{"Transfer", "io.Copy", "io.read"}, {"Transfer", "io.Copy", "io.write"}, {"Transfer", "io.ReadFull", "io.read"}, {"Transfer", "io.WriteString", "io.write"},
		{"Buffered", "bufio.(*Reader).Read", "io.read"}, {"Buffered", "bufio.(*Writer).Flush", "io.write"}, {"Buffered", "bufio.(*Scanner).Scan", "io.read"},
		{"JSONStream", "encoding/json.(*Decoder).Decode", "io.read"}, {"JSONStream", "encoding/json.(*Encoder).Encode", "io.write"},
		{"XMLStream", "encoding/xml.(*Decoder).Decode", "io.read"}, {"XMLStream", "encoding/xml.(*Encoder).Encode", "io.write"},
		{"GobStream", "encoding/gob.(*Decoder).Decode", "io.read"}, {"GobStream", "encoding/gob.(*Encoder).Encode", "io.write"},
		{"TextOutput", "text/template.(*Template).ExecuteTemplate", "io.write"}, {"HTMLOutput", "html/template.(*Template).Execute", "io.write"},
		{"GzipStream", "compress/gzip.(*Reader).Read", "io.read"}, {"GzipStream", "compress/gzip.(*Writer).Close", "io.write"}, {"GzipHeader", "compress/gzip.NewReader", "io.read"},
		{"FlateStream", "compress/flate.(*Writer).Flush", "io.write"}, {"ZlibStream", "compress/zlib.(*Writer).Close", "io.write"}, {"ZlibHeader", "compress/zlib.NewReader", "io.read"},
		{"TarStream", "archive/tar.(*Reader).Next", "io.read"}, {"TarStream", "archive/tar.(*Writer).WriteHeader", "io.write"}, {"PrintOutput", "fmt.Fprintln", "io.write"}, {"PrintOutput", "builtin.print", "io.write"}, {"PrintOutput", "builtin.println", "io.write"},
		{"HTTPHelpers", "net/http.Error", "http.response"}, {"HTTPHelpers", "net/http.NotFound", "http.response"}, {"HTTPHelpers", "net/http.Redirect", "http.response"}, {"HTTPHelpers", "net/http.ServeContent", "http.response"}, {"HTTPHelpers", "net/http.ServeFile", "http.response"}, {"HTTPHelpers", "net/http.ServeFile", "fs.read"}, {"HTTPHelpers", "net/http.ServeFileFS", "http.response"},
		{"HeaderPrep", "net/http.(Header).Set", "http.header.write"}, {"HeaderPrep", "net/http.SetCookie", "http.header.write"}, {"HeaderOutput", "net/http.(Header).Write", "io.write"},
		{"Register", "net/http.HandleFunc", "http.handler.register"}, {"Register", "net/http.(*ServeMux).Handle", "http.handler.register"}, {"Proxy", "net/http/httputil.(*ReverseProxy).ServeHTTP", "net.connect"}, {"Proxy", "net/http/httputil.(*ReverseProxy).ServeHTTP", "http.response"},
		{"Logging", "log.Print", "log.write"}, {"Logging", "log.(*Logger).Output", "log.write"}, {"Fatal", "log.Fatal", "process.exit"}, {"Fatal", "log.Fatal", "log.write"}, {"LogPanic", "log.Panic", "control.panic"}, {"LogPanic", "log.Panic", "log.write"}, {"Structured", "log/slog.Info", "log.write"}, {"Structured", "log/slog.(*Logger).LogAttrs", "log.write"},
		{"Socket", "net.(Conn).Read", "net.read"}, {"Socket", "net.(Conn).Write", "net.write"}, {"Socket", "net.(PacketConn).ReadFrom", "net.read"}, {"Socket", "net.(PacketConn).WriteTo", "net.write"}, {"Socket", "net.(*conn).Write", "net.write"}, {"Socket", "net.(*UDPConn).ReadFromUDP", "net.read"}, {"Socket", "net.(*IPConn).WriteToIP", "net.write"}, {"Socket", "net.(*UnixConn).ReadFromUnix", "net.read"},
		{"SocketTransfer", "net.(*TCPConn).ReadFrom", "net.write"}, {"SocketTransfer", "net.(*TCPConn).WriteTo", "net.read"}, {"DNS", "net.LookupHost", "net.lookup"}, {"DNS", "net.(*Resolver).LookupIPAddr", "net.lookup"},
		{"RPCDial", "net/rpc.Dial", "resource.acquire"}, {"RPCDial", "net/rpc.DialHTTPPath", "net.connect"}, {"RPC", "net/rpc.(*Client).Call", "io.write"}, {"RPC", "net/rpc.(*Client).Go", "io.read"}, {"RPC", "net/rpc.(*Client).Close", "resource.release"},
		{"TLS", "crypto/tls.(*Conn).Handshake", "tls.handshake"}, {"TLS", "crypto/tls.(*Conn).HandshakeContext", "tls.handshake"}, {"TLS", "crypto/tls.(*Conn).Read", "net.read"}, {"TLS", "crypto/tls.(*Conn).Write", "net.write"}, {"KeyFiles", "crypto/tls.LoadX509KeyPair", "fs.read"},
		{"Random", "crypto/rand.Read", "crypto.random"}, {"Random", "crypto/rand.Text", "crypto.random"}, {"TryLocks", "sync.(*Mutex).TryLock", "concurrency.lock"}, {"TryLocks", "sync.(*RWMutex).TryRLock", "concurrency.lock"},
		{"Clock", "time.Now", "clock.read"}, {"Clock", "time.Since", "clock.read"}, {"Clock", "time.Until", "clock.read"},
		{"Deadlines", "net.(Conn).SetDeadline", "io.deadline"}, {"Deadlines", "os.(*File).SetReadDeadline", "io.deadline"}, {"Deadlines", "crypto/tls.(*Conn).SetWriteDeadline", "io.deadline"}, {"Deadlines", "net/http.(*ResponseController).SetReadDeadline", "io.deadline"}, {"ConcreteDeadlines", "net.(*conn).SetDeadline", "io.deadline"},
		{"LoadPlugin", "plugin.Open", "process.load"}, {"PanicNow", "ssa.Panic", "control.panic"},
	}
	if slices.Contains([]string{"darwin", "linux", "freebsd", "openbsd", "netbsd", "dragonfly"}, runtime.GOOS) {
		cases = append(cases, want{"Credentials", "syscall.Setuid", "process.identity"}, want{"Credentials", "syscall.Setgroups", "process.identity"})
	}
	inv := inventoryByFunction(r)
	byKey := map[string]analyzer.Site{}
	for _, s := range r.Sites {
		for _, op := range s.Operations {
			byKey[s.Function+"\x00"+op.API] = s
		}
	}
	for _, w := range cases {
		t.Run(w.fn+"/"+w.api+"/"+w.effect, func(t *testing.T) {
			s, ok := byKey[w.fn+"\x00"+w.api]
			if !ok {
				t.Errorf("missing exact API %s in %s", w.api, w.fn)
				return
			}
			if !slices.Contains(s.Effects, w.effect) {
				t.Errorf("missing %s: %+v", w.effect, s)
			}
			if !slices.ContainsFunc(inv[w.fn].Effects, func(e analyzer.Effect) bool { return e.Identifier == w.effect && e.Direct }) {
				t.Errorf("direct summary evidence missing: %+v", inv[w.fn])
			}
			if filepath.Base(s.Location.File) != "operations.go" && filepath.Base(s.Location.File) != "credentials_unix.go" {
				t.Errorf("not fixture source: %+v", s)
			}
			source, err := os.ReadFile(s.Location.File)
			if err != nil {
				t.Fatal(err)
			}
			lines := strings.Split(string(source), "\n")
			if s.Location.Line < 1 || s.Location.Line > len(lines) || s.Location.Column < 1 || s.Location.Column > len(lines[s.Location.Line-1]) {
				t.Fatalf("invalid source location: %+v", s)
			}
			if w.api != "ssa.Panic" && lines[s.Location.Line-1][s.Location.Column-1] != '(' {
				t.Errorf("not source invocation parenthesis: %+v", s)
			}
		})
	}
	for _, fn := range []string{"Constructors", "Memory", "Native", "Lookalikes", "RecoverOnly", "GzipReadClose", "BasicFormatting"} {
		if len(inv[fn].Effects) != 0 || len(inv[fn].PossibleEffects) != 0 {
			t.Errorf("control %s gained effects: %+v", fn, inv[fn])
		}
	}
	for _, fn := range []string{"ReadStream", "WriteStream", "Transfer", "Buffered", "JSONStream", "XMLStream", "GobStream", "TextOutput", "HTMLOutput", "MemoryWriter"} {
		for _, e := range inv[fn].Effects {
			if strings.HasPrefix(e.Identifier, "fs.") || strings.HasPrefix(e.Identifier, "net.") {
				t.Errorf("generic/memory IO invented resource in %s: %+v", fn, inv[fn])
			}
		}
	}
	for _, check := range []struct {
		fn, api string
		effects []string
	}{
		{"CopyFile", "os.(*File).ReadFrom", []string{"fs.write", "io.read"}},
		{"CopyOut", "os.(*File).WriteTo", []string{"fs.read", "io.write"}},
		{"SocketTransfer", "net.(*TCPConn).ReadFrom", []string{"io.read", "net.write"}},
		{"SocketTransfer", "net.(*TCPConn).WriteTo", []string{"io.write", "net.read"}},
	} {
		if s := byKey[check.fn+"\x00"+check.api]; !slices.Equal(s.Effects, check.effects) {
			t.Errorf("reversed transfer direction: %+v", s)
		}
	}
	for _, fn := range []string{"GzipStream", "FlateStream", "ZlibStream", "TarStream"} {
		if slices.ContainsFunc(inv[fn].Effects, func(e analyzer.Effect) bool { return e.Identifier == "resource.release" }) {
			t.Errorf("codec release became underlying-handle release: %+v", inv[fn])
		}
	}
	if s := byKey["ConcreteDeadlines\x00net.(*conn).SetDeadline"]; !slices.Contains(s.Reasons, "state_condition") || slices.Contains(s.Reasons, "interface_dispatch") {
		t.Errorf("concrete deadline invented interface dispatch: %+v", s)
	}
	if s := byKey["DNS\x00net.(*Resolver).LookupIPAddr"]; !slices.Contains(s.Reasons, "callback") || slices.Contains(s.Reasons, "interface_dispatch") {
		t.Errorf("resolver hook confused with receiver dispatch: %+v", s)
	}
	for _, api := range []string{"crypto/rand.Read", "crypto/rand.Text"} {
		if s := byKey["Random\x00"+api]; !s.Uncertain || !slices.Contains(s.Reasons, "callback") {
			t.Errorf("mutable random Reader override lost opacity: %+v", s)
		}
	}
	if slices.ContainsFunc(inv["HeaderPrep"].Effects, func(e analyzer.Effect) bool { return e.Identifier == "http.response" }) {
		t.Error("header preparation became response transmission")
	}
	if slices.ContainsFunc(inv["Register"].Effects, func(e analyzer.Effect) bool { return e.Identifier == "net.listen" || e.Identifier == "http.response" }) {
		t.Error("registration became invocation/listen")
	}
	if slices.ContainsFunc(inv["LogPanic"].Effects, func(e analyzer.Effect) bool { return e.Identifier == "process.exit" }) {
		t.Error("panic became process exit")
	}
	if !reflect.DeepEqual(before, snapshot(t, dir)) {
		t.Error("analysis modified fixture")
	}
}

func TestOperationCoverageFlagsAndFormatting(t *testing.T) {
	dir := operationCoverageFixture(t)
	r, _ := effectResult(t, dir, "-all", "-max-depth", "1")
	inv := inventoryByFunction(r)
	identifiers := func(fn string) []string {
		var out []string
		for _, e := range inv[fn].Effects {
			out = append(out, e.Identifier)
		}
		return out
	}
	for _, fn := range []string{"RootRead", "RootBound", "PackageRead"} {
		if !slices.Equal(identifiers(fn), []string{"fs.read", "resource.acquire"}) {
			t.Errorf("readonly %s = %+v", fn, inv[fn])
		}
	}
	if !slices.Equal(identifiers("RootWrite"), []string{"fs.create", "fs.write", "resource.acquire"}) {
		t.Error(inv["RootWrite"])
	}
	if !slices.Equal(identifiers("RootDynamic"), []string{"fs.create", "fs.read", "fs.write", "resource.acquire"}) {
		t.Error(inv["RootDynamic"])
	}
	if !slices.Equal(identifiers("PackageTruncate"), []string{"fs.read", "fs.write", "resource.acquire"}) {
		t.Error(inv["PackageTruncate"])
	}
	if inv["BasicFormatting"].Uncertain {
		t.Error("basic/no-argument formatting became opaque")
	}
	for _, fn := range []string{"OpaqueFormatting", "AliasFormatting", "AfterFormattingMutation", "RepeatedFormatting", "EmptyFormatting", "LargeFormattingArray"} {
		if !inv[fn].Uncertain || len(inv[fn].Effects) != 0 {
			t.Errorf("opaque formatting %s must not infer callbacks: %+v", fn, inv[fn])
		}
	}
	u, _ := effectResult(t, dir, "-unresolved", "-function", "=OpaqueFormatting")
	if len(u.Sites) != 6 {
		t.Fatalf("formatter opacity sites=%d: %+v", len(u.Sites), u.Sites)
	}
	for _, s := range u.Sites {
		if !s.Uncertain || !slices.Contains(s.Reasons, "callback") || len(s.Effects) != 0 {
			t.Error(s)
		}
	}
	for _, s := range r.Sites {
		if s.Function == "TryLocks" && (!s.Uncertain || !slices.Contains(s.Reasons, "state_condition") || len(s.PossibleEffects) > 0) {
			t.Errorf("runtime condition is not candidate effect: %+v", s)
		}
	}
}

func TestOperationCoverageRetrievalAndPropagation(t *testing.T) {
	dir := operationCoverageFixture(t)
	base, text := effectResult(t, dir, "-all", "-max-depth", "2")
	_, again := effectResult(t, dir, "-all", "-max-depth", "2")
	if text != again {
		t.Fatal("nondeterministic operation report")
	}
	for _, effect := range []string{"io.read", "io.write", "log.write", "net.read", "net.write", "net.lookup", "tls.handshake", "crypto.random", "fs.sync", "io.deadline", "clock.read", "process.load", "http.handler.register", "http.header.write", "control.panic"} {
		t.Run(effect, func(t *testing.T) {
			r, _ := effectResult(t, dir, "-effect", effect, "-max-depth", "0")
			if len(r.Sites) == 0 {
				t.Fatalf("effect not retrievable: %s", effect)
			}
			for _, s := range r.Sites {
				if !slices.Contains(s.Effects, effect) && !slices.Contains(s.PossibleEffects, effect) {
					t.Error(s)
				}
			}
		})
	}
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		r, _ := effectResult(t, dir, "-effect", "process.identity", "-max-depth", "0")
		want := 5
		if runtime.GOOS == "linux" {
			want = 7
		}
		if len(r.Sites) != want {
			t.Fatal(r.Sites)
		}
	}
	known, _ := effectResult(t, dir, "-api", "log.Print", "-function", "=Logging")
	if len(known.Sites) != 1 || !known.Sites[0].Uncertain {
		t.Fatal(known.Sites)
	}
	site := known.Sites[0]
	for _, args := range [][]string{{"-site", site.ID}, {"-api", "log.Print", "-function", "=Logging"}, {"-operation", "log.Print", "-function", "=Logging"}, {"-effect", "log.write", "-function", "=Logging", "-reason", "callback"}} {
		r, _ := effectResult(t, dir, args...)
		found := false
		for _, s := range r.Sites {
			found = found || s.ID == site.ID
		}
		if !found {
			t.Fatalf("lost source ID with %v", args)
		}
	}
	for _, args := range [][]string{{"-unresolved", "-api", "log.Print", "-function", "=Logging"}, {"-api", "log.Print", "-effect", "fs.write"}, {"-api", "Print"}, {"-api", "crypto/x509.(*Certificate).Verify"}} {
		r, _ := effectResult(t, dir, args...)
		if len(r.Sites) > 0 {
			t.Errorf("negative filter %v: %+v", args, r.Sites)
		}
	}
	unresolved, _ := effectResult(t, dir, "-unresolved", "-api", "crypto/x509.(*Certificate).Verify")
	if len(unresolved.Sites) != 1 || len(unresolved.Sites[0].Effects) != 0 {
		t.Fatal("known identity is not crypto classification", unresolved.Sites)
	}
	inv := inventoryByFunction(base)
	if !slices.ContainsFunc(inv["WriteCaller"].Effects, func(e analyzer.Effect) bool { return e.Identifier == "io.write" && !e.Direct }) {
		t.Error("missing inherited stream operation", inv["WriteCaller"])
	}
	for _, kind := range []struct{ fn, kind string }{{"DeferredWrite", "defer"}, {"PanicTask", "go"}} {
		found := false
		for _, e := range base.Edges {
			if e.Source != nil && e.Source.Caller.Function == kind.fn && e.Source.Kind == kind.kind {
				found = true
			}
		}
		if !found {
			t.Errorf("lost %s source edge from %s", kind.kind, kind.fn)
		}
	}
	a := inv["Alternatives"]
	if len(a.Effects) > 0 || !slices.Equal(a.PossibleEffects, []string{"log.write"}) || !a.Uncertain {
		t.Error("finite alternatives promoted or opacity lost", a)
	}
	pair, err, status := invoke(t, dir, nil, "-view", "functions", "-function", "=Alternatives")
	if status != 0 || err != "" || pair != "" {
		t.Fatalf("pair compatibility: %d %q %q", status, pair, err)
	}
	out, err, status := invoke(t, dir, nil, "-function", "=Alternatives")
	if status != 0 || err != "" || !strings.Contains(out, "candidate-only") || !strings.Contains(out, "log.write") {
		t.Fatalf("text lost candidate distinction: %d %s %s", status, out, err)
	}
	r, _ := effectResult(t, dir, "-api", "io.(Writer).Write", "-function", "=WriteCaller", "-max-depth", "1")
	if len(r.Sites) != 1 || r.Sites[0].Function != "WriteStream" || len(r.Edges) != 1 || r.Edges[0].Source == nil || r.Edges[0].Source.Resolution != "established" {
		t.Fatalf("reverse source caller: %+v", r)
	}
	cf, diagnostic, status := invoke(t, dir, nil, "-control-flow", "-function", "=PanicNow")
	if status != 0 || diagnostic != "" {
		t.Fatal(status, diagnostic)
	}
	var flow analyzer.ControlFlow
	if err := json.Unmarshal([]byte(cf), &flow); err != nil {
		t.Fatal(err)
	}
	if len(flow.Rows) != 1 || flow.Rows[0].API != "ssa.Panic" || flow.Rows[0].Status != "complete" {
		t.Fatalf("panic CFG mapping: %+v", flow)
	}
}
