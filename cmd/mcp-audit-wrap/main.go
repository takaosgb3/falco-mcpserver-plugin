package main

import (
    "flag"
    "fmt"
    "io"
    "log"
    "math/rand"
    "os"
    "os/exec"
    "strings"
    "sync/atomic"
    "time"

    "github.com/takaosgb3/falco-mcpserver-plugin/pkg/audit"
)

func main() {
    sink := flag.String("sink", "stdout", "audit sink file path or 'stdout'")
    session := flag.String("session-id", "", "session id (optional)")
    client := flag.String("client-process", "unknown", "client process name (e.g., claude-code, codex-cli)")
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: %s [--sink path|stdout] [--session-id id] [--client-process name] -- <cmd> [args...]\n", os.Args[0])
        flag.PrintDefaults()
    }
    flag.Parse()
    sep := flag.NArg()
    // Require -- separator and command
    args := os.Args
    pos := -1
    for i, a := range args {
        if a == "--" {
            pos = i
            break
        }
    }
    if pos == -1 || pos+1 >= len(args) {
        flag.Usage()
        os.Exit(2)
    }
    cmdPath := args[pos+1]
    cmdArgs := args[pos+2:]

    if *session == "" {
        *session = fmt.Sprintf("sess-%d", time.Now().UnixNano()+int64(rand.Intn(1000)))
    }

    w, err := audit.NewWriter(*sink)
    if err != nil {
        log.Fatalf("open sink: %v", err)
    }
    defer w.Close()

    cmd := exec.Command(cmdPath, cmdArgs...)
    childStdin, err := cmd.StdinPipe()
    if err != nil {
        log.Fatalf("stdin pipe: %v", err)
    }
    childStdout, err := cmd.StdoutPipe()
    if err != nil {
        log.Fatalf("stdout pipe: %v", err)
    }
    cmd.Stderr = os.Stderr

    if err := cmd.Start(); err != nil {
        log.Fatalf("start child: %v", err)
    }

    var reqBytes uint64
    var respBytes uint64

    done := make(chan struct{}, 2)
    go func() {
        n, _ := io.Copy(childStdin, os.Stdin)
        atomic.AddUint64(&reqBytes, uint64(n))
        _ = childStdin.Close()
        done <- struct{}{}
    }()
    go func() {
        n, _ := io.Copy(os.Stdout, childStdout)
        atomic.AddUint64(&respBytes, uint64(n))
        done <- struct{}{}
    }()

    <-done
    <-done

    if err := cmd.Wait(); err != nil {
        // still emit audit event with non-zero error exit
        log.Printf("child exited with error: %v", err)
    }

    // Emit one summary event (skeleton)
    _ = w.Emit(audit.Event{
        SessionID:       *session,
        ClientProcess:   *client,
        ServerHost:      "stdio",
        ServerPort:      0,
        TLS:             false,
        Method:          "",
        RequestBytes:    atomic.LoadUint64(&reqBytes),
        ResponseBytes:   atomic.LoadUint64(&respBytes),
        ToolInvokeCount: 0,
        FileAccessCount: 0,
    })

    // Propagate exit code from child
    if cmd.ProcessState != nil {
        if status, ok := cmd.ProcessState.Sys().(interface{ ExitStatus() int }); ok {
            os.Exit(status.ExitStatus())
        }
    }
    // Fallback
    if strings.Contains(fmt.Sprint(err), "exit status ") {
        os.Exit(1)
    }
}

