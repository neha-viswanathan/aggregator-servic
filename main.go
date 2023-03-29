package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"log"
	"net/http"
	"sync"
	"time"
)

type Aggregator struct {
	discoveryUrl  string
	uniqueFlavors map[string]bool // hashset of flavors
	mu            sync.Mutex      // mutex to synchronize access to the hashset
}

func NewAggregator(url string) *Aggregator {
	return &Aggregator{discoveryUrl: url}
}

func (a *Aggregator) GetUniqueFlavors() []string {
	a.mu.Lock()         // acquire the lock before reading from the flavor hashset
	defer a.mu.Unlock() // release the lock when exiting

	// build the results from the current hashset
	var flavors []string
	for flavor := range a.uniqueFlavors {
		flavors = append(flavors, flavor)
	}
	return flavors
}

func (a *Aggregator) SetUniqueFlavors(flavors map[string]bool) {
	a.mu.Lock()         // acquire the lock before updating the flavor hashset
	defer a.mu.Unlock() // release the lock when exiting

	// update the current hashset
	a.uniqueFlavors = flavors
}

// Start starts the background updates of the flavors at regular intervals
func (a *Aggregator) Start(refreshRate time.Duration) {
	go func() {
		go a.RetrieveFlavors()
		for range time.Tick(refreshRate) {
			go a.RetrieveFlavors()
		}
	}()
}

func (a *Aggregator) RetrieveFlavors() {
	res, err := http.Get(a.discoveryUrl)
	if err != nil {
		log.Println("failed to discover shops")
		return
	}
	defer res.Body.Close()

	reader := csv.NewReader(res.Body)
	shops, err := reader.ReadAll()
	if err != nil {
		log.Println(err)
		return
	}

	// Channel to handle shop responses
	shopResponses := make(chan []string)

	// Get flavors for each shop
	for _, r := range shops {
		shopName := r[0]
		shopURL := r[1]
		log.Printf("processing shop: [%s], url: [%s]", shopName, shopURL)
		// Process the shop URL
		go a.getIceCreamFlavors(shopURL, shopResponses)
	}

	// Read responses from the channel when the go routine has completed
	uniqueFlavors := map[string]bool{}
	for range shops {
		flavors := <-shopResponses
		for _, entry := range flavors {
			uniqueFlavors[entry] = true
		}
	}

	// Update the unique flavors
	a.SetUniqueFlavors(uniqueFlavors)
}

// getIceCreamFlavors gets the ice cream flavors for a given shop
func (a *Aggregator) getIceCreamFlavors(url string, out chan []string) {
	res, err := http.Get(url)
	if err != nil {
		log.Println("failed to get flavors")
		out <- []string{}
		return
	}
	defer res.Body.Close()

	scanner := bufio.NewScanner(res.Body)
	scanner.Split(bufio.ScanLines)

	var flavors []string
	for scanner.Scan() {
		flavors = append(flavors, scanner.Text())
	}
	out <- flavors
}

var url = flag.String("discovery-url", "", "the url to load the list of icecream shop urls")
var port = flag.String("port", ":8081", "the port to bind the http server to")

var aggregator *Aggregator

func main() {
	flag.Parse()
	if *url == "" {
		log.Fatalln("Bad args. discovery-url not defined")
	}

	start(*port, *url)
	// Wait forever
	select {}
}

func start(addr, url string) *http.Server {
	// Instantiate a new aggregator
	aggregator = NewAggregator(url)

	// Start aggregating flavors from the discovery service
	aggregator.Start(10 * time.Second)

	// Dump all the flavors
	http.HandleFunc("/flavors", func(w http.ResponseWriter, r *http.Request) {
		flavors := aggregator.GetUniqueFlavors()
		writer := csv.NewWriter(w)
		for _, flavor := range flavors {
			writer.Write([]string{flavor})
		}
		writer.Flush()
	})

	srv := &http.Server{Addr: addr}
	go func() {
		log.Println("http server exited: ", srv.ListenAndServe())
	}()
	return srv
}
