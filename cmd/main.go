package main

import (
	"encoding/json"
	"errors"
	"github.com/masihyeganeh/binance-ingest/internal/binance"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
)

func main() {
	// Handling graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Reading symbols from ENV
	symbols, err := getSymbolsFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	app, err := binance.New(symbols)
	if err != nil {
		log.Fatal(err)
	}

	defer app.Close()

	done := make(chan struct{})

	go func() {
		err = app.Start(done, interrupt)
		if err != nil {
			log.Println(err.Error())
		}
	}()

	log.Println("Started receiving trades")

	err = app.WatchSymbol("NEOBTC")
	if err != nil {
		log.Print(err.Error())
	} else {
		log.Println("Watched NEOBTC")
	}

	err = app.UnwatchSymbol("BNBBTC")
	if err != nil {
		log.Print(err.Error())
	} else {
		log.Println("Unwatched BNBBTC")
	}

	recv := app.Receive()
	for {
		select {
		case <-done:
			return
		case res := <-recv:
			message, _ := json.MarshalIndent(res.Data, "", "	")
			log.Printf("received message from %s to process:\n%s\n\n", res.Stream, message)
		}
	}
}

func getSymbolsFromEnv() (map[string]bool, error) {
	symbols := make(map[string]bool)
	envSymbols := strings.Split(os.Getenv("symbols"), ",")

	for _, symbol := range envSymbols {
		symbol = strings.TrimSpace(strings.ToLower(symbol))
		if len(symbol) == 0 {
			continue
		}
		if _, exists := symbols[symbol]; !exists {
			symbols[symbol] = true
		}
	}

	if len(symbols) == 0 {
		return nil, errors.New("no symbol is requested")
	}

	if len(symbols) > 1024 {
		return nil, errors.New("no more than 1024 streams are allowed. you asked for " + strconv.Itoa(len(symbols)))
	}

	return symbols, nil
}
