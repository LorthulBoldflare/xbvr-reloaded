package api

import (
	"encoding/json"
	"html"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/tidwall/gjson"

	"github.com/xbapps/xbvr/pkg/common"
	"github.com/xbapps/xbvr/pkg/models"
	"github.com/xbapps/xbvr/pkg/scrape"
)

// trailerHTTPClient is shared by the trailer fetchers; the timeout prevents
// a hung upstream from tying up a request handler forever, and the
// SSRF-safe transport re-validates redirect targets against the public-IP
// rules.
var trailerHTTPClient = &http.Client{Timeout: 30 * time.Second, Transport: common.SSRFSafeTransport{}}

func LoadHeresphereScene(scrapeParams string) HeresphereVideo {
	var params models.TrailerScrape
	if strings.HasPrefix(scrapeParams, "http") {
		// keep backwards compatible with just url
		params.SceneUrl = scrapeParams
	} else {
		json.Unmarshal([]byte(scrapeParams), &params)
	}

	req, err := http.NewRequest(http.MethodPost, params.SceneUrl, nil)
	if err != nil {
		log.Errorf("Error building request for %s: %s", params.SceneUrl, err)
		return HeresphereVideo{}
	}

	if err := common.ValidateOutboundURL(params.SceneUrl); err != nil {
		log.Warnf("Rejected trailer URL %q: %v", params.SceneUrl, err)
		return HeresphereVideo{}
	}

	if params.KVHttpConfig == "" {
		params.KVHttpConfig = scrape.GetCoreDomain(params.SceneUrl) + "-trailers"
	}
	log.Debugf("Using Header/Cookies from %s", params.KVHttpConfig)
	scrape.SetupHtmlRequest(params.KVHttpConfig, req)
	response, err := trailerHTTPClient.Do(req)

	if err != nil {
		return HeresphereVideo{}
	}
	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Errorf("Error from %s %s", params.SceneUrl, err)
	}

	var video HeresphereVideo
	err = json.Unmarshal(responseData, &video)
	if err != nil {
		log.Errorf("Error from %s %s", params.SceneUrl, err)
	}

	return video
}

func LoadDeovrScene(scrapeParams string) DeoScene {
	var params models.TrailerScrape
	if strings.HasPrefix(scrapeParams, "http") {
		// keep backwards compatible with just url
		params.SceneUrl = scrapeParams
	} else {
		json.Unmarshal([]byte(scrapeParams), &params)
	}

	req, err := http.NewRequest(http.MethodGet, params.SceneUrl, nil)
	if err != nil {
		log.Errorf("Error building request for %s: %s", params.SceneUrl, err)
		return DeoScene{}
	}

	if err := common.ValidateOutboundURL(params.SceneUrl); err != nil {
		log.Warnf("Rejected trailer URL %q: %v", params.SceneUrl, err)
		return DeoScene{}
	}

	if params.KVHttpConfig == "" {
		params.KVHttpConfig = scrape.GetCoreDomain(params.SceneUrl) + "-trailers"
	}
	log.Debugf("Using Header/Cookies from %s", params.KVHttpConfig)
	scrape.SetupHtmlRequest(params.KVHttpConfig, req)

	response, err := trailerHTTPClient.Do(req)

	if err != nil {
		return DeoScene{}
	}
	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Errorf("Error from %s %s", params.SceneUrl, err)
	}

	var video DeoScene
	err = json.Unmarshal(responseData, &video)
	if err != nil {
		log.Errorf("Error from %s %s response: %s", params.SceneUrl, err, string(responseData))
	}

	return video
}

func ScrapeHtml(scrapeParams string) models.VideoSourceResponse {
	c := colly.NewCollector(colly.UserAgent(scrape.UserAgent))
	var params models.TrailerScrape
	json.Unmarshal([]byte(scrapeParams), &params)
	if err := common.ValidateOutboundURL(params.SceneUrl); err != nil {
		log.Warnf("Rejected trailer URL %q: %v", params.SceneUrl, err)
		return models.VideoSourceResponse{}
	}

	if params.KVHttpConfig == "" {
		params.KVHttpConfig = scrape.GetCoreDomain(params.SceneUrl) + "-trailers"
	}
	log.Debugf("Using Header/Cookies from %s", params.KVHttpConfig)
	scrape.SetupCollector(params.KVHttpConfig, c)

	var srcs []models.VideoSource
	c.OnHTML(`html`, func(e *colly.HTMLElement) {
		e.ForEach(params.HtmlElement, func(id int, e *colly.HTMLElement) {
			if params.ExtractRegex == "" {
				origURLtmp := e.Attr(params.ContentPath)
				quality := e.Attr(params.QualityPath)
				if origURLtmp != "" {
					if params.ContentBaseUrl != "" && !strings.HasPrefix(origURLtmp, params.ContentBaseUrl) {
						origURLtmp = params.ContentBaseUrl + origURLtmp
					}
					srcs = append(srcs, models.VideoSource{URL: origURLtmp, Quality: quality})
				}
			} else {
				//  extract match with regex expression if one was specified
				re := regexp.MustCompile(params.ExtractRegex)
				results := re.FindAllStringSubmatch(e.Text, -1)
				for _, result := range results {
					if len(result) < 2 {
						continue
					}
					parsedURL, _ := url.Parse(result[0])
					filename := path.Base(parsedURL.Path)
					baseFilename := strings.TrimSuffix(filename, path.Ext(filename))
					srcs = append(srcs, models.VideoSource{URL: result[1], Quality: baseFilename})
				}
			}
		})
	})
	c.Visit(params.SceneUrl)

	r := models.VideoSourceResponse{
		VideoSources: srcs,
	}
	return r
}

func ScrapeJson(scrapeParams string) models.VideoSourceResponse {
	c := colly.NewCollector(colly.UserAgent(scrape.UserAgent))
	var params models.TrailerScrape
	json.Unmarshal([]byte(scrapeParams), &params)
	if err := common.ValidateOutboundURL(params.SceneUrl); err != nil {
		log.Warnf("Rejected trailer URL %q: %v", params.SceneUrl, err)
		return models.VideoSourceResponse{}
	}

	if params.KVHttpConfig == "" {
		params.KVHttpConfig = scrape.GetCoreDomain(params.SceneUrl) + "-trailers"
	}
	log.Debugf("Using Header/Cookies from %s", params.KVHttpConfig)
	scrape.SetupCollector(params.KVHttpConfig, c)
	var srcs []models.VideoSource
	c.OnHTML(`html`, func(e *colly.HTMLElement) {
		e.ForEach(params.HtmlElement, func(id int, e *colly.HTMLElement) {
			txt := e.Text
			//  extract json with regex expression if one was specified
			if params.ExtractRegex != "" {
				re := regexp.MustCompile(params.ExtractRegex)
				r := re.FindStringSubmatch(txt)
				if len(r) > 0 {
					if r[1] != "" {
						txt = r[1]
					}
				}
			}

			srcs = extractFromJson(txt, params, srcs)
		})
	})
	c.Visit(params.SceneUrl)

	r := models.VideoSourceResponse{
		VideoSources: srcs,
	}
	return r
}

func LoadJson(scrapeParams string) models.VideoSourceResponse {
	var params models.TrailerScrape
	json.Unmarshal([]byte(scrapeParams), &params)

	req, err := http.NewRequest(http.MethodGet, params.SceneUrl, nil)
	if err != nil {
		log.Errorf("Error building request for %s: %s", params.SceneUrl, err)
		return models.VideoSourceResponse{}
	}

	if err := common.ValidateOutboundURL(params.SceneUrl); err != nil {
		log.Warnf("Rejected trailer URL %q: %v", params.SceneUrl, err)
		return models.VideoSourceResponse{}
	}

	if params.KVHttpConfig == "" {
		params.KVHttpConfig = scrape.GetCoreDomain(params.SceneUrl) + "-trailers"
	}
	log.Debugf("Using Header/Cookies from %s", params.KVHttpConfig)
	scrape.SetupHtmlRequest(params.KVHttpConfig, req)
	response, err := trailerHTTPClient.Do(req)

	if err != nil {
		return models.VideoSourceResponse{}
	}
	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Errorf("Error from %s %s", params.SceneUrl, err)
	}

	var resp models.VideoSourceResponse
	resp.VideoSources = extractFromJson(string(responseData), params, []models.VideoSource{})

	return resp
}

func extractFromJson(inputJson string, params models.TrailerScrape, srcs []models.VideoSource) []models.VideoSource {
	JsonMetadata := strings.TrimSpace(inputJson)

	// if the path to a record exists loop through each video source
	if params.RecordPath != "" {
		u := gjson.Get(JsonMetadata, params.RecordPath)
		u.ForEach(func(key, value gjson.Result) bool {
			url := params.ContentBaseUrl
			url += gjson.Get(value.String(), params.ContentPath).String()
			quality := gjson.Get(value.String(), params.QualityPath).String()
			encoding := ""
			if params.EncodingPath != "" {
				encoding = gjson.Get(value.String(), params.EncodingPath).String() + "-"
			}

			if url != "" {
				srcs = append(srcs, models.VideoSource{URL: url, Quality: encoding + quality})
			}
			return true // keep iterating
		})
	} else {
		// get single entry, ie not a repeating group of video sources
		if gjson.Get(JsonMetadata, params.ContentPath).String() != "" {
			quality := strings.TrimSpace(html.UnescapeString(gjson.Get(JsonMetadata, params.QualityPath).String()))
			url := strings.TrimSpace(html.UnescapeString(gjson.Get(JsonMetadata, params.ContentPath).String()))
			encoding := ""
			if params.EncodingPath != "" {
				encoding = gjson.Get(JsonMetadata, params.EncodingPath).String() + "-"
			}
			if params.ContentBaseUrl != "" && !strings.HasPrefix(url, params.ContentBaseUrl) {
				if params.ContentBaseUrl[len(params.ContentBaseUrl)-1:] == "/" && string(url[0]) == "/" {
					url = params.ContentBaseUrl + url[1:]
				} else {
					url = params.ContentBaseUrl + url
				}
			}
			srcs = append(srcs, models.VideoSource{URL: url, Quality: encoding + quality})
		}
	}
	return srcs
}

func LoadUrl(url string) models.VideoSourceResponse {
	var srcs []models.VideoSource
	srcs = append(srcs, models.VideoSource{URL: url, Quality: "Unknown"})

	r := models.VideoSourceResponse{
		VideoSources: srcs,
	}
	return r
}
