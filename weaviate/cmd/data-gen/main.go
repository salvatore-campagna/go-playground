package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
)

const (
	numSegments           = 7
	numDocsPerSegment     = 100_000
	maxDocID              = 1_000_000
	defaultJsonOutputFIle = "output.json"
)

type JsonDocument struct {
	Term          string  `json:"term"`
	DocID         uint32  `json:"doc_id"`
	TermFrequency float32 `json:"term_frequency"`
}

type JsonSegment struct {
	Documents []JsonDocument `json:"documents"`
}

type Root struct {
	Segments [][]JsonDocument `json:"segments"`
}

var vocabulary = []string{
	"jedi", "force", "skywalker", "sith", "lightsaber", "empire", "rebellion", "droid",
	"blaster", "starship", "yoda", "clone", "trooper", "battle", "padawan", "hologram",
	"bounty", "hunter", "coruscant", "tatooine", "deathstar", "vader", "han", "chewbacca",
	"leia", "luke", "anakin", "grievous", "obiwan", "qui-gon", "naboo", "geonosis",
	"kamino", "mustafar", "dagobah", "endor", "hoth", "alderaan", "kashyyyk", "lando",
	"carbonite", "lightspeed", "hyperdrive", "holocron", "starfighter", "speeder", "cantina",
	"protocol", "gungan", "wookiee",
}

// generateRandomDocument generates a single document with random values
func generateRandomDocument(term string, docID uint32) JsonDocument {
	return JsonDocument{
		Term:          term,
		DocID:         docID,
		TermFrequency: rand.Float32(), // Random term frequency between 0.0 and 1.0
	}
}

// generateSegment generates a segment containing a list of documents
func generateSegment(segmentID int) []JsonDocument {
	segment := []JsonDocument{}
	var docID uint32

	for i := 0; i < numDocsPerSegment; i++ {
		term := vocabulary[rand.Intn(len(vocabulary))]
		doc := generateRandomDocument(term, docID)
		segment = append(segment, doc)
		docID++
	}

	return segment
}

// generateSegments generates a JSON file with multiple segments
func generateSegments() Root {
	root := Root{
		Segments: make([][]JsonDocument, numSegments),
	}

	for i := 0; i < numSegments; i++ {
		root.Segments[i] = generateSegment(i)
	}

	return root
}

// writeJsonToFile writes the generated segments to a JSON file
func writeJsonToFile(root Root, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(root); err != nil {
		return fmt.Errorf("failed to write JSON to file: %w", err)
	}

	return nil
}

func main() {
	jsonOutputFile, exists := os.LookupEnv("JSON_OUTPUT_FILE")
	if !exists {
		jsonOutputFile = defaultJsonOutputFIle
	}

	fmt.Printf("Writing file: %s\n", jsonOutputFile)
	err := writeJsonToFile(generateSegments(), jsonOutputFile)
	if err != nil {
		fmt.Printf("Error writing JSON to file: %v\n", err)
		return
	}

	fmt.Printf("JSON file generated successfully: %s\n", jsonOutputFile)
}
