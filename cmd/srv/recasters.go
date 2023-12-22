package main

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/uncleDecart/ics-container/pkg/pemanager"
	"github.com/uncleDecart/ics-container/pkg/recaster"
)

func renderStaticPage(w http.ResponseWriter, page string, view any) {
	tmpl, err := template.ParseFiles(filepath.Join("templates", page))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, view)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func RenderRecasterIndexPage(model *recaster.RecasterManager) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		renderStaticPage(w, "index.html", model.Recasters)
	}
}

func RenderRecasterEditPage(model *recaster.RecasterManager) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		selectedRecaster := r.URL.Query().Get("name")
		if selectedRecaster == "" {
			http.Error(w, "Recaster is not specified", http.StatusBadRequest)
			return
		}

		idx := model.GetIdx(selectedRecaster)
		if idx == -1 {
			http.Error(w, "Recaster not found", http.StatusNotFound)
			return
		}

		type view struct {
			Envelopes pemanager.PatchEnvelopeView
			Recaster  recaster.Recaster
		}

		renderStaticPage(w, "edit.html", view{
			Envelopes: model.PeMgr.View(),
			Recaster:  *model.Recasters[idx],
		})
	}
}

func HandleRecasterEdit(model *recaster.RecasterManager) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "Empty name", http.StatusNotFound)
			return
		}

		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	}
}

const specialSymbol = "$"

type transformRequest struct {
	Message string
	Config  recaster.Config
}

func HandleRecasterTransform(model *recaster.RecasterManager) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handling transformt")

		var trReq transformRequest

		err := json.NewDecoder(r.Body).Decode(&trReq)
		if err != nil {
			fmt.Println("Can't Decode JSON")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		trReq.Config.Driver = &recaster.DummpyOutput{}

		rec := recaster.NewRecaster(trReq.Config, model.PeMgr)
		fmt.Println(rec.Config)
		result, err := rec.TransformString(trReq.Message)

		if err != nil {
			fmt.Println("Failed to parse body")
			http.Error(w, fmt.Sprintf("Failed to parse body: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Write([]byte(result))
	}
}
