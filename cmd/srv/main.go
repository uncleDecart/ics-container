package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/uncleDecart/ics-container/pkg/pemanager"
	"github.com/uncleDecart/ics-container/pkg/recaster"
)

func main() {
	r := chi.NewRouter()

	configPath := flag.String("cfg", "config.json", "path to configuration file")

	data, err := os.ReadFile(*configPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	mngr := &recaster.RecasterManager{}

	err = json.Unmarshal(data, mngr)
	if err != nil {
		fmt.Println(err)
		return
	}

	mngr.ConfigPath = *configPath

	peMngr, err := pemanager.NewPatchEnvelopeManager(mngr.PeServerURL)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = peMngr.Fetch()
	if err != nil {
		fmt.Println(err)
		return
	}

	mngr.PeMgr = peMngr

	err = mngr.Update()
	if err != nil {
		fmt.Println(err)
		return
	}

	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	r.Get("/index", RenderRecasterIndexPage(mngr))
	r.Get("/edit", RenderRecasterEditPage(mngr))

	r.Route("/api/", func(r chi.Router) {
		r.Get("/recasters", HandleRecasterGet(mngr))

		r.Post("/recaster", HandleRecasterEdit(mngr))
		r.Delete("/recaster", HandleRecasterDelete(mngr))

		r.Get("/envelopes", HandlePatchEnvelopesGet(mngr))

		r.Post("/render", HandleRecasterTransform(mngr))

		r.Get("/status", HandleRecastersStatus(mngr))
	})

	if err := http.ListenAndServe(":8080", r); err != nil {
		fmt.Println(err)
	}
}
