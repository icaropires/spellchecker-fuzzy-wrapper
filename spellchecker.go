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

func trainModel() {
	const (
		_ = iota + 1
		DepthArgsIdx
		VocabularyArgsIdx
	)

	vocabularyTxt, err := ioutil.ReadFile(os.Args[VocabularyArgsIdx])
	if err != nil {
		log.Fatal(err)
	}
	vocabulary := strings.Split(string(vocabularyTxt), "\n")

	model := fuzzy.NewModel()
	model.SetThreshold(1)

	depth, err := strconv.Atoi(os.Args[DepthArgsIdx])
	if err != nil {
		log.Fatal(err)
	}
	model.SetDepth(depth)

	fmt.Println("[INFO] Training model...")
	model.Train(vocabulary)
	fmt.Println("[INFO] Model trained!")

	fmt.Println("[INFO] Saving model...")
	model.SaveLight(ModelFile)
	fmt.Printf("[INFO] Model saved to '%s'!\n", ModelFile)
}

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "train" {
		if len(os.Args) < 3 {
			log.Fatal("Wrong Usage. Example: ./spellchecker train [depth of search] [vocabulary]")
		}

		trainModel()
		os.Exit(0)
	} else if len(os.Args) < 2 {
		log.Fatal("Wrong Usage.\nUsage syntax: ./spellchecker [tokens file]")
	}

	fmt.Println("[INFO] Loading Model...")

	model, err := fuzzy.Load(ModelFile)
	if err != nil {
		log.Fatal("Model not trained. Train it first with: ./spellchecker train [depth of search] [vocabulary]")
	}

	TokensArgsIdx := 1
	tokensTxt, err := ioutil.ReadFile(os.Args[TokensArgsIdx])
	if err != nil {
		log.Fatal(err)
	}
	tokens := strings.Split(string(tokensTxt), "\n")

	fmt.Printf("[INFO] Total of words to be corrected = %d\n", len(tokens))

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
