package frontend

import (
	"GoInNuvola/core"
	"errors"
	"github.com/gorilla/mux"
	"io"
	"net/http"
)

type RestFrontend struct {
	Store *core.KeyValueStore
}

func (f *RestFrontend) Start(store *core.KeyValueStore) error {
	f.Store = store
	r := mux.NewRouter()
	r.HandleFunc("/v1/key/{key}", f.KeyValuePutHandler).Methods("PUT")
	r.HandleFunc("/v1/key/{key}", f.KeyValueGetHandler).Methods("GET")
	r.HandleFunc("/v1/key/{key}", f.KeyValueDeleteHandler).Methods("DELETE")

	return http.ListenAndServe(":8080", r)

}

func (f *RestFrontend) KeyValuePutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = f.Store.Put(key, string(value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	f.Store.Transact.WritePut(key, string(value))

	w.WriteHeader(http.StatusCreated)
}

func (f *RestFrontend) KeyValueGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := f.Store.Get(key)
	if errors.Is(err, core.ErrorNoSuchKey) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(value))
}

func (f *RestFrontend) KeyValueDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	err := f.Store.Delete(key)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	f.Store.Transact.WriteDelete(key)
	w.WriteHeader(http.StatusOK)
}
