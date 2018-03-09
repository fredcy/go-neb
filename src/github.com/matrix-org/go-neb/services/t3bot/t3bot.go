// Package t3bot implements a Service for Tezos room !commands.
package t3bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/jaytaylor/html2text"
	"github.com/matrix-org/go-neb/types"
	"github.com/matrix-org/gomatrix"
)

// ServiceType of the T3bot service
const ServiceType = "t3bot"

// Service represents the T3bot service. It has no Config fields.
type Service struct {
	types.DefaultService
}

var roomsMessageHTML = `
Tezos General Chat: <a href="https://riot.im/app/#/room/#tezos:matrix.org">#tezos:matrix.org</a><br>
Tezos Price Chat: <a href="https://riot.im/app/#/room/#tezostrader:matrix.org">#tezostrader:matrix.org</a><br>
Tezos Media Chat: <a href="https://riot.im/app/#/room/#tezosmedia:matrix.org">#tezosmedia:matrix.org</a><br>
Tezos Governance Chat: <a href="https://riot.im/app/#/room/#tezosgovernance:matrix.org">#tezosgoverance:matrix.org</a><br>
Tezos Ideas and Collaboration Chat: <a href="https://riot.im/app/#/room/#tezosfoundry:matrix.org">#tezosfoundry:matrix.org</a><br>
Tezos IRC Tech Chat: <a href="https://riot.im/app/#/room/#freenode_#tezos:matrix.org">#freenode_#tezos:matrix.org</a><br>
Tezos Random Chat: <a href="https://riot.im/app/#/room/#tezosrandom:matrix.org">#tezosrandom:matrix.org</a>
`

var roomsHTMLMessage = gomatrix.HTMLMessage{
	MsgType:       "m.notice",
	Format:        "org.matrix.custom.html",
	FormattedBody: roomsMessageHTML,
}

func init() {
	text, err := html2text.FromString(roomsMessageHTML, html2text.Options{OmitLinks: true})
	if err != nil {
		panic(err)
	}
	roomsHTMLMessage.Body = text
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
			Path: []string{"mom-am-i-rich-yet"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return &gomatrix.TextMessage{"m.notice", "Not yet, dear boy. Go back to work."}, nil
			},
		},
	}
}

type cmcTickerResponse struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	PriceUSD string `json:"price_usd"`
	Rank     string `json:"rank"`
	Pct1H    string `json:"percent_change_1h"`
	Pct24H   string `json:"percent_change_24h"`
	Pct7D    string `json:"percent_change_7d"`
}

func (s *Service) cmdCMC(client *gomatrix.Client, roomID, userID string, args []string) (*gomatrix.TextMessage, error) {
	var coinID string
	if len(args) == 0 {
		coinID = "bitcoin"
	} else if len(args) == 1 {
		coinID = args[0]
	} else {
		coinID = args[0] // TODO: handle more?
	}

	response, err := queryCMC(coinID + "/")
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{"response": string(*response)}).Info("CMC response")

	var tickers []cmcTickerResponse
	err2 := json.Unmarshal(*response, &tickers)
	if err2 != nil {
		return nil, err // TODO hide from user
	}

	//fake := fmt.Sprintf("%v", tickers)
	textMessage := fmt.Sprintf("%s (%s): $%s  rank=%s 1h=%s%%",
		tickers[0].Symbol, tickers[0].Id, tickers[0].PriceUSD,
		tickers[0].Rank, tickers[0].Pct1H)

	message := gomatrix.TextMessage{"m.notice", textMessage}
	return &message, nil
}

func queryCMC(query string) (*[]byte, error) {
	log.Info("querying CMC for ", query)

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

// Match message with bad words. Constuct pattern that it matches only once per
// message so that it responds only once. Otherwise it seems to respond once per
// match.
var badwordsRegex = regexp.MustCompile(`(?i:^.*\b(gevers|guido|tzlibre)\b.*$)`)

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

func (s *Service) Expansions(cli *gomatrix.Client) []types.Expansion {
	return []types.Expansion{
		badwordsExpand,
	}
}

func init() {
	types.RegisterService(func(serviceID, serviceUserID, webhookEndpointURL string) types.Service {
		return &Service{
			DefaultService: types.NewDefaultService(serviceID, serviceUserID, ServiceType),
		}
	})
}
