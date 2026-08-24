package tasks

import (
	"bytes"
	"image"
	"image/png"
	"io"
	"net"
	"path/filepath"
	"time"

	"github.com/nfnt/resize"
	"github.com/xbapps/xbvr/pkg/config"
	"github.com/xbapps/xbvr/pkg/dms/dlna/dms"
	"github.com/xbapps/xbvr/ui"
)

type dmsConfig struct {
	Path                string
	IfName              string
	Http                string
	FriendlyName        string
	LogHeaders          bool
	NoTranscode         bool
	NoProbe             bool
	StallEventSubscribe bool
	NotifyInterval      time.Duration
	IgnoreHidden        bool
	IgnoreUnreadable    bool
}

var dmsServer *dms.Server
var dmsStarted bool

func initDMS() error {
	var dmsConfig = &dmsConfig{
		Path:           "",
		IfName:         "",
		Http:           ":1338",
		FriendlyName:   "",
		LogHeaders:     false,
		NotifyInterval: 30 * time.Second,
	}

	httpConn, err := net.Listen("tcp", dmsConfig.Http)
	if err != nil {
		return err
	}

	dmsServer = &dms.Server{
		Interfaces: func(ifName string) (ifs []net.Interface) {
			var err error
			if ifName == "" {
				ifs, err = net.Interfaces()
			} else {
				var if_ *net.Interface
				if_, err = net.InterfaceByName(ifName)
				if if_ != nil {
					ifs = append(ifs, *if_)
				}
			}
			if err != nil {
				log.Error(err)
			}
			var tmp []net.Interface
			for _, if_ := range ifs {
				if if_.Flags&net.FlagUp == 0 || if_.MTU <= 0 {
					continue
				}
				tmp = append(tmp, if_)
			}
			ifs = tmp
			return
		}(dmsConfig.IfName),
		HTTPConn:       httpConn,
		FriendlyName:   dmsConfig.FriendlyName,
		RootObjectPath: filepath.Clean(dmsConfig.Path),
		LogHeaders:     dmsConfig.LogHeaders,
		NoTranscode:    dmsConfig.NoTranscode,
		NoProbe:        dmsConfig.NoProbe,
		Icons: []dms.Icon{
			{
				Width:      48,
				Height:     48,
				Depth:      8,
				Mimetype:   "image/png",
				ReadSeeker: readIcon(config.Config.Interfaces.DLNA.ServiceImage, 48),
			},
			{
				Width:      128,
				Height:     128,
				Depth:      8,
				Mimetype:   "image/png",
				ReadSeeker: readIcon(config.Config.Interfaces.DLNA.ServiceImage, 128),
			},
		},
		StallEventSubscribe: dmsConfig.StallEventSubscribe,
		NotifyInterval:      dmsConfig.NotifyInterval,
		IgnoreHidden:        dmsConfig.IgnoreHidden,
		IgnoreUnreadable:    dmsConfig.IgnoreUnreadable,
	}
	return nil
}

func getIconReader(fn string) (io.Reader, error) {
	b, err := ui.Assets.ReadFile("dist/dlna/" + fn + ".png")
	return bytes.NewReader(b), err
}

func readIcon(path string, size uint) *bytes.Reader {
	r, err := getIconReader(path)
	if err != nil {
		log.Errorf("Cannot read DLNA icon %s: %v", path, err)
		return bytes.NewReader(nil)
	}
	imageData, _, err := image.Decode(r)
	if err != nil {
		log.Errorf("Cannot decode DLNA icon %s: %v", path, err)
		return bytes.NewReader(nil)
	}
	return resizeImage(imageData, size)
}

func resizeImage(imageData image.Image, size uint) *bytes.Reader {
	img := resize.Resize(size, size, imageData, resize.Lanczos3)
	var buff bytes.Buffer
	png.Encode(&buff, img)
	return bytes.NewReader(buff.Bytes())
}

func StartDMS() {
	if err := initDMS(); err != nil {
		log.Errorf("DLNA server not started: %v", err)
		return
	}
	go func() {
		log.Info("Starting DLNA")
		if err := dmsServer.Serve(); err != nil {
			log.Errorf("DLNA server stopped with error: %v", err)
		}
	}()
	dmsStarted = true
}

func StopDMS() {
	log.Info("Stopping DLNA")
	err := dmsServer.Close()
	if err != nil {
		log.Errorf("Error stopping DLNA server: %v", err)
	}
	dmsStarted = false
}

func IsDMSStarted() bool {
	return dmsStarted
}
