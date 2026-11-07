// Package operationcoverage is analyzed statically, never executed.
package operationcoverage

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"fmt"
	html "html/template"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/rpc"
	"net/url"
	"os"
	"path/filepath"
	"plugin"
	"sync"
	text "text/template"
	"time"
)

func RootRead(r *os.Root)               { _, _ = r.OpenFile("never", os.O_RDONLY, 0) }
func RootWrite(r *os.Root)              { _, _ = r.OpenFile("never", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600) }
func RootDynamic(r *os.Root, flags int) { _, _ = r.OpenFile("never", flags, 0600) }
func RootBound(r *os.Root)              { fn := r.OpenFile; _, _ = fn("never", os.O_RDONLY, 0) }
func PackageRead()                      { _, _ = os.OpenFile("never", os.O_RDONLY, 0) }
func PackageTruncate()                  { _, _ = os.OpenFile("never", os.O_TRUNC, 0) }
func CopyFile(f *os.File, r io.Reader)  { _, _ = f.ReadFrom(r) }
func CopyOut(f *os.File, w io.Writer)   { _, _ = f.WriteTo(w) }
func CopyTree(f fs.FS)                  { _ = os.CopyFS("never", f) }
func WalkTree()                         { _ = filepath.Walk("never", func(string, fs.FileInfo, error) error { return nil }) }
func WalkDirTree() {
	_ = filepath.WalkDir("never", func(string, fs.DirEntry, error) error { return nil })
}
func FSFacade(f fs.FS) {
	_, _ = fs.ReadFile(f, "never")
	_, _ = fs.ReadDir(f, "never")
	_, _ = fs.Stat(f, "never")
	_, _ = fs.Glob(f, "*")
	_ = fs.WalkDir(f, ".", func(string, fs.DirEntry, error) error { return nil })
	_, _ = fs.ReadLink(f, "never")
}
func FSOpen(f fs.FS) { _, _ = f.Open("never") }
func FSInterfaces(r fs.ReadFileFS, d fs.ReadDirFS, s fs.StatFS, l fs.ReadLinkFS, f fs.File, dir fs.ReadDirFile) {
	_, _ = r.ReadFile("never")
	_, _ = d.ReadDir("never")
	_, _ = s.Stat("never")
	_, _ = l.ReadLink("never")
	_, _ = f.Read(nil)
	_, _ = f.Stat()
	_ = f.Close()
	_, _ = dir.ReadDir(0)
}
func ZipOpen()                   { _, _ = zip.OpenReader("never") }
func ZipClose(z *zip.ReadCloser) { _ = z.Close() }
func FileSync(f *os.File)        { _ = f.Sync() }

func ReadStream(r io.Reader)  { _, _ = r.Read(nil) }
func WriteStream(w io.Writer) { _, _ = w.Write(nil) }
func IOInterfaces(r io.ReaderAt, b io.ByteReader, u io.RuneReader, w io.WriterAt, s io.StringWriter, bw io.ByteWriter, rf io.ReaderFrom, wt io.WriterTo) {
	_, _ = r.ReadAt(nil, 0)
	_, _ = b.ReadByte()
	_, _, _ = u.ReadRune()
	_, _ = w.WriteAt(nil, 0)
	_, _ = s.WriteString("")
	_ = bw.WriteByte(0)
	_, _ = rf.ReadFrom(nil)
	_, _ = wt.WriteTo(nil)
}
func Transfer(w io.Writer, r io.Reader) {
	_, _ = io.Copy(w, r)
	_, _ = io.CopyN(w, r, 1)
	_, _ = io.CopyBuffer(w, r, nil)
	_, _ = io.ReadAll(r)
	_, _ = io.ReadFull(r, nil)
	_, _ = io.ReadAtLeast(r, nil, 1)
	_, _ = io.WriteString(w, "never")
}
func Buffered(r *bufio.Reader, w *bufio.Writer, s *bufio.Scanner) {
	_, _ = r.Read(nil)
	_, _ = r.ReadByte()
	_, _, _ = r.ReadRune()
	_, _ = r.ReadString('\n')
	_, _ = r.ReadBytes('\n')
	_, _ = r.ReadSlice('\n')
	_, _ = r.Peek(1)
	_, _ = r.Discard(1)
	_, _ = r.WriteTo(w)
	_, _ = w.Write(nil)
	_ = w.WriteByte(0)
	_, _ = w.WriteRune('x')
	_, _ = w.WriteString("never")
	_, _ = w.ReadFrom(r)
	_ = w.Flush()
	_ = s.Scan()
}
func JSONStream(r *json.Decoder, w *json.Encoder, v any) {
	_ = r.Decode(v)
	_, _ = r.Token()
	_ = r.More()
	_ = w.Encode(v)
}
func XMLStream(r *xml.Decoder, w *xml.Encoder, v any) {
	_ = r.Decode(v)
	_, _ = r.Token()
	_, _ = r.RawToken()
	_ = r.Skip()
	_ = w.Encode(v)
	_ = w.EncodeToken(xml.StartElement{})
	_ = w.Flush()
	_ = w.Close()
}
func GobStream(r *gob.Decoder, w *gob.Encoder, v any) { _ = r.Decode(v); _ = w.Encode(v) }
func TextOutput(t *text.Template, w io.Writer) {
	_ = t.Execute(w, nil)
	_ = t.ExecuteTemplate(w, "never", nil)
}
func HTMLOutput(t *html.Template, w io.Writer) {
	_ = t.Execute(w, nil)
	_ = t.ExecuteTemplate(w, "never", nil)
}
func GzipStream(r *gzip.Reader, w *gzip.Writer) {
	_, _ = r.Read(nil)
	_, _ = w.Write(nil)
	_ = w.Flush()
	_ = w.Close()
}
func GzipHeader(r io.Reader)       { _, _ = gzip.NewReader(r) }
func GzipReadClose(r *gzip.Reader) { _ = r.Close() }
func FlateStream(w *flate.Writer)  { _, _ = w.Write(nil); _ = w.Flush(); _ = w.Close() }
func ZlibStream(w *zlib.Writer)    { _, _ = w.Write(nil); _ = w.Flush(); _ = w.Close() }
func ZlibHeader(r io.Reader)       { _, _ = zlib.NewReader(r) }
func TarStream(r *tar.Reader, w *tar.Writer) {
	_, _ = r.Next()
	_, _ = r.Read(nil)
	_, _ = w.Write(nil)
	_ = w.WriteHeader(&tar.Header{})
	_ = w.Flush()
	_ = w.Close()
}
func PrintOutput(w io.Writer) {
	_, _ = fmt.Print("never")
	_, _ = fmt.Printf("%s", "never")
	_, _ = fmt.Println("never")
	_, _ = fmt.Fprint(w, "never")
	_, _ = fmt.Fprintf(w, "%s", "never")
	_, _ = fmt.Fprintln(w, "never")
	print("never")
	println("never")
}

func HTTPHelpers(w http.ResponseWriter, r *http.Request, f fs.FS, c io.ReadSeeker) {
	http.Error(w, "never", 500)
	http.NotFound(w, r)
	http.Redirect(w, r, "never", 302)
	http.ServeContent(w, r, "never", time.Time{}, c)
	http.ServeFile(w, r, "never")
	http.ServeFileFS(w, r, f, "never")
}
func HeaderPrep(w http.ResponseWriter, h http.Header) {
	h.Add("never", "never")
	h.Set("never", "never")
	h.Del("never")
	http.SetCookie(w, &http.Cookie{Name: "never"})
}
func HeaderOutput(h http.Header, w io.Writer) { _ = h.Write(w); _ = h.WriteSubset(w, nil) }
func Register(m *http.ServeMux, h http.Handler) {
	http.Handle("/never", h)
	http.HandleFunc("/never", func(http.ResponseWriter, *http.Request) {})
	m.Handle("/never", h)
	m.HandleFunc("/never", func(http.ResponseWriter, *http.Request) {})
}
func Proxy(p *httputil.ReverseProxy, w http.ResponseWriter, r *http.Request) { p.ServeHTTP(w, r) }
func Logging(l *log.Logger) {
	log.Print("never")
	log.Printf("%s", "never")
	log.Println("never")
	_ = log.Output(1, "never")
	l.Print("never")
	l.Printf("%s", "never")
	l.Println("never")
	_ = l.Output(1, "never")
}
func Fatal(l *log.Logger)    { log.Fatal("never"); l.Fatalf("%s", "never") }
func LogPanic(l *log.Logger) { log.Panic("never"); l.Panicln("never") }
func Structured(l *slog.Logger, ctx context.Context) {
	slog.Info("never")
	slog.DebugContext(ctx, "never")
	slog.Log(ctx, slog.LevelInfo, "never")
	slog.LogAttrs(ctx, slog.LevelInfo, "never")
	l.Error("never")
	l.WarnContext(ctx, "never")
	l.Log(ctx, slog.LevelInfo, "never")
	l.LogAttrs(ctx, slog.LevelInfo, "never")
}

func Socket(c net.Conn, p net.PacketConn, t *net.TCPConn, u *net.UDPConn, i *net.IPConn, x *net.UnixConn) {
	_, _ = c.Read(nil)
	_, _ = c.Write(nil)
	_, _, _ = p.ReadFrom(nil)
	_, _ = p.WriteTo(nil, nil)
	_, _ = t.Read(nil)
	_, _ = t.Write(nil)
	_, _, _ = u.ReadFromUDP(nil)
	_, _ = u.WriteToUDP(nil, nil)
	_, _, _ = i.ReadFromIP(nil)
	_, _ = i.WriteToIP(nil, nil)
	_, _, _ = x.ReadFromUnix(nil)
	_, _ = x.WriteToUnix(nil, nil)
}
func SocketTransfer(c *net.TCPConn, r io.Reader, w io.Writer) {
	_, _ = c.ReadFrom(r)
	_, _ = c.WriteTo(w)
}
func DNS(r *net.Resolver, ctx context.Context) {
	_, _ = net.LookupHost("never.invalid")
	_, _ = net.LookupIP("never.invalid")
	_, _ = net.LookupAddr("127.0.0.1")
	_, _ = net.LookupCNAME("never.invalid")
	_, _ = net.LookupPort("tcp", "never")
	_, _ = net.LookupTXT("never.invalid")
	_, _ = net.LookupMX("never.invalid")
	_, _ = net.LookupNS("never.invalid")
	_, _, _ = net.LookupSRV("never", "tcp", "never.invalid")
	_, _ = r.LookupHost(ctx, "never.invalid")
	_, _ = r.LookupIPAddr(ctx, "never.invalid")
	_, _ = r.LookupNetIP(ctx, "ip", "never.invalid")
}
func RPCDial() {
	_, _ = rpc.Dial("tcp", "never.invalid")
	_, _ = rpc.DialHTTP("tcp", "never.invalid")
	_, _ = rpc.DialHTTPPath("tcp", "never.invalid", "/never")
}
func RPC(c *rpc.Client) {
	_ = c.Call("never", nil, nil)
	_ = c.Go("never", nil, nil, nil)
	_ = c.Close()
}
func TLS(c *tls.Conn, ctx context.Context) {
	_ = c.Handshake()
	_ = c.HandshakeContext(ctx)
	_, _ = c.Read(nil)
	_, _ = c.Write(nil)
}
func KeyFiles()                               { _, _ = tls.LoadX509KeyPair("never", "never") }
func Random()                                 { _, _ = rand.Read(nil); _ = rand.Text() }
func TryLocks(m *sync.Mutex, r *sync.RWMutex) { _ = m.TryLock(); _ = r.TryLock(); _ = r.TryRLock() }
func Clock(t time.Time)                       { _ = time.Now(); _ = time.Since(t); _ = time.Until(t) }
func Deadlines(c net.Conn, p net.PacketConn, f *os.File, t *tls.Conn, h *http.ResponseController) {
	z := time.Time{}
	_ = c.SetDeadline(z)
	_ = c.SetReadDeadline(z)
	_ = c.SetWriteDeadline(z)
	_ = p.SetDeadline(z)
	_ = f.SetDeadline(z)
	_ = f.SetReadDeadline(z)
	_ = f.SetWriteDeadline(z)
	_ = t.SetDeadline(z)
	_ = t.SetReadDeadline(z)
	_ = t.SetWriteDeadline(z)
	_ = h.SetReadDeadline(z)
	_ = h.SetWriteDeadline(z)
}
func ConcreteDeadlines(c *net.TCPConn, l *net.TCPListener) {
	_ = c.SetDeadline(time.Time{})
	_ = l.SetDeadline(time.Time{})
}
func LoadPlugin()               { _, _ = plugin.Open("never") }
func PanicNow()                 { panic("never") }
func RecoverOnly() any          { return recover() }
func PanicTask()                { go PanicNow() }
func WriteCaller(w io.Writer)   { WriteStream(w) }
func DeferredWrite(w io.Writer) { defer WriteStream(w) }
func Alternatives(b bool) {
	fn := log.Print
	if b {
		fn = log.Println
	}
	fn("never")
}

// Pure constructors/values are controls, not promises of complete purity.
func Constructors(r io.Reader, w io.Writer, u *url.URL, h http.Handler) {
	_ = io.LimitReader(r, 1)
	_ = io.TeeReader(r, w)
	_ = io.MultiReader(r)
	_ = io.MultiWriter(w)
	_ = io.NopCloser(r)
	_ = bufio.NewReader(r)
	_ = bufio.NewWriter(w)
	_ = bufio.NewScanner(r)
	_ = json.NewDecoder(r)
	_ = json.NewEncoder(w)
	_ = xml.NewDecoder(r)
	_ = xml.NewEncoder(w)
	_ = gob.NewDecoder(r)
	_ = gob.NewEncoder(w)
	_ = tar.NewReader(r)
	_ = tar.NewWriter(w)
	_ = gzip.NewWriter(w)
	_ = zlib.NewWriter(w)
	_ = http.NewServeMux()
	_ = http.RedirectHandler("never", 302)
	_ = http.StripPrefix("/never", h)
	_ = http.ProxyURL(u)
	_ = httputil.NewSingleHostReverseProxy(u)
}
func Memory(v any, t time.Time) {
	_, _ = json.Marshal(v)
	_ = json.Unmarshal(nil, v)
	_ = t.Add(1)
	_ = t.Before(t)
	_ = filepath.Join("a", "b")
}
func Verify(c *x509.Certificate) { _, _ = c.Verify(x509.VerifyOptions{}) }
func Native(m map[string]int, n int) []byte {
	m["x"]++
	delete(m, "y")
	clear(m)
	return make([]byte, n)
}
func MemoryWriter() { var b bytes.Buffer; var w io.Writer = &b; _, _ = w.Write(nil) }

type fake struct{}

func (fake) Read([]byte) (int, error)                            { return 0, nil }
func (fake) Write([]byte) (int, error)                           { return 0, nil }
func (fake) Handshake() error                                    { return nil }
func (fake) TryLock() bool                                       { return true }
func (fake) OpenFile(string, int, os.FileMode) (*os.File, error) { return nil, nil }
func Lookalikes() {
	f := fake{}
	_, _ = f.Read(nil)
	_, _ = f.Write(nil)
	_ = f.Handshake()
	_ = f.TryLock()
	_, _ = f.OpenFile("never", 0, 0)
}

type stringer int

func (stringer) String() string { _, _ = os.ReadFile("never"); return "never" }
func BasicFormatting(n int) {
	_ = fmt.Sprint(n)
	_ = fmt.Sprintln(n)
	_ = fmt.Sprintf("%d", n)
	_ = fmt.Append(nil, n)
	_ = fmt.Appendf(nil, "%d", n)
	_ = fmt.Appendln(nil, n)
	_ = fmt.Sprint()
	_ = fmt.Append(nil)
}
func OpaqueFormatting() {
	_ = fmt.Sprint(stringer(0))
	_ = fmt.Sprintln(stringer(0))
	_ = fmt.Append(nil, stringer(0))
	_ = fmt.Appendf(nil, "%s", stringer(0))
	_ = fmt.Appendln(nil, stringer(0))
	_ = fmt.Sprintf("%s", stringer(0))
}
func mutate(a []any)           { a[0] = stringer(0) }
func AliasFormatting()         { a := []any{1}; mutate(a); _ = fmt.Sprint(a...) }
func AfterFormattingMutation() { a := []any{1}; _ = fmt.Sprint(a...); a[0] = stringer(0) }
func RepeatedFormatting()      { a := []any{1}; _ = fmt.Sprint(a...); _ = fmt.Sprint(a...) }
func EmptyFormatting(a []any)  { _ = fmt.Sprint(a[:0]...) }
