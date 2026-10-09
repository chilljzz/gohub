package ws

import (
	"sync"
	"testing"
)

var benchmarkCountSink int

func BenchmarkManagerConnectionCount(b *testing.B) {
	manager := NewManager()

	b.ReportAllocs()
	b.ResetTimer()

	var result int

	for i := 0; i < b.N; i++ {
		result = manager.ConnectionCount()
	}

	benchmarkCountSink = result
}

func BenchmarkManagerConnectionCountParallel(b *testing.B) {
	manager := NewManager()

	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = manager.ConnectionCount()
		}
	})
}

func TestManagerMultipleConnections(t *testing.T) {
	manager := NewManager()

	clientA := NewClient(1001, nil, manager, nil, nil)
	clientB := NewClient(1001, nil, manager, nil, nil)

	if !manager.Register(clientA) {
		t.Fatal("register clientA failed")
	}

	if !manager.Register(clientB) {
		t.Fatal("register clientB failed")
	}

	if got := manager.ConnectionCount(); got != 2 {
		t.Fatalf("connections = %d, want 2", got)
	}

	if got := manager.OnlineUserCount(); got != 1 {
		t.Fatalf("online users = %d, want 1", got)
	}

	manager.Unregister(clientA)

	if got := manager.ConnectionCount(); got != 1 {
		t.Errorf("connections = %d, want 1", got)
	}

	manager.Unregister(clientB)

	if got := manager.OnlineUserCount(); got != 0 {
		t.Errorf("online users = %d, want 0", got)
	}
}

func TestManagerConcurrentRegisterUnregister(t *testing.T) {
	manager := NewManager()

	const workers = 8
	const iterations = 30

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func(worker int) {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				client := NewClient(
					uint(1001+worker%4),
					nil,
					manager,
					nil,
					nil,
				)

				if !manager.Register(client) {
					t.Errorf("register failed")
					return
				}

				manager.JoinConversation(4003, client)

				_ = manager.ConnectionCount()
				_ = manager.OnlineUserCount()
				_ = manager.IsInConversation(4003, client)
				_ = manager.RoomUserIDs(4003)

				manager.BroadcastToConversation(
					4003,
					[]byte("test"),
				)

				manager.Unregister(client)
			}
		}(i)
	}

	wg.Wait()

	if got := manager.ConnectionCount(); got != 0 {
		t.Errorf("connections = %d, want 0", got)
	}

	if got := manager.OnlineUserCount(); got != 0 {
		t.Errorf("online users = %d, want 0", got)
	}

	if users := manager.RoomUserIDs(4003); len(users) != 0 {
		t.Errorf("room still has %d users", len(users))
	}
}

func TestManagerBroadcast(t *testing.T) {
	manager := NewManager()

	clientA := NewClient(
		1001, nil, manager, nil, nil,
	)
	clientB := NewClient(
		1002, nil, manager, nil, nil,
	)

	if !manager.Register(clientA) {
		t.Fatal("register clientA failed")
	}
	if !manager.Register(clientB) {
		t.Fatal("register clientB failed")
	}

	defer manager.Unregister(clientA)
	defer manager.Unregister(clientB)

	manager.JoinConversation(4001, clientA)
	manager.JoinConversation(4001, clientB)

	payload := []byte("hello gohub")

	count := manager.BroadcastToConversation(
		4001, payload,
	)

	if count != 2 {
		t.Fatalf(
			"broadcast count = %d, want 2",
			count,
		)
	}

	for _, client := range []*Client{clientA, clientB} {
		if len(client.Send) != 1 {
			t.Fatalf(
				"send queue length = %d, want 1",
				len(client.Send),
			)
		}

		got := <-client.Send

		if string(got) != string(payload) {
			t.Errorf(
				"got %q, want %q",
				got,
				payload,
			)
		}
	}
}
