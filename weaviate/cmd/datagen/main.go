package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"weaviate/fetcher"
)

const (
	numSegments           = 7
	numDocsPerSegment     = 100_000
	maxDocID              = 1_000_000
	defaultJsonOutputFile = "output.json"
)

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
func generateRandomDocument(term string, docID uint32) fetcher.TermPosting {
	return fetcher.TermPosting{
		Term:          term,
		DocID:         docID,
		TermFrequency: rand.Float32(), // Random term frequency between 0.0 and 1.0
	}
}

// generateSegment generates a segment containing a list of documents
func generateSegment(segmentID int) []fetcher.TermPosting {
	segment := []fetcher.TermPosting{}
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
func generateSegments() fetcher.TermPostingRoot {
	root := fetcher.TermPostingRoot{
		Segments: make([][]fetcher.TermPosting, numSegments),
	}

	for i := 0; i < numSegments; i++ {
		root.Segments[i] = generateSegment(i)
	}

	return root
}

// writeJsonToFile writes the generated segments to a JSON file
func writeJsonToFile(root fetcher.TermPostingRoot, filename string) error {
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

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
	path := flag.String("path", defaultJsonOutputFile, "Output JSON file path")
	flag.Parse()

	fmt.Printf("Writing file: %s\n", *path)
	err := writeJsonToFile(generateSegments(), *path)
	if err != nil {
		fmt.Printf("Error writing JSON to file: %v\n", err)
		return
	}

	fmt.Printf("JSON file generated successfully: %s\n", *path)
}
