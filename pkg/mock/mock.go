package mock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	c "github.com/uncleDecart/ics-api/go/client"
)

func CreateListener() (l net.Listener, close func()) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	return l, func() {
		_ = l.Close()
	}
}

func sendError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(code)
	fmt.Fprintf(w, "{\"message}\": \"%s\"", msg)
}

func mockDescription(desc *[]c.PatchEnvelopeDescription) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(desc)
	}
}

func mockDownload(desc *map[string]string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		patchID := chi.URLParam(r, "patch")
		if patchID == "" {
			sendError(w, http.StatusNotFound, "patch is empty")
			return
		}
		fileName := chi.URLParam(r, "file")
		if fileName == "" {
			sendError(w, http.StatusNotFound, "file is empty")
			return
		}

		data, ok := (*desc)[patchID+":"+fileName]
		if !ok {
			sendError(w, http.StatusNotFound, "desc is empty")
			return
		}

		fmt.Fprintf(w, data)
	}
}

func CreateMockMetadataServer(envelopesAvailable *[]c.PatchEnvelopeDescription, data *map[string]string) *http.Server {
	r := chi.NewRouter()
	r.Get("/patch/description.json", mockDescription(envelopesAvailable))
	r.Get("/patch/download/{patch}/{file}", mockDownload(data))

	return &http.Server{
		Handler: r,
	}
}

type MockConsumer struct {
	Data map[string]string
}

func (mc *MockConsumer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	endpoint := chi.URLParam(r, "endpoint")
	if endpoint == "" {
		sendError(w, http.StatusNotFound, "no endpoint")
		return
	}

	buf := &bytes.Buffer{}
	io.Copy(buf, r.Body)
	defer r.Body.Close()

	fmt.Printf("Saving %s with %s\n", endpoint, buf.String())

	mc.Data[endpoint] = buf.String()

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func CreateMockConsumer() (*MockConsumer, *http.Server) {
	mc := &MockConsumer{
		Data: make(map[string]string),
	}

	r := chi.NewRouter()
	r.Post("/{endpoint}", mc.ServeHTTP)
	r.Get("/{endpoint}", mc.ServeHTTP)

	return mc, &http.Server{
		Handler: r,
	}
}
