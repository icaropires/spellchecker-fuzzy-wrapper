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
	"strings"

	"github.com/sajari/fuzzy"
)

func main() {
	const TokensArgsIdx int = 1
	const CorpusArgsIdx int = 2

	if len(os.Args) < 3 {
		log.Fatal("Wrong Usage.\nUsage syntax: ./spellchecker [tokens_string] [corpus_string]\nSeparate words with spaces")
	}

	corpus := strings.Split(os.Args[CorpusArgsIdx], " ")

	model := fuzzy.NewModel()

	model.SetThreshold(1)
	model.SetDepth(5)

	model.Train(corpus)

	tokens := strings.Split(os.Args[TokensArgsIdx], " ")

	// Print like json, for easy loading in python code
	corrections := make(map[string]string)
	for _, token := range tokens {
		corrections[token] = model.SpellCheck(token)
	}

	correctionsDict, err := json.Marshal(corrections)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(correctionsDict))
}
