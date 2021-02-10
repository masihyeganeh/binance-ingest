package structs

type TradePayload struct {
	EventType                string `json:"e"`
	EventTime                int    `json:"E"`
	Symbol                   string `json:"s"`
	TradeId                  int    `json:"t"`
	Price                    string `json:"p"`
	Quantity                 string `json:"q"`
	BuyerOrderId             int    `json:"b"`
	SellerOrderId            int    `json:"a"`
	TradeTime                int    `json:"T"`
	IsTheBuyerTheMarketMaker bool   `json:"m"`
	Ignore                   bool   `json:"M"`
}

type ResponseStreamPayload struct {
	Stream string       `json:"stream"`
	Data   TradePayload `json:"data"`
}

type WrappedStreamPayload struct {
	Stream string      `json:"stream"`
	Data   interface{} `json:"data"`
}
