package common

import (
	"io"
	"os"

	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var Log = *logrus.New()

type WampHook struct{}

func NewWampHook() *WampHook {
	return &WampHook{}
}

func (hook *WampHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *WampHook) Fire(entry *logrus.Entry) error {
	publisher, err := GetWampClient()
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
		// drop the shared connection so the next log line reconnects
		ResetWampClient()
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
