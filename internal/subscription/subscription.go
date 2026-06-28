package subscription

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Server struct {
	Remarks string          `json:"remarks"`
	Config  json.RawMessage `json:"config"`
}

type Meta struct {
	Title          string
	Announce       string
	Upload         int64
	Download       int64
	Total          int64
	Expire         int64
	UpdateInterval int
}

type Group struct {
	URL     string   `json:"url"`
	Meta    Meta     `json:"meta"`
	Servers []Server `json:"servers"`
}

type Fetcher struct{ client *http.Client }

func NewFetcher(c *http.Client) *Fetcher {
	if c == nil {
		c = http.DefaultClient
	}
	return &Fetcher{client: c}
}

func (f *Fetcher) Fetch(url string) ([]Server, Meta, error) {
	resp, err := f.client.Get(url)
	if err != nil {
		return nil, Meta{}, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, Meta{}, err
	}
	meta := parseMeta(resp.Header)
	servers, err := parse(body)
	return servers, meta, err
}

func parseMeta(h http.Header) Meta {
	var m Meta
	if ui := h.Get("subscription-userinfo"); ui != "" {
		for _, part := range strings.Split(ui, ";") {
			kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
			if len(kv) != 2 {
				continue
			}
			val, err := strconv.ParseInt(strings.TrimSpace(kv[1]), 10, 64)
			if err != nil {
				continue
			}
			switch strings.TrimSpace(kv[0]) {
			case "upload":
				m.Upload = val
			case "download":
				m.Download = val
			case "total":
				m.Total = val
			case "expire":
				m.Expire = val
			}
		}
	}
	if pt := h.Get("profile-title"); pt != "" {
		if strings.HasPrefix(pt, "base64:") {
			decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(pt, "base64:"))
			if err == nil {
				m.Title = string(decoded)
			}
		} else {
			m.Title = pt
		}
	}
	if pi := h.Get("profile-update-interval"); pi != "" {
		if v, err := strconv.Atoi(pi); err == nil {
			m.UpdateInterval = v
		}
	}
	if ann := h.Get("announce"); ann != "" {
		if strings.HasPrefix(ann, "base64:") {
			decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(ann, "base64:"))
			if err == nil {
				m.Announce = string(decoded)
			}
		} else {
			m.Announce = ann
		}
	}
	return m
}

func parse(body []byte) ([]Server, error) {
	trimmed := strings.TrimSpace(string(body))
	if strings.HasPrefix(trimmed, "[") {
		return parseJSONArray([]byte(trimmed))
	}
	decoded, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(trimmed)
		if err != nil {
			return nil, fmt.Errorf("unrecognized subscription format")
		}
	}
	return parseURIList(string(decoded))
}

func parseJSONArray(data []byte) ([]Server, error) {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse JSON array: %w", err)
	}
	out := make([]Server, 0, len(raw))
	for _, r := range raw {
		var meta struct {
			Remarks string `json:"remarks"`
		}
		if err := json.Unmarshal(r, &meta); err != nil {
			continue
		}
		out = append(out, Server{Remarks: meta.Remarks, Config: r})
	}
	return out, nil
}

func parseURIList(text string) ([]Server, error) {
	var out []Server
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		s, err := ParseURI(line)
		if err != nil {
			continue
		}
		out = append(out, s)
	}
	return out, nil
}

func CacheServers(servers []Server, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(servers)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func LoadCachedServers(path string) ([]Server, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s []Server
	return s, json.Unmarshal(data, &s)
}

func CacheGroups(groups []Group, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(groups)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func LoadCachedGroups(path string) ([]Group, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var g []Group
	return g, json.Unmarshal(data, &g)
}
