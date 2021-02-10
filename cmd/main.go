package main

import (
	"errors"
	"fmt"
	"github.com/masihyeganeh/binance-ingest/internal/binance"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
)

func main() {
	// Handling graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	allowedSymbols, err := getSymbolsFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	streams := getStreamsFromSymbols(allowedSymbols, "trade")

	u := url.URL{
		Scheme:   "wss",
		Host:     "stream.binance.com:9443",
		Path:     "/stream",
		RawQuery: "streams=" + strings.Join(streams, "/"),
	}
	log.Printf("connecting to %s", u.String())

	app, err := binance.New(u, allowedSymbols)
	if err != nil {
		log.Fatal(err)
	}

	defer app.Close()

	done := make(chan struct{})

	go app.ReadFromSocket(done)
	go app.TradeGenerator(done)

	err = app.Start(done, interrupt)
	if err != nil {
		log.Println(err.Error())
	}

}

func getStreamsFromSymbols(symbols map[string]bool, eventType string) []string {
	result := make([]string, len(symbols))

	i := 0
	for symbol := range symbols {
		result[i] = fmt.Sprintf("%s@%s", symbol, eventType)
		i++
	}

	return result
}

func getSymbolsFromEnv() (map[string]bool, error) {
	allowedSymbols := make(map[string]bool)
	symbols := strings.Split(os.Getenv("symbols"), ",")

	for _, symbol := range symbols {
		symbol = strings.TrimSpace(strings.ToLower(symbol))
		if len(symbol) == 0 {
			continue
		}
		if _, exists := allowedSymbols[symbol]; !exists {
			allowedSymbols[symbol] = true
		}
	}

	if len(allowedSymbols) == 0 {
		return nil, errors.New("no symbol is requested")
	}

	if len(allowedSymbols) > 1024 {
		return nil, errors.New("no more than 1024 streams are allowed. you asked for " + strconv.Itoa(len(allowedSymbols)))
	}

	return allowedSymbols, nil
}
