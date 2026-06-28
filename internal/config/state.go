package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type State struct {
	LastServerIdx  int    `json:"lastServerIdx"`
	LastServerName string `json:"lastServerName"`
}

func StatePath() string {
	return filepath.Join(CacheDir(), "state.json")
}

func LoadState() (*State, error) {
	s := &State{LastServerIdx: -1}
	data, err := os.ReadFile(StatePath())
	if err != nil {
		return s, err
	}
	return s, json.Unmarshal(data, s)
}

func SaveState(s *State) error {
	if err := os.MkdirAll(CacheDir(), 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(StatePath(), data, 0o600)
}
