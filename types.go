package main

// Price 结构体
type Price struct {
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Volume int64   `json:"volume"`
	Time   string  `json:"time"`
}

// PriceResponse 结构体
type PriceResponse struct {
	Ticker string  `json:"ticker"`
	Prices []Price `json:"prices"`
}
