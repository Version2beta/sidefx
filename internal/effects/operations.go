package effects

// addOperationSummaries lists exact operation contracts, not destinations,
// successful transitions, ownership proofs or execution of opaque callbacks.
// Public interface declaring identities live in LookupInterface.
func addOperationSummaries(add func(string, Set, ...string)) {
	add("os", FSWrite|IORead|Unknown, "(*File).ReadFrom")
	add("os", FSRead|IOWrite|Unknown, "(*File).WriteTo")
	add("os", FSRead|FSCreate|FSWrite|Unknown, "CopyFS")
	add("os", FSWrite|FSSync, "(*File).Sync") // Retain the legacy category.
	add("path/filepath", FSRead|Unknown, "Walk", "WalkDir")
	// These name filesystem-namespace operations, including virtual/memory FS.
	add("io/fs", FSRead|Unknown, "ReadFile", "ReadDir", "Stat", "Glob", "WalkDir", "ReadLink")
	add("archive/zip", FSRead|ResourceAcquire, "OpenReader")
	add("archive/zip", ResourceRelease, "(*ReadCloser).Close")

	add("io", IORead|Unknown, "ReadAll", "ReadFull", "ReadAtLeast")
	add("io", IOWrite|Unknown, "WriteString")
	add("io", IORead|IOWrite|Unknown, "Copy", "CopyN", "CopyBuffer")
	add("io", 0, "LimitReader", "TeeReader", "MultiReader", "MultiWriter", "NopCloser", "NewSectionReader", "NewOffsetWriter")
	add("bufio", 0, "NewReader", "NewReaderSize", "NewWriter", "NewWriterSize", "NewReadWriter", "NewScanner")
	add("bufio", IORead|Unknown, "(*Reader).Read", "(*Reader).ReadByte", "(*Reader).ReadRune", "(*Reader).ReadString", "(*Reader).ReadBytes", "(*Reader).ReadSlice", "(*Reader).Peek", "(*Reader).Discard", "(*Scanner).Scan")
	add("bufio", IOWrite|Unknown, "(*Writer).Write", "(*Writer).WriteByte", "(*Writer).WriteRune", "(*Writer).WriteString", "(*Writer).Flush")
	add("bufio", IORead|IOWrite|Unknown, "(*Reader).WriteTo", "(*Writer).ReadFrom")
	for _, pkg := range []string{"encoding/json", "encoding/xml", "encoding/gob"} {
		add(pkg, 0, "NewDecoder", "NewEncoder")
		add(pkg, IORead|Unknown, "(*Decoder).Decode")
		add(pkg, IOWrite|Unknown, "(*Encoder).Encode")
	}
	add("encoding/json", IORead|Unknown, "(*Decoder).Token", "(*Decoder).More")
	add("encoding/xml", IORead|Unknown, "(*Decoder).Token", "(*Decoder).RawToken", "(*Decoder).DecodeElement", "(*Decoder).Skip")
	add("encoding/xml", IOWrite|Unknown, "(*Encoder).EncodeToken", "(*Encoder).EncodeElement", "(*Encoder).Flush", "(*Encoder).Close")
	add("encoding/gob", IORead|Unknown, "(*Decoder).DecodeValue")
	add("encoding/gob", IOWrite|Unknown, "(*Encoder).EncodeValue")
	for _, pkg := range []string{"text/template", "html/template"} {
		add(pkg, IOWrite|Unknown, "(*Template).Execute", "(*Template).ExecuteTemplate")
	}
	// Reader constructors below can consume headers. Writer construction alone
	// does not publish bytes. Closing codecs need not close underlying handles.
	add("compress/gzip", IORead|Unknown, "NewReader", "(*Reader).Read")
	add("compress/gzip", 0, "NewWriter", "NewWriterLevel", "(*Reader).Close")
	add("compress/gzip", IOWrite|Unknown, "(*Writer).Write", "(*Writer).Flush", "(*Writer).Close")
	add("compress/flate", 0, "NewReader", "NewReaderDict", "NewWriter", "NewWriterDict")
	add("compress/flate", IOWrite|Unknown, "(*Writer).Write", "(*Writer).Flush", "(*Writer).Close")
	add("compress/zlib", IORead|Unknown, "NewReader", "NewReaderDict")
	add("compress/zlib", 0, "NewWriter", "NewWriterLevel", "NewWriterLevelDict")
	add("compress/zlib", IOWrite|Unknown, "(*Writer).Write", "(*Writer).Flush", "(*Writer).Close")
	add("archive/tar", 0, "NewReader", "NewWriter")
	add("archive/tar", IORead|Unknown, "(*Reader).Next", "(*Reader).Read")
	add("archive/tar", IOWrite|Unknown, "(*Writer).Write", "(*Writer).WriteHeader", "(*Writer).Flush", "(*Writer).Close")
	add("fmt", IOWrite|Unknown, "Print", "Printf", "Println", "Fprint", "Fprintf", "Fprintln")

	add("net/http", HTTPResponse|Unknown, "Error", "NotFound", "Redirect", "ServeContent")
	add("net/http", FSRead|HTTPResponse|Unknown, "ServeFile", "ServeFileFS")
	add("net/http", HTTPHeaderWrite, "(Header).Add", "(Header).Set", "(Header).Del")
	add("net/http", HTTPHeaderWrite|Unknown, "SetCookie")
	add("net/http", IOWrite|Unknown, "(Header).Write", "(Header).WriteSubset")
	add("net/http", HTTPHandlerRegister, "Handle", "HandleFunc", "(*ServeMux).Handle", "(*ServeMux).HandleFunc")
	add("net/http/httputil", NetConnect|HTTPResponse|Unknown, "(*ReverseProxy).ServeHTTP")
	add("net/http", 0, "NewServeMux", "RedirectHandler", "NotFoundHandler", "StripPrefix", "FileServer", "FileServerFS", "ProxyURL")
	add("net/http/httputil", 0, "NewSingleHostReverseProxy")

	add("log", LogWrite|Unknown, "Print", "Printf", "Println", "Output", "(*Logger).Print", "(*Logger).Printf", "(*Logger).Println", "(*Logger).Output")
	add("log", LogWrite|ProcessExit|Unknown, "Fatal", "Fatalf", "Fatalln", "(*Logger).Fatal", "(*Logger).Fatalf", "(*Logger).Fatalln")
	add("log", LogWrite|ControlPanic|Unknown, "Panic", "Panicf", "Panicln", "(*Logger).Panic", "(*Logger).Panicf", "(*Logger).Panicln")
	for _, method := range []string{"Log", "LogAttrs", "Debug", "DebugContext", "Info", "InfoContext", "Warn", "WarnContext", "Error", "ErrorContext"} {
		add("log/slog", LogWrite|Unknown, method, "(*Logger)."+method)
	}

	// Promoted net methods retain their declaring (*conn) identity in go/types.
	add("net", NetRead, "(*conn).Read")
	add("net", NetWrite, "(*conn).Write")
	for _, recv := range []string{"(*UDPConn)", "(*IPConn)", "(*UnixConn)"} {
		add("net", NetRead, recv+".ReadFrom")
		add("net", NetWrite, recv+".WriteTo")
	}
	// TCP ReadFrom reads a generic peer into the socket; WriteTo is inverse.
	add("net", NetWrite|IORead|Unknown, "(*TCPConn).ReadFrom")
	add("net", NetRead|IOWrite|Unknown, "(*TCPConn).WriteTo")
	add("net", NetRead, "(*UDPConn).ReadFromUDP", "(*UDPConn).ReadFromUDPAddrPort", "(*UDPConn).ReadMsgUDP", "(*UDPConn).ReadMsgUDPAddrPort", "(*IPConn).ReadFromIP", "(*IPConn).ReadMsgIP", "(*UnixConn).ReadFromUnix", "(*UnixConn).ReadMsgUnix")
	add("net", NetWrite, "(*UDPConn).WriteToUDP", "(*UDPConn).WriteToUDPAddrPort", "(*UDPConn).WriteMsgUDP", "(*UDPConn).WriteMsgUDPAddrPort", "(*IPConn).WriteToIP", "(*IPConn).WriteMsgIP", "(*UnixConn).WriteToUnix", "(*UnixConn).WriteMsgUnix")
	for _, method := range []string{"LookupHost", "LookupIP", "LookupAddr", "LookupCNAME", "LookupPort", "LookupTXT", "LookupMX", "LookupNS", "LookupSRV"} {
		add("net", NetLookup|Unknown, method, "(*Resolver)."+method)
	}
	add("net", NetLookup|Unknown, "(*Resolver).LookupIPAddr", "(*Resolver).LookupNetIP")
	add("net/rpc", NetConnect|ResourceAcquire|Unknown, "Dial", "DialHTTP", "DialHTTPPath")
	add("net/rpc", IORead|IOWrite|Unknown, "(*Client).Call", "(*Client).Go")
	add("net/rpc", ResourceRelease|Unknown, "(*Client).Close")

	add("crypto/tls", FSRead, "LoadX509KeyPair")
	add("crypto/tls", TLSHandshake|Unknown, "(*Conn).Handshake", "(*Conn).HandshakeContext")
	add("crypto/tls", NetRead|Unknown, "(*Conn).Read")
	add("crypto/tls", NetWrite|Unknown, "(*Conn).Write")
	// Reader is mutable; Read (and Text through Read) can invoke an override.
	add("crypto/rand", CryptoRandom|Unknown, "Read", "Text")
	add("sync", ConcurrencyLock|Unknown, "(*Mutex).TryLock", "(*RWMutex).TryLock", "(*RWMutex).TryRLock")
	add("time", ClockRead, "Now", "Since", "Until")
	add("context", ConcurrencyTimerSchedule|Unknown, "WithDeadline", "WithDeadlineCause", "WithTimeout", "WithTimeoutCause")
	for _, method := range []string{"SetDeadline", "SetReadDeadline", "SetWriteDeadline"} {
		add("os", IODeadline|Unknown, "(*File)."+method)
		add("crypto/tls", IODeadline|Unknown, "(*Conn)."+method)
		add("net", IODeadline|Unknown, "(*conn)."+method)
	}
	add("net", IODeadline|Unknown, "(*TCPListener).SetDeadline", "(*UnixListener).SetDeadline")
	add("net/http", IODeadline|Unknown, "(*ResponseController).SetReadDeadline", "(*ResponseController).SetWriteDeadline")
	add("plugin", ProcessLoad|Unknown, "Open")
	// APIs absent on a target platform cannot resolve to these exact keys.
	add("syscall", ProcessIdentity|Unknown, "Setuid", "Setgid", "Setreuid", "Setregid", "Setresuid", "Setresgid", "Setgroups")
}
