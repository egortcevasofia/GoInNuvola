package main

import (
	"GoInNuvola/core"
	"GoInNuvola/frontend"
	"GoInNuvola/transact"
	_ "github.com/lib/pq"
	"log"
)

func main() {

	tl, err := transact.NewTransactionLogger("postgres")
	if err != nil {
		log.Fatal(err)
	}
	store := core.NewKeValueStore(tl)
	err = store.Restore()
	if err != nil {
		return
	}

	fe, err := frontend.NewFrontend()
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(fe.Start(store))
}
