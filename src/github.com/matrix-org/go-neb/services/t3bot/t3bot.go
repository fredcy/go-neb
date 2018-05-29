// Package t3bot implements a Service for Tezos room !commands.
package t3bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/jaytaylor/html2text"
	"github.com/matrix-org/go-neb/types"
	"github.com/matrix-org/gomatrix"
	"github.com/shopspring/decimal"
)

// ServiceType of the T3bot service
const ServiceType = "t3bot"

// Service represents the T3bot service. It has no Config fields.
type Service struct {
	types.DefaultService
}

var roomsMessageHTML = `
General &nbsp;&nbsp;&nbsp; <a href="https://riot.im/app/#/room/#tezos:matrix.org">#tezos:matrix.org</a><br>
Tech &nbsp;&nbsp;&nbsp; <a href="https://riot.im/app/#/room/#freenode_#tezos:matrix.org">#freenode_#tezos:matrix.org</a><br>
Price &nbsp;&nbsp;&nbsp; <a href="https://riot.im/app/#/room/#tezostrader:matrix.org">#tezostrader:matrix.org</a><br>
Media &nbsp;&nbsp;&nbsp; <a href="https://riot.im/app/#/room/#tezosmedia:matrix.org">#tezosmedia:matrix.org</a><br>
Random &nbsp;&nbsp;&nbsp; <a href="https://riot.im/app/#/room/#tezosrandom:matrix.org">#tezosrandom:matrix.org</a><br>
Governance &nbsp;&nbsp;&nbsp; <a href="https://riot.im/app/#/room/#tezosgovernance:matrix.org">#tezosgoverance:matrix.org</a><br>
Announcements &nbsp;&nbsp;&nbsp; <a href="https://riot.im/app/#/room/#tezosannouncements:matrix.org">#tezosannouncements:matrix.org</a><br>
Ideas and Collaboration &nbsp;&nbsp;&nbsp; <a href="https://riot.im/app/#/room/#tezosfoundry:matrix.org">#tezosfoundry:matrix.org</a>
`
var sitesMessageHTML = `
Tezos Developer Documentation: <a href="http://doc.tzalpha.net/">doc.tzalpha.net</a><br>
Tezos development repository: <a href="https://gitlab.com/tezos/tezos">gitlab.com/tezos/tezos</a><br>
Tezos Commons Foundation: <a href="https://tezoscommons.org/">tezoscommons.org</a><br>
Tezos Community: <a href="https://www.tezos.community/">www.tezos.community</a><br>
Tezos rocks: <a href="https://tezos.rocks/">tezos.rocks</a><br>
Tezos help: <a href="http://www.tezos.help/">www.tezos.help</a><br>
`

var tezosMessageHTML = `
Official Tezos sites:<br>
https://www.reddit.com/user/TezosReddit<br>
https://twitter.com/TezosFoundation<br>
https://www.tezos.ch/<br>
https://tezos.com/<br>
`

var roomsHTMLMessage = gomatrix.HTMLMessage{
	MsgType:       "m.notice",
	Format:        "org.matrix.custom.html",
	FormattedBody: roomsMessageHTML,
}

var sitesHTMLMessage = gomatrix.HTMLMessage{
	MsgType:       "m.notice",
	Format:        "org.matrix.custom.html",
	FormattedBody: sitesMessageHTML,
}

var tezosHTMLMessage = gomatrix.HTMLMessage{
	MsgType:       "m.notice",
	Format:        "org.matrix.custom.html",
	FormattedBody: tezosMessageHTML,
}

/*
func init() {
	text, err := html2text.FromString(roomsMessageHTML, html2text.Options{OmitLinks: true})
	if err != nil {
		panic(err)
	}
	roomsHTMLMessage.Body = text

	text, err = html2text.FromString(sitesMessageHTML, html2text.Options{OmitLinks: true})
	if err != nil {
		panic(err)
	}
	sitesHTMLMessage.Body = text

	text, err = html2text.FromString(tezosMessageHTML, html2text.Options{OmitLinks: true})
	if err != nil {
		panic(err)
	}
	tezosHTMLMessage.Body = text
}
*/

func simpleMessage(message string) *gomatrix.HTMLMessage {
	msg := gomatrix.HTMLMessage{
		MsgType:       "m.notice",
		Format:        "org.matrix.custom.html",
		FormattedBody: message,
		Body:          message,
	}
	return &msg
}

// Commands supported:
//    !t3bot some message
// Responds with a notice of "some message".
func (e *Service) Commands(cli *gomatrix.Client) []types.Command {
	return []types.Command{
		types.Command{
			Path: []string{"t3bot"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return &gomatrix.TextMessage{"m.notice", strings.Join(args, " ")}, nil
			},
		},
		types.Command{
			Path: []string{"rooms"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return roomsHTMLMessage, nil
			},
		},
		types.Command{
			Path: []string{"sites"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return sitesHTMLMessage, nil
			},
		},
		types.Command{
			Path: []string{"tezos"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return tezosHTMLMessage, nil
			},
		},
		types.Command{
			Path: []string{"hitbtc"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				query := strings.Join(args, "/")
				response, err := e.cmdHitBTC(cli, roomID, userID, query)
				if err != nil {
					return nil, err
				} else {
					var out bytes.Buffer
					json.Indent(&out, *response, "", "    ")
					return &gomatrix.TextMessage{"m.notice", out.String()}, nil
				}
			},
		},
		types.Command{
			Path: []string{"ticker"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				query := "ticker/" + strings.Join(args, "")
				response, err := e.cmdHitBTC(cli, roomID, userID, query)
				if err != nil {
					return nil, err
				} else {
					var out bytes.Buffer
					json.Indent(&out, *response, "", "    ")
					return &gomatrix.TextMessage{"m.notice", out.String()}, nil
				}
			},
		},
		types.Command{
			Path: []string{"iou"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				query := "ticker/XTZUSD"
				response, err := e.cmdHitBTC(cli, roomID, userID, query)
				if err != nil {
					return nil, err
				} else {
					var t ticker
					err2 := json.Unmarshal(*response, &t)
					if err2 != nil {
						return nil, err
					}
					return &gomatrix.TextMessage{"m.notice", t.Last + " $/ꜩ"}, nil
				}
			},
		},
		types.Command{
			Path: []string{"cmc"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return e.cmdCMC(cli, roomID, userID, args)
			},
		},

		types.Command{
			Path: []string{"top"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return e.cmdTop(cli, roomID, userID, 10, args)
			},
		},

		types.Command{
			Path: []string{"toƿ"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return e.cmdTop(cli, roomID, userID, 100, args)
			},
		},

		types.Command{
			Path: []string{"mom-am-i-rich-yet"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return &gomatrix.TextMessage{"m.notice", "Not yet, dear boy. Go back to work."}, nil
			},
		},

		callAndResponse("bear", "ʕ ·(エ)· ʔ"),
		callAndResponse("bull", `ᓷ( ఠൠఠ )ᓸ`),
		callAndResponse("cub", `ʕ •ᴥ•ʔ`),
		callAndResponse("koala", `ʕ •ᴥ•ʔ`),					
		callAndResponse("seal", `(◕ᴥ◕)`),
		callAndResponse("whale", `. ><(((.______)`),
		callAndResponse("otter", `(:ᘌꇤ⁐  三`),
		callAndResponse("shrug", `¯\_(ツ)_/¯`),
		callAndResponse("dealwithit", `(•_•)   ( •_•)>⌐■-■    (⌐■_■)`),
		callAndResponse("disapprove", `ಠ_ಠ`),
		callAndResponse("skeptical", `ಠಿ_ಠ`),
		callAndResponse("flip", `(╯°□°）╯︵ ┻━┻`),
		callAndResponse("moon", `┗(°0°)┛`),
		callAndResponse("cthulhu", `^(;,;)^`),
		callAndResponse("fhqwhgads", "Everybody. TO THE LIMIT."),
	}
}


func callAndResponse(cmd, response string) types.Command {
	return types.Command{
		Path: []string{cmd},
		Command: func(roomID, userID string, args []string) (interface{}, error) {
			return &gomatrix.TextMessage{"m.notice", response}, nil
		},
	}
}

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

func (s *Service) cmdCMC(client *gomatrix.Client, roomID, userID string, args []string) (*gomatrix.HTMLMessage, error) {
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

	for _, arg := range args {
		coinID, err := findCoinID(arg, &allTickers)
		if err != nil {
			return nil, err // TODO handle gradefully
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

type ticker struct {
	Last string `json:"last"`
}

func (s *Service) cmdHitBTC(client *gomatrix.Client, roomID, userID, query string) (*[]byte, error) {
	log.Info("querying HitBTC for ", query)

	url := "https://api.hitbtc.com/api/2/public/" + query
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

// Match message with bad words. Constuct the pattern such that it matches only
// once per message so that it responds only once. Otherwise it seems to respond
// once per match.
var badwordsRegex = regexp.MustCompile(`(?i:^.*\b(gevers|guido|tzlibre)\b.*$)`)
//var badwordsRegex = regexp.MustCompile(`(?i:^(?:.*?\b(gevers|guido|tzlibre)\b)+.*?$)`)

var badwordsExpand = types.Expansion{
	Regexp: badwordsRegex,

	Expand: func(roomID, userID string, matches []string) interface{} {
		log.WithFields(log.Fields{"room_id": roomID, "user_id": userID, "matches": matches}).Print("badwords matched")
		return &gomatrix.TextMessage{
			"m.notice",
			fmt.Sprintf("%s used a bad word", userID),
		}
	},
}

/*
func (s *Service) Expansions(cli *gomatrix.Client) []types.Expansion {
	return []types.Expansion{
		badwordsExpand,
	}
}
*/

func init() {
	types.RegisterService(func(serviceID, serviceUserID, webhookEndpointURL string) types.Service {
		return &Service{
			DefaultService: types.NewDefaultService(serviceID, serviceUserID, ServiceType),
		}
	})
}
