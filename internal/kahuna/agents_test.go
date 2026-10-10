/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/kowabunga-cloud/common/wsrpc"
)

// newTestAgentSession returns a WebSocket RPC client connected to a dummy
// peer that just consumes messages until the connection is closed.
func newTestAgentSession(t *testing.T, id string) *KowabungaAgent {
	t.Helper()
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() {
			_ = c.Close()
		}()
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				return
			}
		}
	}))
	t.Cleanup(srv.Close)

	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	return &KowabungaAgent{
		ID:          id,
		Client:      wsrpc.NewWsRpcClient(conn, false),
		Type:        "test",
		IsConnected: true,
	}
}

func registerTestAgentSession(ag *KowabungaAgent) {
	agentsLock.Lock()
	GetAgents()[ag.ID] = ag
	agentsLock.Unlock()
	go ag.WatchKeepalive()
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestDisconnectAgentDoesNotBlock(t *testing.T) {
	ag := newTestAgentSession(t, "agent-disconnect")
	registerTestAgentSession(ag)

	done := make(chan struct{})
	go func() {
		DisconnectAgent(ag.ID)
		DisconnectAgent(ag.ID) // second call on an already disconnected agent
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("DisconnectAgent blocked")
	}

	if GetAgent(ag.ID) != nil {
		t.Error("disconnected agent should no longer be registered")
	}
}

func TestStaleAgentSessionDoesNotUnregisterNewOne(t *testing.T) {
	old := newTestAgentSession(t, "agent-reconnect")
	registerTestAgentSession(old)

	// agent reconnects: new session replaces the old one, which gets closed
	renewed := newTestAgentSession(t, old.ID)
	registerTestAgentSession(renewed)
	if err := old.Client.Close(); err != nil {
		t.Fatal(err)
	}

	// give the old session's watcher time to notice its termination
	time.Sleep(200 * time.Millisecond)
	if current := GetAgent(old.ID); current == nil || current.Client != renewed.Client {
		t.Fatal("old session termination unregistered the new session")
	}

	// while the current session's termination unregisters the agent
	if err := renewed.Client.Close(); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "agent unregistration", func() bool { return GetAgent(old.ID) == nil })
}
