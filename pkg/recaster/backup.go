package recaster

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/uncleDecart/ics-container/pkg/utils"
)

const (
	BackupTypeNoBackup = "nobackup"
	BackupTypeHTTP     = "http"
)

type BackupStrategy interface {
	Load(params map[string]string) error
	Restore(params map[string]string) error
	Enabled() bool

	Params() map[string]string
	Data() string
	Type() string
}

type NoBackup struct{}

func (nb *NoBackup) Load(map[string]string) error    { return nil }
func (nb *NoBackup) Restore(map[string]string) error { return nil }
func (nb *NoBackup) Enabled() bool                   { return false }
func (nb *NoBackup) Type() string                    { return BackupTypeNoBackup }

func (nb *NoBackup) Params() map[string]string { return nil }
func (nb *NoBackup) Data() string              { return "" }

type HTTPBackup struct {
	LoadType string
	LoadURL  string

	RestoreType string
	RestoreURL  string

	Parameters map[string]string

	data []byte
}

func (hb *HTTPBackup) Enabled() bool             { return true }
func (hb *HTTPBackup) Params() map[string]string { return hb.Parameters }
func (hb *HTTPBackup) Data() string              { return string(hb.data) }
func (hb *HTTPBackup) Type() string              { return BackupTypeHTTP }

func (hb *HTTPBackup) Load(params map[string]string) error {
	var body io.Reader
	if bodyParam, ok := params["load_body"]; ok {
		body = strings.NewReader(bodyParam)
	}
	hdrs := params["load_headers"]
	u, err := url.Parse(hb.LoadURL)
	if err != nil {
		return err
	}

	resp, err := utils.DoHTTPRequest(hb.LoadType, *u, body, hdrs)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	hb.data, err = io.ReadAll(resp.Body)

	return err
}

func (hb *HTTPBackup) Restore(params map[string]string) error {
	hdrs := params["restore_headers"]
	body := bytes.NewReader(hb.data)
	u, err := url.Parse(hb.RestoreURL)

	resp, err := utils.DoHTTPRequest(hb.RestoreType, *u, body, hdrs)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return nil
	}

	return fmt.Errorf("Response code is not successful %d", resp.StatusCode)
}
