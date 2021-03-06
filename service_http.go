/*
Spellchecker binary inteded to be used as a standalone binary.
Not using files for making easier the use thorugh another source code written in another language.
*/
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/sajari/fuzzy"
)

// Tasks which are words to correct
type Tasks struct {
	id        int
	toCorrect string
}

// Checker will hold information of the main flow the application
type Checker struct {
	last       int
	tasks      chan Tasks
	results    map[int]string
	model      *fuzzy.Model
	muxTasks   sync.Mutex
	muxResults sync.Mutex
}

// ModelFile is the name of the file which is the spell model
const ModelFile string = "spell_model.json"

const (
	maxJobsEnv     = "MAX_JOBS"
	portEnv        = "PORT"
	portDefault    = "8080"
	maxJobsDefault = 4
)

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

	log.Println("Training model...")
	model.Train(vocabulary)
	log.Println("Model trained!")

	log.Println("Saving model...")
	model.SaveLight(ModelFile)
	log.Printf("Model saved to '%s'!\n", ModelFile)
}

func (c *Checker) check(tokens []string) string {
	result := ""

	result += "{" // Open json
	for i, token := range tokens {
		result += fmt.Sprintf("\"%s\": \"%s\"", token, c.model.SpellCheck(token)) // Json "key: value"

		if i != len(tokens)-1 {
			result += ",\n"
		}
	}
	result += "}" // Close json

	return result
}

func (c *Checker) get(w http.ResponseWriter, r *http.Request) {
	c.muxResults.Lock()

	path := r.URL.Path[1:]
	id, _ := strconv.Atoi(path)

	if result, ok := c.results[id]; ok {
		w.Header().Add("Content-Type", "text/plain; charset=latin-1")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, result)

		log.Printf("Result \"%d\" got from database", id)
		delete(c.results, id)
	} else {
		log.Printf("\"%s\" tried to get unknown result \"%s\"", r.RemoteAddr, path)
		http.Error(w, "Result not found on database, maybe is processing yet.", http.StatusNotFound)
	}

	c.muxResults.Unlock()
}

func (c *Checker) post(w http.ResponseWriter, r *http.Request) {
	task := Tasks{}

	body, _ := ioutil.ReadAll(r.Body)
	task.toCorrect = string(body)

	if task.toCorrect == "" {
		log.Printf("Invalid task received from \"%s\"", r.RemoteAddr)
		http.Error(w, "Not a valid string received", http.StatusBadRequest)

		return
	}

	c.muxTasks.Lock()
	c.last++
	task.id = c.last
	c.muxTasks.Unlock()

	c.tasks <- task

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Your request \"%d\" will be processed. Soon get your result at \"http://%s/%d\"", task.id, r.Host, task.id)

	log.Printf("Received request %d from %s", task.id, r.RemoteAddr)
}

func getChecker(model *fuzzy.Model) *Checker {
	checker := Checker{}
	checker.model = model

	checker.tasks = make(chan Tasks)
	checker.results = make(map[int]string)

	return &checker
}

func getMaxJobs() int {
	maxJobs := os.Getenv(maxJobsEnv)
	if maxJobs == "" {
		os.Setenv(maxJobsEnv, strconv.Itoa(maxJobsDefault))
		maxJobs = os.Getenv(maxJobsEnv)
	}

	maxJobsInt, err := strconv.Atoi(maxJobs)
	if err != nil {
		log.Fatalf("Invalid port selected: \"%s\"", maxJobs)
	}

	return maxJobsInt
}

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "train" {
		if len(os.Args) < 3 {
			log.Fatal("Wrong Usage. Example: ./spellchecker train [depth of search] [vocabulary]")
		}

		trainModel()
		os.Exit(0)
	}

	log.Println("Loading Model...")
	model, err := fuzzy.Load(ModelFile)
	if err != nil {
		log.Fatal("Model not trained. Train it first with: ./spellchecker train [depth of search] [vocabulary]")
	}
	log.Println("Model loaded!")

	checker := getChecker(model)

	maxJobs := getMaxJobs()
	limitConcurrentJobsQueue := make(chan int, maxJobs)

	go func() {
		for {
			task := <-checker.tasks
			limitConcurrentJobsQueue <- -1 // Whatever the number

			go func() {
				log.Printf("Correcting result \"%d\"...", task.id)

				tokens := strings.Split(task.toCorrect, "\n")
				log.Printf("Total of words from task \"%d\" = %d\n", task.id, len(tokens))
				result := checker.check(tokens)

				checker.muxResults.Lock()
				checker.results[task.id] = result
				checker.muxResults.Unlock()

				log.Printf("Saved result \"%d\"!", task.id)
				<-limitConcurrentJobsQueue
			}()
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			checker.get(w, r)
		case "POST":
			checker.post(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	port := os.Getenv(portEnv)
	if port == "" {
		os.Setenv(portEnv, portDefault)
		port = os.Getenv(portEnv)
	}
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
