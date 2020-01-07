/*
Spellchecker binary inteded to be used as a standalone binary.
Not using files for making easier the use thorugh another source code written in another language.
*/
package main

import (
	//	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/sajari/fuzzy"
)

// ModelFile is the name of the file which is the spell model
const ModelFile string = "spell_model.json"

// Meaning of indexes from Args
const (
	_ = iota
	DepthArgsIdx
	TokensArgsIdx
	VocabularyArgsIdx
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Wrong Usage.\nUsage syntax: ./spellchecker [DepthArgsIdx] [tokens_file] [vocabulary_file]\nSeparate words with spaces")
	}

	vocabularyTxt, err := ioutil.ReadFile(os.Args[VocabularyArgsIdx])
	if err != nil {
		log.Fatal(err)
	}
	vocabulary := strings.Split(string(vocabularyTxt), "\n")

	model, err := fuzzy.Load(ModelFile)

	if err != nil {
		model = fuzzy.NewModel()

		model.SetThreshold(1)

		depth, err := strconv.Atoi(os.Args[DepthArgsIdx])
		if err != nil {
			log.Fatal(err)
		}
		model.SetDepth(depth)

		fmt.Println("[INFO] No model found. Training...")
		model.Train(vocabulary)
		fmt.Printf("[INFO] Model trained and saved to %s!\n", ModelFile)

		model.SaveLight(ModelFile)
	}

	tokensTxt, err := ioutil.ReadFile(os.Args[TokensArgsIdx])
	if err != nil {
		log.Fatal(err)
	}
	tokens := strings.Split(string(tokensTxt), "\n")

	fmt.Printf("{") // Open json
	// Print like json, for easy loading in python code
	for i, token := range tokens {
		fmt.Printf("\"%s\": \"%s\"", token, model.SpellCheck(token)) // Json key: value

		if i != len(tokens)-1 {
			fmt.Printf(",\n")
		}
	}
	fmt.Printf("}\n") // Close json
}
