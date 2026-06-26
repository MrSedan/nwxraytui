package ipc_test

import (
	"bufio"
	"net"
	"path/filepath"
	"testing"

	"github.com/mrsedan/nwxraytui/internal/ipc"
)

func TestServerClient(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "daemon.sock")

	srv, err := ipc.NewServer(sock)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	connCh := make(chan net.Conn, 1)
	go func() {
		conn, err := srv.Accept()
		if err == nil {
			connCh <- conn
		}
	}()

	client, err := ipc.NewClient(sock)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	if err := client.Send(ipc.CmdRefresh{}); err != nil {
		t.Fatal(err)
	}

	serverConn := <-connCh
	defer serverConn.Close()

	sc := bufio.NewScanner(serverConn)
	if !sc.Scan() {
		t.Fatal("no data from client")
	}
	env, err := ipc.Decode(sc.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if env.Type != ipc.TypeCmdRefresh {
		t.Fatalf("type: got %q", env.Type)
	}
}
