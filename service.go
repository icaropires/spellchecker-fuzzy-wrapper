/*
Spellchecker binary inteded to be used as a standalone binary.
Not using files for making easier the use thorugh another source code written in another language.
*/
package main

import (
	"net/http"
	//	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/sajari/fuzzy"
)

type Request struct {
	id        int
	toCorrect string
}

type Checker struct {
	last       int
	requests   chan Request
	results    map[int]string
	model      *fuzzy.Model
	muxRequest sync.Mutex
	muxResult  sync.Mutex
}

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

func (c *Checker) check(s string) string {
	tokens := strings.Split(s, "\n")

	fmt.Printf("[INFO] Total of words to be corrected = %d\n", len(tokens))

	result := ""
	result += "{" // Open json
	for i, token := range tokens {
		var aux string
		fmt.Sprintf(aux, "\"%s\": \"%s\"", token, c.model.SpellCheck(token)) // Json key: value

		if i != len(tokens)-1 {
			aux += ",\n"
		}

		result += aux
	}
	result += "}" // Close json

	return result
}

func (c *Checker) get(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hey get")
}

func (c *Checker) post(w http.ResponseWriter, r *http.Request) {
	request := Request{}

	body, _ := ioutil.ReadAll(r.Body)
	request.toCorrect = string(body)

	if request.toCorrect == "" {
		log.Printf("Invalid request received from \"%s\"", r.RemoteAddr)
		http.Error(w, "Not a valid string received", http.StatusBadRequest)

		return
	}

	c.muxRequest.Lock()
	c.last++
	request.id = c.last
	c.muxRequest.Unlock()

	c.requests <- request

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Your request \"%d\" will be processed. Soon get your result at \"%s/%d\"", request.id, r.Host, request.id)

	log.Printf("Received request %d from %s", request.id, r.RemoteAddr)
}

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "train" {
		if len(os.Args) < 3 {
			log.Fatal("Wrong Usage. Example: ./spellchecker train [depth of search] [vocabulary]")
		}

		trainModel()
		os.Exit(0)
	}

	fmt.Println("[INFO] Loading Model...")
	model, err := fuzzy.Load(ModelFile)
	if err != nil {
		log.Fatal("Model not trained. Train it first with: ./spellchecker train [depth of search] [vocabulary]")
	}
	fmt.Println("[INFO] Model loaded!")

	checker := Checker{}
	checker.model = model
	checker.requests = make(chan Request, 10)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "404 not found.", http.StatusNotFound)
			return
		}

		switch r.Method {
		case "GET":
			checker.get(w, r)
		case "POST":
			checker.post(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
