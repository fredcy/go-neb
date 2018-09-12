// CoinMarketCap.com API v2 functions
package t3bot

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/jaytaylor/html2text"
	"github.com/matrix-org/gomatrix"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type cmcListing2 struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Slug   string `json:"website_slug"`
}

type cmcTickerResponse2 struct {
	Data     []cmcTicker2 `json:"data"`
	Metadata struct {
		Timestamp int    `json:"timestamp"`
		Error     string `json:"error"`
	} `json:"metadata"`
}

func (s *Service) cmdCMC2(client *gomatrix.Client, roomID, userID string, args []string) (*gomatrix.HTMLMessage, error) {
	var tickers []cmcTicker2

	// Make sure we have data about all coins known by CMC, which findCoinID()
	// uses to look up the canonical id for a coin given the user-entered name.
	if len(allTickers2) == 0 {
		allTickersNew, err := getAllTickers2()
		if err != nil {
			return nil, fmt.Errorf("getAllTickers2: %v", err)
		}
		allTickers2 = *allTickersNew
	}

	if len(args) == 0 {
		args = []string{"tezos"}
	}

	for _, arg := range args {
		coinID, err := findCoinID2(arg, &allTickers2)
		if err != nil {
			return nil, fmt.Errorf("findCoinID2: %v", err) // TODO handle gracefully
		}

		responseBytes, err := queryCMC2("/ticker/" + strconv.Itoa(coinID) + "/")
		if err != nil {
			return nil, fmt.Errorf("queryCMC2: %v", err)
		}
		//log.WithFields(log.Fields{"response": string(*response)}).Info("CMC response")

		var response cmcTickerResponse2
		err2 := json.Unmarshal(*responseBytes, &response)
		if err2 != nil {
			return nil, fmt.Errorf("Unmarshal: %v", err) // TODO hide from user
		}
		if len(response.Metadata.Error) > 0 {
			return nil, fmt.Errorf("response error: %s", response.Metadata.Error)
		}
		tickers = append(tickers, response.Data...)
	}

	return displayTickers2(&tickers)
}

func queryCMC2(query string) (*[]byte, error) {
	log.WithFields(log.Fields{"query": query}).Info("queryCMC2")

	url := "https://api.coinmarketcap.com/v2/" + query
	resp, err := http.Get(url)
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

type cmcListingResponse2 struct {
	Data     []cmcListing2 `json:"data"`
	Metadata struct {
		Timestamp            int    `json:"timestamp"`
		Num_Cryptocurrencies int    `json:"num_cryptocurrencies"`
		Error                string `json:"error"`
	} `json:"metadata"`
}

func getAllTickers2() (*[]cmcListing2, error) {
	bytes, err := queryCMC2("listings/")
	if err != nil {
		return nil, err
	}

	var response cmcListingResponse2
	err2 := json.Unmarshal(*bytes, &response)
	if err2 != nil {
		return nil, err2
	}
	log.WithFields(log.Fields{"len": len(response.Data)}).Info("getAllTickers2")
	return &response.Data, nil
}

// findCoin takes a user-supplied coin name and tries to find the canonical CMC
// id for it (as needed for queries to their API), referencing an array of
// ticker info that gives the Id, Name, and Symbol for a set coin types.
func findCoinID2(arg string, tickers *[]cmcListing2) (int, error) {
	target := strings.ToLower(arg)
	for _, t := range *tickers {
		if target == strings.ToLower(t.Symbol) || target == strings.ToLower(t.Name) {
			log.WithFields(log.Fields{"arg": arg, "return": t.Id}).Info("findCoinID2")
			return t.Id, nil
		}
	}
	return 0, fmt.Errorf("coin name '%s' not found", arg)
}

type cmcTicker2 struct {
	Id      int     `json:"id"`
	Name    string  `json:"name"`
	Symbol  string  `json:"symbol"`
	Slug    string  `json:"website_slug"`
	Rank    int     `json:"rank"`
	CircSup float64 `json:"circulating_supply"`
	TotSup  float64 `json:"total_supply"`
	MaxSup  float64 `json:"max_supply"`
	Quotes  struct {
		USD struct {
			Price  float64 `json:"price"`
			Vol24  float64 `json:"volume_24h"`
			Cap    float64 `json:"market_cap"`
			Pct1H  float64 `json:"percent_change_1h"`
			Pct24H float64 `json:"percent_change_24h"`
			Pct7D  float64 `json:"percent_change_7d"`
		} `json:"USD"`
	} `json:"quotes"`
}

var allTickers2 []cmcListing2

func displayTickers2(tickers *[]cmcTicker2) (*gomatrix.HTMLMessage, error) {
	thead := `<thead><tr>
<th>symbol</th>
<th>Latest (USD)</th>
<th>1H %Δ</th>
<th>24H %Δ</th>
<th>7D %Δ</th>
<th>Rank</th>
<th>Mkt Cap (M USD)</th>
</tr></thead>`

	rowFormat := `<tr>
<td>%s</td>
<td>%f</td>
<td>%f</td>
<td>%f</td>
<td>%f</td>
<td>%f</td>
<td>%f</td>
</tr>`

	tbody := `<tbody>`
	for _, ticker := range *tickers {
		tbody += fmt.Sprintf(rowFormat, ticker.Symbol, ticker.Quotes.USD.Price,
			ticker.Quotes.USD.Pct1H, ticker.Quotes.USD.Pct24H, ticker.Quotes.USD.Pct7D,
			ticker.Rank, ticker.Quotes.USD.Cap/1000000)
	}
	tbody += `</tbody>`
	table := `<table>` + thead + tbody + `</table>`

	// Convert the HTML table to text alternative format. Unfortunately, Riot
	// clients on android and IOS ignore this and display a half-fast rendering
	// of the HTML without any tabular layout.
	tableText, err3 := html2text.FromString(table, html2text.Options{PrettyTables: true})
	if err3 != nil {
		return nil, err3
	}

	htmlMessage := gomatrix.HTMLMessage{
		MsgType:       "m.notice",
		Format:        "org.matrix.custom.html",
		FormattedBody: table,
		Body:          tableText,
	}

	return &htmlMessage, nil
}
