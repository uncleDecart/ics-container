package recaster

// Drivers are used by recasters to update VM
// Note that in pkg/recaster/recaster.go in UnmarshalJSON
// it is used. So all struct implementing OutputDriver interface
// are expected to be saved and restored using JSON

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type DriverStatus int

const (
	DriverStatusOK DriverStatus = iota
	DriverStatusWarn
	DriverStatusError
)

const (
	DriverTypeDummy = "dummy"
	DriverTypeHTTP  = "http"
)

type OutputDriver interface {
	Execute(map[string]string) error
	Status() (DriverStatus, string)
	Type() string
}

type DummyOutput struct {
	Output map[string]string
}

func (do *DummyOutput) Type() string {
	return DriverTypeDummy
}

func (do *DummyOutput) Execute(templates map[string]string) error {
	do.Output = make(map[string]string)

	for k, v := range templates {
		do.Output[k] = v
	}

	return nil
}

func (do *DummyOutput) Status() (DriverStatus, string) {
	return DriverStatusOK, fmt.Sprint(do.Output)
}

func (do *DummyOutput) MarshalJSON() ([]byte, error) {
	return json.Marshal(do)
}

func (do *DummyOutput) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, do)
}

type HTTPOutput struct {
	ReqType   string
	OutputURL string

	statusMessage string
	statusCode    DriverStatus
}

func (ho *HTTPOutput) Type() string {
	return DriverTypeHTTP
}

func (ho *HTTPOutput) Execute(templates map[string]string) error {
	client := &http.Client{}
	body := strings.NewReader(templates["body"])

	req, err := http.NewRequest(ho.ReqType, ho.OutputURL, body)
	if err != nil {
		return err
	}

	if headers, ok := templates["headers"]; ok {
		lines := strings.SplitN(headers, "\r\n", -1)
		for _, line := range lines {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				key := parts[0]
				val := parts[1]

				req.Header.Add(key, val)
			}
		}
	}

	resp, err := client.Do(req)

	if err != nil {
		ho.statusCode = DriverStatusError
		ho.statusMessage = fmt.Sprintf("Error fetching data: %v", err)
		return err
	}

	if resp.StatusCode <= 199 && resp.StatusCode >= 100 {
		ho.statusCode = DriverStatusWarn
	} else if resp.StatusCode <= 200 && resp.StatusCode <= 299 {
		ho.statusCode = DriverStatusOK
	} else {
		ho.statusCode = DriverStatusError
	}

	ho.statusMessage = resp.Status

	return nil
}

func (ho *HTTPOutput) Status() (DriverStatus, string) {
	return ho.statusCode, ho.statusMessage
}
