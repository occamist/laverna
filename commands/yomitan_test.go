package commands

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"
)

func freePort(t *testing.T) int {
	t.Helper()

	var lc net.ListenConfig
	ln, err := lc.Listen(t.Context(), "tcp", "localhost:0")
	if err != nil {
		t.Fatalf("net.Listen(): %v", err)
	}
	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("%T.Addr(): got %T, want *net.TCPAddr", ln, ln.Addr())
	}
	if err := ln.Close(); err != nil {
		t.Fatalf("%T.Close(): %v", ln, err)
	}
	return addr.Port
}

func TestYomitanCmd(t *testing.T) {
	t.Parallel()

	port := freePort(t)
	ctx, cancel := context.WithCancel(t.Context())

	errCh := make(chan error, 1)
	go func() {
		errCh <- yomitanCmd(ctx, yomitanCmdFlags{Host: "localhost", Port: port})
	}()

	url := fmt.Sprintf("http://localhost:%d/?term=Hello&language=en", port)
	var resp *http.Response
	var err error
	for range 50 {
		resp, err = http.Get(url) //nolint:gosec,noctx // test-only request against a URL we just constructed
		if err == nil {
			break
		}
	}
	if err != nil {
		cancel()
		t.Fatalf("http.Get(%q): %v", url, err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("http.Get(%q): status=%d, want %d", url, resp.StatusCode, http.StatusOK)
	}
	if got := resp.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("http.Get(%q): Content-Type=%q, want %q", url, got, "application/json")
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("yomitanCmd(): got err=%v, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("yomitanCmd() did not shut down in time")
	}
}
