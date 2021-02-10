package binance

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	. "github.com/masihyeganeh/binance-ingest/internal/structs"
	"github.com/pkg/errors"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

type Binance struct {
	conn           *websocket.Conn
	allowedSymbols map[string]bool
	tradesQueue    chan TradePayload
}

func New(u url.URL, allowedSymbols map[string]bool) (*Binance, error) {
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "error while dialing")
	}

	return &Binance{
		conn:           c,
		allowedSymbols: allowedSymbols,
		tradesQueue:    make(chan TradePayload, TradesQueueSize),
	}, nil
}

func (b *Binance) Start(done chan struct{}, interrupt chan os.Signal) error {
	for {
		select {
		case <-done:
			return nil
		case trade := <-b.tradesQueue:
			// Only allow designated symbols and "trade" event type
			if _, ok := b.allowedSymbols[strings.ToLower(trade.Symbol)]; !ok || trade.EventType != "trade" {
				continue
			}

			err := b.conn.WriteJSON(WrappedStreamPayload{
				Stream: fmt.Sprintf("%s@%s", trade.Symbol, trade.EventType),
				Data:   trade,
			})
			if err != nil {
				return errors.Wrap(err, "error while writing")
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := b.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				return errors.Wrap(err, "error while closing")
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return nil
		}
	}
}

// TradeGenerator is simulating starting part of pipeline that trades are ingested to
func (b *Binance) TradeGenerator(done chan struct{}) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			b.tradesQueue <- TradePayload{
				EventType:                "trade",
				EventTime:                123456789,
				Symbol:                   "BNBBTC",
				TradeId:                  12345,
				Price:                    "0.001",
				Quantity:                 "100",
				BuyerOrderId:             88,
				SellerOrderId:            50,
				TradeTime:                123456785,
				IsTheBuyerTheMarketMaker: true,
				Ignore:                   true,
			}
		}
	}
}

func (b *Binance) ReadFromSocket(done chan struct{}) {
	defer close(done)
	for {
		_, message, err := b.conn.ReadMessage()
		if err != nil {
			log.Println("error while reading:", err)
			return
		}
		var res ResponseStreamPayload
		if err = json.Unmarshal(message, &res); err != nil {
			log.Printf("received message: %s", message)
			continue
		}

		message, _ = json.MarshalIndent(res.Data, "", "	")
		log.Printf("received message from %s:\n%s\n\n", res.Stream, message)
	}
}

func (b *Binance) Close() error {
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}
