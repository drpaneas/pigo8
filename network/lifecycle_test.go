package network

import (
	"net"
	"testing"
	"time"
)

func testServerConfig(address string) *Config {
	return &Config{
		Role:       RoleServer,
		Address:    address,
		Port:       0,
		PlayerID:   "test-player",
		BufferSize: 4,
		GameName:   "test-game",
	}
}

func TestInitNetworkTwiceReturns(t *testing.T) {
	ShutdownNetwork()

	if err := InitNetwork(testServerConfig("127.0.0.1")); err != nil {
		t.Fatalf("first init failed: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- InitNetwork(testServerConfig("127.0.0.1"))
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("second init failed: %v", err)
		}
		ShutdownNetwork()
	case <-time.After(500 * time.Millisecond):
		t.Fatal("second InitNetwork call did not return")
	}
}

func TestInitNetworkHonorsLoopbackBind(t *testing.T) {
	ShutdownNetwork()
	defer ShutdownNetwork()

	if err := InitNetwork(testServerConfig("127.0.0.1")); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	if networkManager == nil || networkManager.udpConn == nil {
		t.Fatal("expected UDP listener to be initialized")
	}

	localAddr, ok := networkManager.udpConn.LocalAddr().(*net.UDPAddr)
	if !ok {
		t.Fatalf("expected UDP local address, got %T", networkManager.udpConn.LocalAddr())
	}

	if !localAddr.IP.IsLoopback() {
		t.Fatalf("expected loopback bind, got %s", localAddr.IP.String())
	}
}
