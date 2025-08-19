package main

import (
    "bufio"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "math/rand"
    "os"
    "strings"
    "time"
)

type rpcReq struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      json.RawMessage `json:"id"`
    Method  string          `json:"method"`
    Params  json.RawMessage `json:"params"`
}

type rpcRes struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      json.RawMessage `json:"id"`
    Result  any             `json:"result,omitempty"`
    Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

func main() {
    mode := flag.String("mode", "stdio", "server mode: stdio | ws (ws is TBD)")
    listen := flag.String("listen", ":8080", "listen address for ws mode (ignored for stdio)")
    delayMS := flag.Int("delay-ms", 0, "artificial latency in milliseconds per request")
    respSize := flag.Int("resp-size-bytes", 0, "size of response payload bytes for tools.exec")
    errorRate := flag.Float64("error-rate", 0.0, "probability (0..1) to return an error")
    burstCalls := flag.Int("burst-calls", 0, "internal call amplification (for load simulation; single response)")
    seed := flag.Int64("seed", time.Now().UnixNano(), "random seed")
    flag.Parse()

    rand.Seed(*seed)

    switch *mode {
    case "stdio":
        runSTDIO(*delayMS, *respSize, *errorRate, *burstCalls)
    case "ws":
        log.Printf("ws mode is not yet implemented; please use --mode stdio for now or wrap with a proxy")
        os.Exit(2)
    default:
        log.Fatalf("unknown --mode: %s", *mode)
    }

    _ = listen // reserved for future ws mode
}

func runSTDIO(delayMS int, respSize int, errorRate float64, burstCalls int) {
    in := bufio.NewScanner(os.Stdin)
    // allow bigger frames
    in.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
    out := bufio.NewWriter(os.Stdout)
    defer out.Flush()

    for in.Scan() {
        line := strings.TrimSpace(in.Text())
        if line == "" {
            continue
        }
        var req rpcReq
        if err := json.Unmarshal([]byte(line), &req); err != nil {
            // Ignore malformed input but emit an error response with null id
            writeResp(out, rpcRes{JSONRPC: "2.0", Error: &rpcError{Code: -32700, Message: "parse error"}})
            continue
        }

        // artificial delay
        if delayMS > 0 {
            time.Sleep(time.Duration(delayMS) * time.Millisecond)
        }

        // internal amplification (no extra responses to keep JSON-RPC semantics)
        if burstCalls > 0 {
            _ = burstCalls // place holder for internal accounting in future
        }

        // random error
        if errorRate > 0 && rand.Float64() < errorRate {
            writeResp(out, rpcRes{JSONRPC: "2.0", ID: req.ID, Error: &rpcError{Code: -32000, Message: "simulated error"}})
            continue
        }

        // build result by method
        var result any
        switch req.Method {
        case "tools.list":
            result = map[string]any{"tools": []string{"echo", "cat"}}
        case "prompts.get":
            result = map[string]any{"id": "demo", "text": "Hello from mcp-test-server"}
        case "tools.exec":
            payload := ""
            if respSize > 0 {
                // generate a repeated 'A' string of requested size (bounded)
                // guard an upper bound to avoid runaway memory
                max := respSize
                if max > 5*1024*1024 { // 5 MiB safe cap
                    max = 5 * 1024 * 1024
                }
                sb := strings.Builder{}
                sb.Grow(max)
                for i := 0; i < max; i++ {
                    sb.WriteByte('A')
                }
                payload = sb.String()
            }
            result = map[string]any{"ok": true, "size": respSize, "payload": payload}
        default:
            result = map[string]any{"ok": true}
        }

        writeResp(out, rpcRes{JSONRPC: "2.0", ID: req.ID, Result: result})
    }

    if err := in.Err(); err != nil {
        log.Printf("stdin read error: %v", err)
    }
}

func writeResp(w *bufio.Writer, res rpcRes) {
    b, err := json.Marshal(res)
    if err != nil {
        log.Printf("marshal response error: %v", err)
        return
    }
    b = append(b, '\n')
    if _, err := w.Write(b); err != nil {
        log.Printf("write error: %v", err)
        return
    }
    _ = w.Flush()
}

