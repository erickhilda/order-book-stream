package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
)

type Client struct {
	name   string
	events chan *DashBoard
}

type Order struct {
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

type DashBoard struct {
	Symbol    string  `json:"symbol"`
	Timestamp int64   `json:"timestamp"`
	Bids      []Order `json:"bids"`
	Asks      []Order `json:"asks"`
}

func generateOrder() []Order {
	var orders = []Order{}
	var orderLength = rand.Intn(6) + 1
	for i := 0; i < orderLength; i++ {
		orders = append(orders, Order{
			Price:    rand.Float64() * 100000000,
			Quantity: rand.Intn(10000) + 1,
		})
	}
	return orders
}

var globalBids = generateOrder()
var globalAsks = generateOrder()

func main() {
	app := fiber.New()
	app.Get("/events", adaptor.HTTPHandler(handler(dashboardHandler)))
	app.Listen(":9001")
}

func handler(f http.HandlerFunc) http.Handler {
	return http.HandlerFunc(f)
}
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	client := &Client{name: r.RemoteAddr, events: make(chan *DashBoard, 10)}
	go updateDashboard(client)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	timeout := time.After(3 * time.Second)
	select {
	case ev := <-client.events:
		enc.Encode(ev)

		fmt.Fprintf(w, "data: %v\n\n", buf.String())
		// fmt.Printf("data: %v\n", buf.String())
	case <-timeout:
		fmt.Fprintf(w, ": nothing to sent\n\n")
	}

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func generateInitialBidsAndAsks(order string) {
	// Q: how can i prevent the initial bids and asks to genetate on every request?
	// A: use redis to store the initial bids and asks
	// Q: okay i dont have redis, how can i do it?
	// A: use global variable to store the initial bids and asks
	// Q: how to make it global?
	// A: declare it outside the function

	if order == "bid" {
		bidsToChangeIndex := rand.Intn(len(globalBids))
		globalBids[bidsToChangeIndex].Price = rand.Float64() * 1000000
		globalBids[bidsToChangeIndex].Quantity = rand.Intn(10000) + 1
	} else {
		asksToChangeIndex := rand.Intn(len(globalAsks))
		globalAsks[asksToChangeIndex].Price = rand.Float64() * 1000000
		globalAsks[asksToChangeIndex].Quantity = rand.Intn(10000) + 1
	}
}

func updateDashboard(client *Client) {

	var orderType []string = []string{"bid", "ask"}
	var order string = orderType[rand.Intn(len(orderType))]
	generateInitialBidsAndAsks(order)

	for {
		db := &DashBoard{
			Symbol:    "BTC_IDR",
			Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
			Bids:      globalBids,
			Asks:      globalAsks,
		}

		client.events <- db
	}
}
