package common

import (
	"context"
	"sync"

	"github.com/gammazero/nexus/v3/client"
)

var (
	wampClientMu sync.Mutex
	wampClient   *client.Client
)

// GetWampClient returns the shared WAMP client, connecting lazily on first
// use. PublishWS is called for every log line and every DeoVR loop tick, so
// connecting per message was significant churn.
func GetWampClient() (*client.Client, error) {
	wampClientMu.Lock()
	c := wampClient
	wampClientMu.Unlock()
	if c != nil {
		return c, nil
	}

	// Dial outside the mutex: a slow or unreachable WAMP server must not
	// serialize every publisher behind one blocked dial. Concurrent dials are
	// harmless — the winner's client is cached, losers close theirs.
	nc, err := client.ConnectNet(context.Background(), "ws://"+WsAddr+"/ws", client.Config{Realm: "default"})
	if err != nil {
		return nil, err
	}

	wampClientMu.Lock()
	if wampClient != nil {
		existing := wampClient
		wampClientMu.Unlock()
		nc.Close()
		return existing, nil
	}
	wampClient = nc
	wampClientMu.Unlock()
	return nc, nil
}

// ResetWampClient drops the shared client after a publish failure so the
// next call reconnects.
func ResetWampClient() {
	wampClientMu.Lock()
	defer wampClientMu.Unlock()
	if wampClient != nil {
		wampClient.Close()
		wampClient = nil
	}
}

func PublishWS(topic string, message map[string]interface{}) error {
	publisher, err := GetWampClient()
	if err != nil {
		return err
	}
	if EnvConfig.DebugWS {
		Log.Debugf("Sending WAMP message: %v %v", topic, message)
	}
	if err := publisher.Publish(topic, nil, nil, message); err != nil {
		ResetWampClient()
		return err
	}
	return nil
}
