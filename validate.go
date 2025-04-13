package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func handleValidate(w http.ResponseWriter, r *http.Request) {

	type valid struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)

	isValid := valid{}

	err := decoder.Decode(&isValid)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	type returnVal struct {
		Valid bool `json:"valid"`
	}

	type returnErr struct {
		Error string `json:"error"`
	}

	if len(isValid.Body) <= 140 {

		returnVal := returnVal{
			Valid: true,
		}

		data, err := json.Marshal(returnVal)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(data)
	} else {

		returnErr := returnErr{
			Error: "Length is too long",
		}

		data, err := json.Marshal(returnErr)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(data)
	}

}
