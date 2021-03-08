package t3bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/jaytaylor/html2text"
	"github.com/matrix-org/go-neb/types"
	"github.com/matrix-org/gomatrix"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"maunium.net/go/mautrix/id"
)

type cmcTicker struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	PriceUSD string `json:"price_usd"`
	Rank     string `json:"rank"`
	Pct1H    string `json:"percent_change_1h"`
	Pct24H   string `json:"percent_change_24h"`
	Pct7D    string `json:"percent_change_7d"`
	CapUSD   string `json:"market_cap_usd"`
}

var allTickers []cmcTicker

func (s *Service) cmdCMC(client types.MatrixClient, roomID id.RoomID, userID id.UserID, args []string) (*gomatrix.HTMLMessage, error) {
	var tickers []cmcTicker

	// Make sure we have data about all coins known by CMC, which findCoinID()
	// uses to look up the canonical id for a coin given the user-entered name.
	if len(allTickers) == 0 {
		allTickersNew, err := getAllTickers()
		if err != nil {
			return nil, err
		}
		allTickers = *allTickersNew
	}

	if len(args) == 0 {
		args = []string{"tezos"}
	}

	for _, arg := range args {
		coinID, err := findCoinID(arg, &allTickers)
		if err != nil {
			return nil, err // TODO handle gracefully
		}

		response, err := queryCMC(coinID + "/")
		if err != nil {
			return nil, err
		}
		//log.WithFields(log.Fields{"response": string(*response)}).Info("CMC response")

		var ts []cmcTicker
		err2 := json.Unmarshal(*response, &ts)
		if err2 != nil {
			return nil, err // TODO hide from user
		}
		tickers = append(tickers, ts...)
	}

	return displayTickers(&tickers)
}

// queryCMC sends a ticker query to the coinmarketcap.com API and returns the
// response bytes, which should be a JSON value for an array of ticker objects.
func queryCMC(query string) (*[]byte, error) {
	log.WithFields(log.Fields{"query": query}).Info("queryCMC")

	url := "https://api.coinmarketcap.com/v1/ticker/" + query
	resp, err := http.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get of %s returned code %v", url, resp.StatusCode)
	}
	bodyBytes, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return nil, err2
	}
	return &bodyBytes, nil
}

func getAllTickers() (*[]cmcTicker, error) {
	tickerBytes, err := queryCMC("?limit=0")
	if err != nil {
		return nil, err
	}

	var tickers []cmcTicker
	err2 := json.Unmarshal(*tickerBytes, &tickers)
	if err2 != nil {
		return nil, err2
	}
	log.WithFields(log.Fields{"len": len(tickers)}).Info("getAllTickers")
	return &tickers, nil
}

// findCoin takes a user-supplied coin name and tries to find the canonical CMC
// id for it (as needed for queries to their API), referencing an array of
// ticker info that gives the Id, Name, and Symbol for a set coin types.
func findCoinID(arg string, tickers *[]cmcTicker) (string, error) {
	target := strings.ToLower(arg)
	for _, t := range *tickers {
		if target == strings.ToLower(t.Symbol) || target == strings.ToLower(t.Id) || target == strings.ToLower(t.Name) {
			log.WithFields(log.Fields{"arg": arg, "return": t.Id}).Info("findCoinID")
			return t.Id, nil
		}
	}
	return "", fmt.Errorf("coin name '%s' not found", arg)
}

func (s *Service) cmdTop(client *gomatrix.Client, roomID, userID string, maxlimit int, args []string) (*gomatrix.HTMLMessage, error) {
	limit := 5
	if len(args) > 0 {
		lim, err := strconv.Atoi(args[0])
		if err == nil {
			limit = lim
		}
	}

	if limit > maxlimit {
		return simpleMessage(fmt.Sprintf("Yeah, that would spam the room. Try %d or fewer.", maxlimit)), nil
	}

	query := fmt.Sprintf("?limit=%d", limit)
	response, err := queryCMC(query)
	if err != nil {
		return nil, err
	}

	var ts []cmcTicker
	err2 := json.Unmarshal(*response, &ts)
	if err2 != nil {
		return nil, err
	}

	return displayTickers(&ts)
}

func displayTickers(tickers *[]cmcTicker) (*gomatrix.HTMLMessage, error) {
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
<td>%s</td>
<td>%s</td>
<td>%s</td>
<td>%s</td>
<td>%s</td>
<td>%s</td>
</tr>`

	tbody := `<tbody>`
	millionD, _ := decimal.NewFromString("1000000")
	var capS string
	for _, ticker := range *tickers {
		capD, err := decimal.NewFromString(ticker.CapUSD)
		if err != nil {
			log.WithFields(log.Fields{"s": ticker.CapUSD, "err": err}).Warn("failed conversion to decimal")
			capS = "?"
		} else {
			// Calculate and display market-cap rounded to millions, using decimal type to maintain precision.
			capS = capD.Div(millionD).Round(0).String()
		}

		tbody += fmt.Sprintf(rowFormat, ticker.Symbol, ticker.PriceUSD,
			ticker.Pct1H, ticker.Pct24H, ticker.Pct7D, ticker.Rank,
			capS)
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
