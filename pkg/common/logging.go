package common

import (
	"context"
	"io"
	"os"
	"sync"

	"github.com/gammazero/nexus/v3/client"
	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var Log = *logrus.New()

type WampHook struct {
	mu        sync.Mutex
	publisher *client.Client
}

func NewWampHook() *WampHook {
	return &WampHook{}
}

// getPublisher lazily connects to the WAMP router and reconnects after the
// connection drops, so logging never panics when the router is unavailable.
func (hook *WampHook) getPublisher() (*client.Client, error) {
	hook.mu.Lock()
	defer hook.mu.Unlock()

	if hook.publisher != nil {
		return hook.publisher, nil
	}

	publisher, err := client.ConnectNet(context.Background(), "ws://"+WsAddr+"/ws", client.Config{
		Realm: "default",
	})
	if err != nil {
		return nil, err
	}
	hook.publisher = publisher
	return publisher, nil
}

func (hook *WampHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *WampHook) Fire(entry *logrus.Entry) error {
	publisher, err := hook.getPublisher()
	if err != nil {
		return err
	}
	err = publisher.Publish("service.log", nil, nil, map[string]interface{}{
		"level":     entry.Level.String(),
		"message":   entry.Message,
		"data":      entry.Data,
		"timestamp": entry.Time.String(),
	})
	if err != nil {
		// drop the connection so the next log line reconnects
		hook.mu.Lock()
		hook.publisher = nil
		hook.mu.Unlock()
		return err
	}
	return nil
}

func InitLogging() {
	//	Log.Out = os.Stdout
	Log.SetLevel(logrus.InfoLevel)
	if EnvConfig.Debug {
		Log.SetLevel(logrus.DebugLevel)
	}

	Log.Formatter = &prefixed.TextFormatter{
		ForceColors: true,
	}

	//	create / open log file in AppDir folder
	lfile, err := os.OpenFile(AppDir+"/xbvr.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		//		defer lfile.Close()
		mw := io.MultiWriter(lfile, os.Stdout)
		Log.Out = ansicolor.NewAnsiColorWriter(mw)
	} else {
		Log.Info("Failed to log to file, using default stderr")
	}
}
