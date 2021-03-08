package t3bot

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type cmcProStatus struct {
	Timestamp     string
	Error_code    int
	Error_message string
	Elapsed       int
	Credit_count  int
}

type cmcProMapItem struct {
	Id     int
	Name   string
	Symbol string
	Slug   string
}

type cmcProMapResponse struct {
	Status cmcProStatus
	Data   []cmcProMapItem
}

type cmcProListing struct {
	Id       int
	Name     string
	Symbol   string
	Cmc_rank int
	Quote    struct {
		USD struct {
			Price              float64
			Volume_24h         float64
			Percent_change_1h  float64
			Percent_change_24h float64
			Percent_change_7d  float64
			Market_cap         float64
			Last_updated       string
		}
	}
}

type cmcProListingResponse struct {
	Status cmcProStatus
	Data   []cmcProListing
}

func getCmcProListings() (*[]cmcProListing, error) {
	bytes, err := queryCMCPro("cryptocurrency/listings/latest")
	if err != nil {
		return nil, err
	}

	var response cmcProListingResponse
	err2 := json.Unmarshal(*bytes, &response)
	if err2 != nil {
		return nil, err2
	}
	return &response.Data, nil
}

func findCoinIDPro(arg string, tickers *[]cmcProMapItem) (int, error) {
	target := strings.ToLower(arg)
	for _, t := range *tickers {
		if target == strings.ToLower(t.Symbol) || target == strings.ToLower(t.Name) {
			return t.Id, nil
		}
	}
	return 0, fmt.Errorf("coin name '%s' not found", arg)
}

func getCmcProMap() (*[]cmcProMapItem, error) {
	bytes, err := queryCMCPro("cryptocurrency/map")
	if err != nil {
		return nil, err
	}

	var response cmcProMapResponse
	err2 := json.Unmarshal(*bytes, &response)
	if err2 != nil {
		return nil, err2
	}
	return &response.Data, nil
}

func queryCMCPro(query string) (*[]byte, error) {
	log.WithFields(log.Fields{"query": query}).Info("queryCMCPro")

	client := &http.Client{}

	url := "https://pro-api.coinmarketcap.com/v1/" + query
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	api_key := os.Getenv("CMC_PRO_API_KEY")
	if len(api_key) == 0 {
		return nil, fmt.Errorf("API key not set in environ")
	}

	req.Header.Add("X-CMC_PRO_API_KEY", api_key)

	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("Get(%s): %v", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get of %s returned code %v", url, resp.StatusCode)
	}
	bodyBytes, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return nil, fmt.Errorf("ReadAll: %v", err2)
	}
	return &bodyBytes, nil
}
