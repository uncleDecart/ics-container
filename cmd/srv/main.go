package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/uncleDecart/ics-container/pkg/pemanager"
	"github.com/uncleDecart/ics-container/pkg/recaster"
)

func main() {
	r := chi.NewRouter()

	addr := "http://127.0.0.1:8889"
	peMgr, err := pemanager.NewPatchEnvelopeManager(addr)
	if err != nil {
		fmt.Println("Error : ", err)
		return
	}
	peMgr.Fetch()

	name := "bazinga"

	cfg := recaster.Config{
		Name: name,
	}

	recaster := &recaster.RecasterManager{
		PeMgr: peMgr,
	}

	recaster.Put(cfg)

	r.Get("/static/index", RenderRecasterIndexPage(recaster))
	r.Get("/static/edit", RenderRecasterEditPage(recaster))

	r.Post("/edit", HandleRecasterEdit(recaster))
	r.Post("/render", HandleRecasterTransform(recaster))

	http.ListenAndServe(":3333", r)

	//transformed, err := recaster.Transform("patch1/config.yml/data")
	//if err != nil {
	//	fmt.Println("FAILED ", err)
	//	return
	//}

	//fmt.Println(transformed)
}
