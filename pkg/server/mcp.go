package server

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/xbapps/xbvr/pkg/api"
	"github.com/xbapps/xbvr/pkg/common"
	"github.com/xbapps/xbvr/pkg/models"
	"github.com/xbapps/xbvr/pkg/tasks"
)

// mcpAuthMiddleware enforces Bearer-token auth on the /mcp endpoint whenever
// UI auth is enabled (UI_USERNAME/UI_PASSWORD set), matching the scope of
// apiAuthFilter for the REST API. The accepted token is the concatenation of
// the configured UI username and UI password, e.g. username "UserA" with
// password "Password123" yields the token "UserAPassword123". Both values are
// plaintext env config, so the token is compared as a whole — never split it
// back into user/password parts. When UI auth is disabled the endpoint is
// open.
func mcpAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !common.IsUIAuthEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		expected := common.EnvConfig.UIUsername + common.EnvConfig.UIPassword
		if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
			http.Error(w, "401: Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type mcpNoArgs struct{}

type mcpScrapeSceneArgs struct {
	URL     string `json:"url" jsonschema:"URL of the scene page to scrape - do not use links requiring a login"`
	SceneID string `json:"scene_id,omitempty" jsonschema:"scene ID (excluding site prefix) - required for wetvr.com scenes"`
}

func mcpTextResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

// newMCPServer builds the MCP server exposing xbvr maintenance tools.
func newMCPServer(version string) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "xbvr",
		Version: version,
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name: "rescan_storage",
		Description: "Rescan all storage folders for new, changed and removed video files. " +
			"Same as Options -> Storage -> 'Rescan all folders' in the web UI. Runs asynchronously.",
	}, mcpRescanStorage)

	mcp.AddTool(server, &mcp.Tool{
		Name: "scrape_scene",
		Description: "Scrape a single scene from its scene-page URL and store it in the library. " +
			"Same as Options -> Create/Import scene -> 'Scrape a scene' in the web UI, including confirming " +
			"the scene detail popup. Returns the scene-id as determined by the scraper, or null if the " +
			"scene could not be scraped. If the scene is already present, its scene-id is returned without " +
			"re-scraping. May take tens of seconds for a new scene.",
	}, mcpScrapeScene)

	mcp.AddTool(server, &mcp.Tool{
		Name: "generate_previews",
		Description: "Start generating video preview clips for all scenes. " +
			"Same as Options -> Previews -> 'Start generating previews' in the web UI. Runs asynchronously.",
	}, mcpGeneratePreviews)

	return server
}

func mcpRescanStorage(ctx context.Context, req *mcp.CallToolRequest, args mcpNoArgs) (*mcp.CallToolResult, any, error) {
	if models.CheckLock("rescan") {
		return mcpTextResult("A storage rescan is already running"), nil, nil
	}

	go tasks.RescanVolumes(-1)
	return mcpTextResult("Rescan of all storage folders started"), nil, nil
}

func mcpGeneratePreviews(ctx context.Context, req *mcp.CallToolRequest, args mcpNoArgs) (*mcp.CallToolResult, any, error) {
	if models.CheckLock("previews") {
		status, _ := json.Marshal(tasks.GetPreviewQueueStatus())
		return mcpTextResult("Preview generation is already running: " + string(status)), nil, nil
	}

	go tasks.GeneratePreviews(nil)
	return mcpTextResult("Preview generation started"), nil, nil
}

// mcpSiteForURL replicates the scraper (site) selection the web UI performs in
// OptionsSceneCreate.vue: match the URL against registered scraper domains,
// then apply the single-scene overrides for the vendor aggregators.
func mcpSiteForURL(url string) string {
	lowerURL := strings.ToLower(url)

	site := ""
	for _, scraper := range models.GetScrapers() {
		if scraper.Domain != "" && strings.Contains(lowerURL, strings.ToLower(scraper.Domain)) {
			site = scraper.ID
		}
	}

	overrides := []struct {
		domain string
		site   string
	}{
		{"sexlikereal.com", "slr-single_scene"},
		{"czechvrnetwork.com", "czechvr-single_scene"},
		{"povr.com", "povr-single_scene"},
		{"vrporn.com", "vrporn-single_scene"},
		{"vrphub.com", "vrphub-single_scene"},
		{"realvr.com", "realvr-single_scene"},
		{"stashdb.org", "single_scene-stashdb"},
	}
	for _, o := range overrides {
		if strings.Contains(lowerURL, o.domain) {
			site = o.site
		}
	}

	return site
}

// mcpBuildEditRequest replicates the normalization the EditScene modal applies
// when the user confirms ("Save Scene Details"): fall back to the first
// gallery image as cover, normalize the gallery (cover first, deduped), and
// rebuild the images metadata preserving existing type/orientation.
func mcpBuildEditRequest(scene models.Scene) api.RequestEditSceneDetails {
	type imageMeta struct {
		URL         string `json:"url"`
		Type        string `json:"type"`
		Orientation string `json:"orientation"`
	}

	var originalImages []imageMeta
	if err := json.Unmarshal([]byte(scene.Images), &originalImages); err != nil {
		originalImages = []imageMeta{}
	}

	gallery := make([]string, 0, len(originalImages))
	for _, img := range originalImages {
		gallery = append(gallery, img.URL)
	}

	coverURL := scene.CoverURL
	if coverURL == "" && len(gallery) > 0 {
		coverURL = gallery[0]
	}

	// Normalize gallery: cover first, no duplicates
	normalized := make([]string, 0, len(gallery))
	seen := make(map[string]bool, len(gallery))
	if coverURL != "" {
		normalized = append(normalized, coverURL)
		seen[coverURL] = true
	}
	for _, u := range gallery {
		if !seen[u] {
			normalized = append(normalized, u)
			seen[u] = true
		}
	}

	images := make([]imageMeta, 0, len(normalized))
	for _, u := range normalized {
		meta := imageMeta{URL: u}
		for _, orig := range originalImages {
			if orig.URL == u {
				meta.Type = orig.Type
				meta.Orientation = orig.Orientation
				break
			}
		}
		if meta.Type != "cover" && meta.Type != "gallery" {
			if u == coverURL {
				meta.Type = "cover"
			} else {
				meta.Type = "gallery"
			}
		}
		images = append(images, meta)
	}
	imagesJSON, _ := json.Marshal(images)

	var files []string
	if err := json.Unmarshal([]byte(scene.FilenamesArr), &files); err != nil || files == nil {
		files = []string{}
	}
	filenamesJSON, _ := json.Marshal(files)

	cast := make([]string, 0, len(scene.Cast))
	for _, c := range scene.Cast {
		cast = append(cast, c.Name)
	}
	tags := make([]string, 0, len(scene.Tags))
	for _, t := range scene.Tags {
		tags = append(tags, t.Name)
	}

	return api.RequestEditSceneDetails{
		Title:        scene.Title,
		Synopsis:     scene.Synopsis,
		Studio:       scene.Studio,
		Site:         scene.Site,
		SceneURL:     scene.SceneURL,
		ReleaseDate:  scene.ReleaseDateText,
		Cast:         cast,
		Tags:         tags,
		FilenamesArr: string(filenamesJSON),
		Images:       string(imagesJSON),
		CoverURL:     coverURL,
		IsMultipart:  scene.IsMultipart,
		Duration:     fmt.Sprintf("%d", scene.Duration),
	}
}

func mcpScrapeScene(ctx context.Context, req *mcp.CallToolRequest, args mcpScrapeSceneArgs) (*mcp.CallToolResult, any, error) {
	if args.URL == "" {
		return nil, nil, fmt.Errorf("url is required")
	}
	if err := common.ValidateOutboundURL(args.URL); err != nil {
		return nil, nil, fmt.Errorf("scene URL is not allowed: %v", err)
	}

	// Already present? Return the existing scene-id without re-scraping.
	commonDb, _ := models.GetCommonDB()
	var existing models.Scene
	commonDb.Where("scene_url like ?", strings.TrimSuffix(args.URL, "/")+"%").First(&existing)
	if existing.ID != 0 {
		return mcpTextResult(existing.SceneID), nil, nil
	}

	site := mcpSiteForURL(args.URL)
	if site == "" {
		return nil, nil, fmt.Errorf("no scrapers exist for this domain")
	}

	additionalInfo := "[]"
	if strings.Contains(strings.ToLower(args.URL), "wetvr.com") {
		// WetVR scenes cannot be scraped without the scene id (the web UI
		// shows a modal asking for it).
		if args.SceneID == "" {
			return nil, nil, fmt.Errorf("wetvr.com scenes require the scene_id argument (the numeric scene id, excluding site prefix)")
		}
		info, _ := json.Marshal([]api.RequestSingleScrapeAdditionInfo{{
			FieldName:   "scene_id",
			FieldPrompt: "Scene Id",
			Placeholder: "eg 69037",
			FieldValue:  args.SceneID,
			Required:    true,
			Type:        "number",
		}})
		additionalInfo = string(info)
	}

	if models.CheckLock("scrape") {
		return nil, nil, fmt.Errorf("a scrape is already running")
	}

	scene := tasks.ScrapeSingleScene(site, args.URL, additionalInfo)
	if scene.ID == 0 {
		return mcpTextResult("null"), nil, nil
	}

	// Confirm the scraped scene the same way the EditScene modal's
	// "Save Scene Details" button would.
	if _, err := api.ApplySceneEdit(scene.ID, mcpBuildEditRequest(scene)); err != nil {
		return nil, nil, fmt.Errorf("scene was scraped but saving it failed: %v", err)
	}

	return mcpTextResult(scene.SceneID), nil, nil
}
