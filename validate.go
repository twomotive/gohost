package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
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
		Valid        bool   `json:"valid"`
		Cleaned_body string `json:"cleaned_body"`
	}

	type returnErr struct {
		Error string `json:"error"`
	}

	if len(isValid.Body) <= 140 {

		text, err := removeBadWords(isValid.Body)
		if err != nil {
			log.Printf("Error removing words from JSON: %s", err)
			w.WriteHeader(500)
			return
		}

		returnVal := returnVal{
			Valid:        true,
			Cleaned_body: text,
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

func removeBadWords(s string) (string, error) {
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Fields(s)

	var result strings.Builder

	for i, word := range words {
		if i > 0 {
			result.WriteString(" ")
		}

		wordLower := strings.ToLower(word)
		if _, exists := badWords[wordLower]; exists {
			result.WriteString("****")
		} else {
			result.WriteString(word)
		}
	}

	return result.String(), nil
}
