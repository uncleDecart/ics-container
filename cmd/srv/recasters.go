package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/uncleDecart/ics-container/pkg/pemanager"
	"github.com/uncleDecart/ics-container/pkg/recaster"
)

func renderStaticPage(w http.ResponseWriter, page string, view any) {
	tmpl, err := template.ParseFiles(filepath.Join("static/html", page))

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
		rec := recaster.Recaster{}

		selectedRecaster := r.URL.Query().Get("name")
		if selectedRecaster != "" {
			idx := model.GetIdx(selectedRecaster)
			if idx == -1 {
				http.Error(w, "Recaster not found", http.StatusNotFound)
				return
			}
			rec = model.Recasters[idx]
		}

		type view struct {
			Envelopes pemanager.PatchEnvelopeView
			Recaster  recaster.Recaster
		}

		renderStaticPage(w, "edit.html", view{
			Envelopes: model.PeMgr.View(),
			Recaster:  rec,
		})
	}
}

func HandlePatchEnvelopesGet(model *recaster.RecasterManager) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(model.PeMgr.View())
	}
}

func HandleRecasterGet(model *recaster.RecasterManager) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")

		if name == "" {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(&model.Recasters)
			return
		}

		idx := model.GetIdx(name)
		if idx == -1 {
			http.Error(w, "Recaster not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&model.Recasters[idx])
	}
}

func HandleRecasterEdit(model *recaster.RecasterManager) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var rec recaster.Recaster
		err := json.NewDecoder(r.Body).Decode(&rec)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = model.Put(rec)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rec)
	}
}

func HandleRecasterDelete(model *recaster.RecasterManager) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var rec recaster.Recaster
		err := json.NewDecoder(r.Body).Decode(&rec)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		deleted := model.Delete(rec)
		if !deleted {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rec)
	}
}

func HandleRecasterTransform(model *recaster.RecasterManager) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type transformRequest struct {
			Message string            `json:"message"`
			Rec     recaster.Recaster `json:"config"`
		}
		var trReq transformRequest

		err := json.NewDecoder(r.Body).Decode(&trReq)
		if err != nil {
			fmt.Println("Can't Decode JSON")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println("RecasterTransform ", trReq.Rec)

		result, err := trReq.Rec.TransformString(trReq.Message, model.PeMgr)

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

func HandleRecastersStatus(model *recaster.RecasterManager) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		status := model.Status()
		name := r.URL.Query().Get("name")

		if name == "" {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(&status)
			return
		}

		idx := -1
		for i, s := range status {
			if s.Name == name {
				idx = i
				break
			}
		}
		if idx == -1 {
			http.Error(w, "Recaster not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&status[idx])
	}
}
