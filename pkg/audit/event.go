package audit

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "sync"
    "time"
)

type Event struct {
    SchemaVersion   string `json:"schema_version"`
    Timestamp       int64  `json:"timestamp"`
    SessionID       string `json:"session_id"`
    ClientProcess   string `json:"client_process"`
    ServerHost      string `json:"server_host"`
    ServerPort      int    `json:"server_port"`
    TLS             bool   `json:"tls"`
    AuthScheme      string `json:"auth_scheme,omitempty"`
    Method          string `json:"method"`
    RequestBytes    uint64 `json:"request_bytes"`
    ResponseBytes   uint64 `json:"response_bytes"`
    ToolInvokeCount uint32 `json:"tool_invoke_count"`
    FileAccessCount uint32 `json:"file_access_count"`
}

type Writer struct {
    mu   sync.Mutex
    sink io.WriteCloser
}

func NewWriter(path string) (*Writer, error) {
    if path == "" || path == "-" || path == "stdout" {
        return &Writer{sink: nopCloser{os.Stdout}}, nil
    }
    f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
    if err != nil {
        return nil, err
    }
    return &Writer{sink: f}, nil
}

func (w *Writer) Close() error {
    w.mu.Lock()
    defer w.mu.Unlock()
    if w.sink != nil {
        return w.sink.Close()
    }
    return nil
}

func (w *Writer) Emit(ev Event) error {
    w.mu.Lock()
    defer w.mu.Unlock()
    if ev.SchemaVersion == "" {
        ev.SchemaVersion = "1"
    }
    if ev.Timestamp == 0 {
        ev.Timestamp = time.Now().UnixNano()
    }
    b, err := json.Marshal(ev)
    if err != nil {
        return fmt.Errorf("marshal event: %w", err)
    }
    if _, err := w.sink.Write(append(b, '\n')); err != nil {
        return fmt.Errorf("write event: %w", err)
    }
    return nil
}

type nopCloser struct{ io.Writer }

func (n nopCloser) Close() error { return nil }

