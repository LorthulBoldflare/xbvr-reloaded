package server

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	auth "github.com/abbot/go-http-auth"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/gammazero/nexus/v3/router"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/go-openapi/spec"
	"github.com/gorilla/mux"
	wwwlog "github.com/gowww/log"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/koding/websocketproxy"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/peterbourgon/diskv"
	"github.com/rs/cors"
	"willnorris.com/go/imageproxy"

	"github.com/xbapps/xbvr/pkg/api"
	"github.com/xbapps/xbvr/pkg/authlog"
	"github.com/xbapps/xbvr/pkg/common"
	"github.com/xbapps/xbvr/pkg/config"
	"github.com/xbapps/xbvr/pkg/migrations"
	"github.com/xbapps/xbvr/pkg/models"
	"github.com/xbapps/xbvr/pkg/session"
	"github.com/xbapps/xbvr/pkg/tasks"
	"github.com/xbapps/xbvr/ui"
	"github.com/xbapps/xbvr/web"
)

var (
	wsAddr = common.WsAddr
	log    = &common.Log
)

func authHandle(pattern string, authEnabled bool, authSecret auth.SecretProvider, handler http.Handler) {
	if authEnabled {
		authenticator := auth.NewBasicAuthenticator("default", authSecret)
		basicWrapped := authenticator.Wrap(func(res http.ResponseWriter, req *auth.AuthenticatedRequest) {
			http.StripPrefix(pattern, handler).ServeHTTP(res, &req.Request)
		})
		// A valid player-session cookie (minted by a native DeoVR/HereSphere
		// login) is accepted as an alternative to Basic Auth, so the headset
		// browser reaches the Web UI without a second login.
		http.HandleFunc(pattern, func(res http.ResponseWriter, req *http.Request) {
			e := authlog.Start("ui", req, nil)
			defer e.Done()
			markPresented(e, req)
			if hasValidPlayerSession(req) {
				e.AuthMethod = "cookie"
				e.AuthResult = "accepted"
				http.StripPrefix(pattern, handler).ServeHTTP(res, req)
				return
			}
			// The player credential pair (DeoVR/HereSphere) is accepted as an
			// alternative to the UI credentials, and mints the session cookie
			// so the Basic prompt does not reappear for every asset.
			if user, password, ok := req.BasicAuth(); ok && checkPlayerBasicAuth(user, password) {
				e.AuthMethod = "basic-player"
				e.AuthUser = user
				e.AuthResult = "success"
				setPlayerSessionCookie(e, res)
				http.StripPrefix(pattern, handler).ServeHTTP(res, req)
				return
			}
			e.Note("delegating to basic auth (UI credentials)")
			basicWrapped(res, req)
		})
	} else {
		http.Handle(pattern, http.StripPrefix(pattern, handler))
	}
}

func StartServer(version, commit, branch, date string) {
	common.CurrentVersion = version

	config.LoadConfig()
	common.CopyXbvrData()

	// Remove old locks
	models.RemoveAllLocks()

	migrations.Migrate("0024-drop-actions-old")

	// Run migrations in background
	go func() {
		config.UpdateMigrationStatus("", 0, 0, "Starting database migrations...")
		migrations.ProcessCustomSceneRemappingFiles()
		migrations.Migrate("")
		config.CompleteMigration()
	}()

	go tasks.CheckDependencies()
	models.CheckVolumes()

	models.InitSites()

	restful.DefaultContainer.EnableContentEncoding(true)
	restful.DefaultContainer.Filter(apiAuthFilter)

	// API endpoints
	ws := new(restful.WebService)
	ws.Route(ws.GET("/").To(func(req *restful.Request, resp *restful.Response) {
		// Everyone lands on the login page first; a valid session cookie
		// there forwards straight to the Web UI.
		e := authlog.Start("root", req.Request, nil)
		e.PlayerClient = isPlayerClient(req.Request)
		e.RedirectTo = "/login"
		e.Done()
		resp.AddHeader("Location", "/login")
		resp.WriteHeader(http.StatusFound)
	}))

	restful.Add(ws)
	restful.Add(api.SceneResource{}.WebService())
	restful.Add(api.ActorResource{}.WebService())
	restful.Add(api.TaskResource{}.WebService())
	restful.Add(api.DMSResource{}.WebService())
	restful.Add(api.ConfigResource{}.WebService())
	restful.Add(api.FilesResource{}.WebService())
	restful.Add(api.DeoVRResource{}.WebService())
	restful.Add(api.DeoVRDeeplinkResource{}.WebService())
	restful.Add(api.HeresphereResource{}.WebService())
	restful.Add(api.PlaylistResource{}.WebService())
	restful.Add(api.AkaResource{}.WebService())
	restful.Add(api.TagGroupResource{}.WebService())
	restful.Add(api.ExternalReference{}.WebService())

	restConfig := restfulspec.Config{
		WebServices: restful.RegisteredWebServices(),
		APIPath:     "/api.json",
		PostBuildSwaggerObjectHandler: func(swo *spec.Swagger) {
			e := spec.VendorExtensible{}
			e.AddExtension("x-logo", map[string]interface{}{
				"url": "/ui/icons/xbvr-512.png",
			})

			swo.Info = &spec.Info{
				InfoProps: spec.InfoProps{
					Title:   "XBVR API",
					Version: common.CurrentVersion,
				},
				VendorExtensible: e,
			}
			swo.Tags = []spec.Tag{
				{
					TagProps: spec.TagProps{
						Name:        "Config",
						Description: "Endpoints used by options screen",
					},
				},
				{
					TagProps: spec.TagProps{
						Name:        "DeoVR",
						Description: "Endpoints for interfacing with DeoVR player",
					},
				},
				{
					TagProps: spec.TagProps{
						Name:        "HereSphere",
						Description: "Endpoints for interfacing with HereSphere player",
					},
				},
			}
		},
	}
	restful.Add(restfulspec.NewOpenAPIService(restConfig))

	// Static files
	authHandle("/ui/", common.IsUIAuthEnabled(), common.GetUISecret, http.FileServer(ui.GetFileSystem(common.EnvConfig.Debug)))
	// New SPA (React), same auth semantics as /ui/.
	authHandle("/web/", common.IsUIAuthEnabled(), common.GetUISecret, web.GetHandler(common.EnvConfig.Debug))

	// Login page (issues the xbvr_player_session cookie). Standalone
	// handler: reachable without prior auth, open to all clients.
	http.HandleFunc("/login", loginHandler)

	// Imageproxy. Two backends behind the context handler: a persistent disk
	// cache for attributed requests, and a NopCache for zero-id (unattributed,
	// transient) requests. See ImageProxyContextHandler for the routing rule.
	r := mux.NewRouter()
	cachingProxy, noCacheProxy := newImageProxyBackends(filepath.Join(common.AppDir, "imageproxy"), false)
	r.PathPrefix("/img/").Handler(newImageProxyHandler(cachingProxy, noCacheProxy))
	hmp := NewHeatmapThumbnailProxy(cachingProxy, diskCache(filepath.Join(common.AppDir, "heatmapthumbnailproxy")))
	r.PathPrefix("/imghm/").Handler(http.StripPrefix("/imghm", hmp))
	downloadhandler := DownloadHandler{}
	r.PathPrefix("/download/").Handler(http.StripPrefix("/download/", downloadhandler))
	myfileshandler := MyFilesHandler{}
	r.PathPrefix("/myfiles/").Handler(http.StripPrefix("/myfiles/", myfileshandler))
	r.SkipClean(true)

	r.PathPrefix("/").Handler(http.DefaultServeMux)

	// CORS
	handler := cors.Default().Handler(r)

	// WAMP router
	routerConfig := &router.Config{
		Debug: false,
		RealmConfigs: []*router.RealmConfig{
			{
				URI:           wamp.URI("default"),
				AnonymousAuth: true,
				AllowDisclose: false,
			},
		},
	}

	wampRouter, err := router.NewRouter(routerConfig, log)
	if err != nil {
		log.Fatal(err)
	}
	defer wampRouter.Close()

	// Run websocket server. Bound to loopback (see common.WsAddr); the
	// default origin check applies (Origin host must match the Host header).
	wss := router.NewWebsocketServer(wampRouter)
	wsCloser, err := wss.ListenAndServe(wsAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer wsCloser.Close()

	// Proxy websocket
	wsURL, err := url.Parse("ws://" + wsAddr)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/ws/", func(w http.ResponseWriter, req *http.Request) {
		// When UI auth is enabled, require the same credentials on the
		// websocket proxy. Browsers send cached basic-auth credentials on the
		// same-origin WS handshake, so the UI works unchanged, while a
		// cross-origin (e.g. DNS-rebinding) page cannot authenticate. A valid
		// player-session cookie (attached by the browser to the same-origin
		// handshake) is accepted as an alternative credential.
		e := authlog.Start("ws", req, nil)
		markPresented(e, req)
		if common.IsUIAuthEnabled() {
			if hasValidPlayerSession(req) {
				e.AuthMethod = "cookie"
				e.AuthResult = "accepted"
			} else {
				user, password, ok := req.BasicAuth()
				uiOK := ok && user == common.EnvConfig.UIUsername && checkUIPassword(password)
				playerOK := !uiOK && checkPlayerBasicAuth(user, password)
				if !uiOK && !playerOK {
					if ok {
						e.AuthMethod = "basic"
						e.AuthUser = user
						e.AuthResult = "failed"
					} else {
						e.AuthMethod = "none"
						e.AuthResult = "denied"
					}
					e.Done()
					w.Header().Set("WWW-Authenticate", `Basic realm="default"`)
					http.Error(w, "401: Unauthorized", http.StatusUnauthorized)
					return
				}
				if playerOK {
					e.AuthMethod = "basic-player"
				} else {
					e.AuthMethod = "basic-ui"
				}
				e.AuthUser = user
				e.AuthResult = "success"
			}
		} else {
			e.Note("auth disabled, allowing handshake")
		}
		// CSWSH note: a websocket initiated by a standard web page (Origin
		// header present) should originate from the XBVR UI itself, i.e. the
		// Origin host equals the request Host. Players are non-standard
		// browsers whose Origin headers cannot be relied upon, so a mismatch
		// is logged but the request is allowed. Non-browser clients without
		// an Origin header are unaffected.
		if !wsOriginAllowed(req) {
			log.Warnf("websocket request with non-matching Origin %q (Host %q) — allowing (non-standard browser)",
				req.Header.Get("Origin"), req.Host)
			e.Note("origin mismatch on handshake (allowed): Origin %q vs Host %q",
				req.Header.Get("Origin"), req.Host)
		}
		e.Note("handshake accepted, proxying (stream payload not captured)")
		e.Done()
		// The origin has been validated above; remove it so the WAMP server's
		// own same-host check (which sees the loopback backend address, not
		// the UI host) does not reject the forwarded request.
		req.Header.Del("Origin")
		handler := websocketproxy.ProxyHandler(wsURL)
		handler.ServeHTTP(w, req)
	})

	// MCP endpoint (Streamable HTTP). Only served when UI auth is enabled —
	// without UI credentials there would be no bearer token to protect it,
	// so the route is not registered at all (requests fall through to 404).
	// See mcpAuthMiddleware for the token check.
	if common.IsUIAuthEnabled() {
		mcpServer := newMCPServer(version)
		http.Handle("/mcp", mcpAuthMiddleware(mcp.NewStreamableHTTPHandler(
			func(r *http.Request) *mcp.Server { return mcpServer }, nil)))
	}

	// Attach logrus hook
	wampHook := common.NewWampHook()
	log.AddHook(wampHook)

	log.Infof("XBVR %v (build date %v) starting...", version, date)

	// DMS
	if config.Config.Interfaces.DLNA.Enabled {
		go tasks.StartDMS()
	}

	// DeoVR remote
	go session.DeoRemote()

	// Cron
	SetupCron()

	// List binding addresses
	addrs, _ := net.InterfaceAddrs()
	ips := []string{}
	for _, addr := range addrs {
		ip, _ := addr.(*net.IPNet)
		if ip.IP.To4() != nil {
			ips = append(ips, fmt.Sprintf("http://%v:%v/", ip.IP, config.Config.Server.Port))
		}
	}

	// Prepare state
	tasks.UpdateState()
	config.State.Server.BoundIP = ips
	config.SaveState()

	log.Infof("Web UI available at %s", strings.Join(ips, ", "))
	if common.IsUIAuthEnabled() {
		log.Infof("MCP endpoint available at /mcp (Streamable HTTP)")
	} else {
		log.Infof("MCP endpoint disabled (set UI_USERNAME/UI_PASSWORD to enable)")
	}
	log.Infof("Web UI Authentication enabled: %v", common.IsUIAuthEnabled())
	log.Infof("Using database: %s", common.DATABASE_URL)

	httpAddr := fmt.Sprintf("%v:%v", config.Config.Server.BindAddress, config.Config.Server.Port)
	if common.EnvConfig.DebugRequests {
		log.Fatal(http.ListenAndServe(httpAddr, wwwlog.Handle(handler, &wwwlog.Options{Color: true})))
	} else {
		log.Fatal(http.ListenAndServe(httpAddr, handler))
	}
}

// newImageProxyBackends builds the two imageproxy instances used behind the
// /img context handler: a persistent disk-cache proxy for attributed
// requests (ForceCache, FollowRedirects) and a NopCache proxy for zero-id
// (unattributed, transient) requests. Both share the force-cache transport
// and the same hardening. Tests use this exact constructor so the behaviour
// suite exercises the production wiring.
//
// skipBlocklist must be false in production. Tests with loopback upstreams
// set it true: the DenyHosts list and the SSRF-safe transport wrapper are
// then omitted (both block loopback by design), while the cache-header
// forcing — the part of the transport that is xbvr code — stays in place.
func newImageProxyBackends(cachePath string, skipBlocklist bool) (caching, noCache *imageproxy.Proxy) {
	transport := NewForceCacheTransport(skipBlocklist)
	caching = imageproxy.NewProxy(transport, diskCache(cachePath))
	caching.ForceCache = true
	noCache = imageproxy.NewProxy(transport, imageproxy.NopCache)
	for _, p := range []*imageproxy.Proxy{caching, noCache} {
		p.FollowRedirects = true
		p.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36"
		if !skipBlocklist {
			// SSRF protection: never fetch loopback, link-local or private-range
			// targets. Relative remote URLs are rejected too (no DefaultBaseURL) —
			// all legitimate uses proxy absolute http(s) image URLs. Non-http(s)
			// schemes fail in the HTTP transport before any fetch happens.
			p.DenyHosts = deniedProxyHosts
		}
		// If the client request has a cache-control header (such as 'no-cache'), pass them
		// onto the imageproxy so that this can be respected.
		p.PassRequestHeaders = append(p.PassRequestHeaders, "Cache-Control")
	}
	return caching, noCache
}

// newImageProxyHandler assembles the full /img handler chain: context
// parsing, recording and cache routing, wrapped in the outgoing short
// client-cache header override.
func newImageProxyHandler(caching, noCache http.Handler) http.Handler {
	return ForceShortCacheHandler(NewImageProxyContextHandler(caching, noCache))
}

func diskCache(path string) *diskcache.Cache {
	d := diskv.New(diskv.Options{
		BasePath:  path,
		Transform: func(s string) []string { return []string{s[0:2], s[2:4]} },
	})
	return diskcache.NewWithDiskv(d)
}
