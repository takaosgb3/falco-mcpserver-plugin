package main

import (
    "crypto/tls"
    "flag"
    "fmt"
    "io"
    "log"
    "net"
    "net/url"
    "os"
    "strconv"
    "strings"
    "sync/atomic"
    "time"

    "github.com/takaosgb3/falco-mcpserver-plugin/pkg/audit"
)

func main() {
    listen := flag.String("listen", ":8989", "listen address (host:port)")
    target := flag.String("target", "", "target MCP endpoint (host:port or ws(s)://host:port/path)")
    sink := flag.String("sink", "stdout", "audit sink file path or 'stdout'")
    client := flag.String("client-process", "unknown", "client process name (e.g., claude-code, codex-cli)")
    session := flag.String("session-id", "", "session id (optional)")
    assumeTLS := flag.Bool("assume-tls", false, "mark TLS=true in audit (useful for wss)")
    flag.Parse()

    if *target == "" {
        log.Fatalf("missing --target")
    }

    host, port, tlsHint, err := parseTarget(*target)
    if err != nil {
        log.Fatalf("parse target: %v", err)
    }
    if *assumeTLS {
        tlsHint = true
    }

    w, err := audit.NewWriter(*sink)
    if err != nil {
        log.Fatalf("open sink: %v", err)
    }
    defer w.Close()

    ln, err := net.Listen("tcp", *listen)
    if err != nil {
        log.Fatalf("listen: %v", err)
    }
    log.Printf("mcp-audit-proxy listening on %s -> %s:%d (tls=%v)", *listen, host, port, tlsHint)
    for {
        c, err := ln.Accept()
        if err != nil {
            log.Printf("accept: %v", err)
            continue
        }
        go handleConn(c, host, port, tlsHint, *client, *session, w)
    }
}

func handleConn(in net.Conn, host string, port int, tlsMark bool, client string, sessionID string, w *audit.Writer) {
    defer in.Close()
    out, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), 10*time.Second)
    if err != nil {
        log.Printf("dial target: %v", err)
        return
    }
    defer out.Close()

    var up, down uint64
    done := make(chan struct{}, 2)
    go func() {
        n, _ := io.Copy(out, in)
        atomic.AddUint64(&up, uint64(n))
        _ = out.(*net.TCPConn).CloseWrite()
        done <- struct{}{}
    }()
    go func() {
        n, _ := io.Copy(in, out)
        atomic.AddUint64(&down, uint64(n))
        _ = in.(*net.TCPConn).CloseWrite()
        done <- struct{}{}
    }()
    <-done
    <-done

    _ = w.Emit(audit.Event{
        SessionID:       nonEmpty(sessionID, fmt.Sprintf("sess-%d", time.Now().UnixNano())),
        ClientProcess:   client,
        ServerHost:      host,
        ServerPort:      port,
        TLS:             tlsMark,
        Method:          "",
        RequestBytes:    atomic.LoadUint64(&up),
        ResponseBytes:   atomic.LoadUint64(&down),
        ToolInvokeCount: 0,
        FileAccessCount: 0,
    })
}

func parseTarget(t string) (host string, port int, tlsHint bool, err error) {
    if strings.HasPrefix(t, "ws://") || strings.HasPrefix(t, "wss://") || strings.HasPrefix(t, "http://") || strings.HasPrefix(t, "https://") {
        u, e := url.Parse(t)
        if e != nil {
            return "", 0, false, e
        }
        host = u.Hostname()
        p := u.Port()
        if p == "" {
            if u.Scheme == "wss" || u.Scheme == "https" {
                p = "443"
            } else {
                p = "80"
            }
        }
        pi, e := strconv.Atoi(p)
        if e != nil {
            return "", 0, false, e
        }
        return host, pi, u.Scheme == "wss" || u.Scheme == "https", nil
    }
    // host:port
    parts := strings.Split(t, ":")
    if len(parts) != 2 {
        return "", 0, false, fmt.Errorf("invalid target: %s", t)
    }
    pi, e := strconv.Atoi(parts[1])
    if e != nil {
        return "", 0, false, e
    }
    // TLS hint: common ports
    tlsGuess := (pi == 443)
    return parts[0], pi, tlsGuess, nil
}

func nonEmpty(a, b string) string {
    if a != "" {
        return a
    }
    return b
}

// avoid unused import error when not using tls directly but keeping for future
var _ = tls.VersionTLS13

