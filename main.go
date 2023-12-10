package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Order struct {
	Prices     []float32
	Quantities []int
	Country    string
	Reduction  string
}

type Total struct {
	Total float32 `json:"total"`
}

var CountryReduction = map[string]float32{
	"DE": 0.2,
	"UK": 0.21,
	"FR": 0.2,
	"IT": 0.25,
	"ES": 0.19,
	"PL": 0.21,
	"RO": 0.2,
	"NL": 0.2,
	"BE": 0.24,
	"EL": 0.2,
	"CZ": 0.19,
	"PT": 0.23,
	"HU": 0.27,
	"SE": 0.23,
	"AT": 0.22,
	"BG": 0.21,
	"DK": 0.21,
	"FI": 0.17,
	"SK": 0.18,
	"IE": 0.21,
	"HR": 0.23,
	"LT": 0.23,
	"SI": 0.24,
	"LV": 0.20,
	"EE": 0.22,
	"CY": 0.21,
	"LU": 0.25,
	"MT": 0.20,
}

type Reply struct {
	Total float32 `json:"total"`
}

func main() {
	http.HandleFunc("/order", handler)
	http.HandleFunc("/feedback", func(rw http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			fmt.Printf("error reading body: %v\n", err)
			rw.WriteHeader(204)
			return
		}
		defer req.Body.Close()

		fmt.Printf("Feedback: %s\n", body)

		rw.WriteHeader(200)
	})

	err := http.ListenAndServe(fmt.Sprintf(":%s", getPort()), nil)
	if err != nil {
		log.Fatal("Listen and serve:", err)
	}
}

func handler(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		rw.WriteHeader(400)
		return
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Printf("error reading body: %v\n", err)
		rw.WriteHeader(400)
		return
	}
	defer req.Body.Close()

	order := Order{}

	err = json.Unmarshal(body, &order)
	log.Println(order)
	if err != nil {
		fmt.Printf("error parsing order: %v\n", err)
		rw.WriteHeader(400)
		return
	}

	_, ok := CountryReduction[order.Country]
	if !ok {
		fmt.Printf("error parsing order: %v\n", err)
		rw.WriteHeader(400)
		return
	}
	total := calculateTotal(order)

	if total == 0 {
		rw.WriteHeader(400)
		return
	}
	if total == -1 {
		rw.WriteHeader(404)
		return
	}

	reply := Total{Total: total}

	rw.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(rw).Encode(reply)
	if err != nil {
		fmt.Printf("error encoding reply: %v\n", err)
		rw.WriteHeader(500)
		return
	}
}

func calculateTotal(order Order) float32 {
	if len(order.Prices) != len(order.Quantities) {
		return 0
	}
	var total float32
	for i := 0; i < len(order.Prices); i++ {
		if order.Prices[i] < 0 || order.Quantities[i] < 0 {
			return 0
		}
		total += order.Prices[i] * float32(order.Quantities[i])
	}
	log.Printf("Total before Vat: %f\n", total)
	total = total + total*CountryReduction[order.Country]

	log.Printf("Total before reduction: %f\n", total)
	if order.Reduction == "STANDARD" {
		switch {
		case total >= 50000:
			total = total * 0.85
		case total >= 10000:
			total = total * 0.9
		case total >= 7000:
			total = total * 0.93
		case total >= 5000:
			total = total * 0.95
		case total >= 1000:
			total = total * 0.97
		}
	}
	if order.Reduction == "HALF PRICE" {
		total = total * 0.5
	}
	if order.Reduction == "" {
		return 0
	}
	if order.Reduction == "BLACK FRIDAY" {
		return -1
	}
	log.Printf("Total after reduction: %f\n", total)
	return total
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return "9000"
	}
	return port
}
