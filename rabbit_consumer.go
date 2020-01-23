/*
Spellchecker binary inteded to be used as a standalone binary.
Not using files for making easier the use thorugh another source code written in another language.
*/
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/sajari/fuzzy"
	"github.com/streadway/amqp"
)

// ModelFile is the name of the file which is the spell model
const ModelFile string = "spell_model.json"

const (
	hostDefault    = "localhost"
	hostEnv        = "HOST"
	portDefault    = "5672"
	portEnv        = "PORT"
	maxJobsEnv     = "MAX_JOBS"
	maxJobsDefault = 4
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

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

// Checker will hold information of the main flow the application
type Checker struct {
	model *fuzzy.Model
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

func getUrl() string {
	host := os.Getenv(hostEnv)
	if host == "" {
		os.Setenv(hostEnv, hostDefault)
		host = os.Getenv(hostEnv)
	}

	port := os.Getenv(portEnv)
	if port == "" {
		os.Setenv(portEnv, portDefault)
		port = os.Getenv(portEnv)
	}

	return fmt.Sprintf("amqp://guest:guest@%s:%s/", host, port)
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

type Task struct {
	Id   string
	Text string
}

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "train" {
		if len(os.Args) < 3 {
			log.Fatal("Wrong Usage. Example: ./spellchecker train [depth of search] [vocabulary]")
		}

		trainModel()
		os.Exit(0)
	}

	conn, err := amqp.Dial(getUrl())
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	tasksQueue, err := ch.QueueDeclare(
		"tasks", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue 'tasks'")

	tasksResults, err := ch.QueueDeclare(
		"results", // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		getMaxJobs(), // prefetch count
		0,            // prefetch size
		false,        // global
	)
	failOnError(err, "Failed to set QoS")

	tasks, err := ch.Consume(
		tasksQueue.Name, // queue
		"",              // consumer
		false,           // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	failOnError(err, "Failed to register a consumer")

	log.Println("Loading Model...")
	model, err := fuzzy.Load(ModelFile)
	if err != nil {
		log.Fatal("Model not trained. Train it first with: ./spellchecker train [depth of search] [vocabulary]")
	}
	log.Println("Model loaded!")

	checker := Checker{model}
	forever := make(chan bool)

	go func() {
		for t := range tasks {
			var task Task
			if err := json.Unmarshal(t.Body, &task); err != nil {
				log.Fatal(err)
			}

			log.Printf("Received text \"%s\"", task.Id)
			tokens := strings.Split(task.Text, " ")

			go func(t amqp.Delivery) {
				log.Printf("Processing... \"%s\"", task.Id)

				result := checker.check(tokens)

				resultMap := map[string]string{task.Id: result}
				resultMsg, err := json.Marshal(resultMap)
				if err != nil {
					log.Fatal(err)
				}

				err = ch.Publish(
					"",                // exchange
					tasksResults.Name, // routing key
					false,             // mandatory
					false,             // immediate
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(resultMsg),
					})
				failOnError(err, "Failed to publish a message")
				log.Printf("Submited result for text \"%s\"", task.Id)

				t.Ack(false)
			}(t)
		}
	}()

	log.Printf("RUNNING. Waiting for tasks...")
	<-forever
}
