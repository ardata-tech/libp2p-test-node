package protocols

import (
	"encoding/json"
	"net/http"
)

type PriceService struct{}

func NewPriceService() *PriceService {
	return &PriceService{}

}

type EthereumPriceResponse struct {
	Ethereum struct {
		Usd float64 `json:"usd"`
	} `json:"ethereum"`
}

func (p *PriceService) GetPrice() (EthereumPriceResponse, error) {

	// http call
	// return price
	req, err := http.Get("https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd")
	if err != nil {
		return EthereumPriceResponse{}, err
	}
	defer req.Body.Close()

	var price EthereumPriceResponse
	err = json.NewDecoder(req.Body).Decode(&price)
	if err != nil {
		return EthereumPriceResponse{}, err

	}
	return price, nil
}
