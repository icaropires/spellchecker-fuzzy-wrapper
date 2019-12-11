/*
Spellchecker binary inteded to be used as a standalone binary.
Not using files for making easier the use thorugh another source code written in another language.
*/
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/sajari/fuzzy"
)

// WordsByRoutine is the number of words processed by each go routine
const WordsByRoutine int = 100

// Meaning of indexes from Args
const (
	_ = iota
	TokensArgsIdx
	CorpusArgsIdx
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Wrong Usage.\nUsage syntax: ./spellchecker [tokens_string] [corpus_string]\nSeparate words with spaces")
	}

	corpus := strings.Split(os.Args[CorpusArgsIdx], " ")

	model := fuzzy.NewModel()

	model.SetThreshold(1)
	model.SetDepth(5)

	model.Train(corpus)

	tokens := strings.Split(os.Args[TokensArgsIdx], " ")

	corrections := make(map[string]string)

	var wg sync.WaitGroup
	processWords := func(words []string) {
		defer wg.Done()

		// Print like json, for easy loading in python code
		for _, token := range tokens {
			corrections[token] = model.SpellCheck(token) // Side effects
		}
	}

	numRoutines := runtime.NumCPU()
	wordsByRoutine := len(tokens) / numRoutines
	if runtime.NumCPU() > len(tokens) {
		numRoutines = len(tokens)
		wordsByRoutine = 1
	}

	for i := 0; i < numRoutines; i++ {
		wg.Add(1)

		step := i * wordsByRoutine

		if i != numRoutines-1 {
			go processWords(tokens[step : step+wordsByRoutine])
		} else {
			go processWords(tokens[step:])
		}
	}

	wg.Wait()

	correctionsDict, err := json.Marshal(corrections)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(correctionsDict))
}
