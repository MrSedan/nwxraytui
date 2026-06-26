package subscription

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Server struct {
	Remarks string          `json:"remarks"`
	Config  json.RawMessage `json:"config"`
}

type Fetcher struct{ client *http.Client }

func NewFetcher(c *http.Client) *Fetcher {
	if c == nil {
		c = http.DefaultClient
	}
	return &Fetcher{client: c}
}

func (f *Fetcher) Fetch(url string) ([]Server, error) {
	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return parse(body)
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
