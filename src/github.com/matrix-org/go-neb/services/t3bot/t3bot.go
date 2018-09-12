// Package t3bot implements a Service for Tezos room !commands.
package t3bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/matrix-org/go-neb/types"
	"github.com/matrix-org/gomatrix"
)

// ServiceType of the T3bot service
const ServiceType = "t3bot"

// Service represents the T3bot service. It has no Config fields.
type Service struct {
	types.DefaultService
	TezosRank string
}

var roomsMessageHTML = `
General &nbsp;&nbsp;&nbsp; <a href="https://riot.im/app/#/room/#tezos:matrix.org">#tezos:matrix.org</a><br>
Tech &nbsp;&nbsp;&nbsp; <a href="https://riot.im/app/#/room/#freenode_#tezos:matrix.org">#freenode_#tezos:matrix.org</a><br>
Price &nbsp;&nbsp;&nbsp; <a href="https://riot.im/app/#/room/#tezostrader:matrix.org">#tezostrader:matrix.org</a><br>
Media &nbsp;&nbsp;&nbsp; <a href="https://riot.im/app/#/room/#tezosmedia:matrix.org">#tezosmedia:matrix.org</a><br>
Random &nbsp;&nbsp;&nbsp; <a href="https://riot.im/app/#/room/#tezosrandom:matrix.org">#tezosrandom:matrix.org</a><br>
Philosophy &nbsp;&nbsp;&nbsp; <a href="https://riot.im/app/#/room/#tezosphilosophy:matrix.org">#tezosphilosophy:matrix.org</a><br>
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
https://www.tezosfoundation.ch<br>
https://www.reddit.com/user/TezosReddit<br>
https://twitter.com/TezosFoundation<br>
https://tezos.com<br>
`

var devSitesMessageHTML = `
Developer sites:<br>
<a href="http://doc.tzalpha.net">Core dev docs</a></br>
<a href="https://gitlab.com/tezos/tezos">Core dev repo</a><br>
<a href="https://github.com/tezoscommunity/FAQ/blob/master/Compile_Betanet.md">Compile Betanet</a><br>
<a href="https://github.com/tezoscommunity/FAQ/blob/master/Rebuilding_Betanet.md">Rebuilding Betanet</a><br>
<a href="https://gist.github.com/dakk/bdf6efe42ae920acc660b20080a506dd">Baking howto</a><br>
<a href="https://github.com/obsidiansystems/ledger-app-tezos/blob/master/README.md">Ledger Applications</a><br>
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

var devSitesHTMLMessage = gomatrix.HTMLMessage{
	MsgType:       "m.notice",
	Format:        "org.matrix.custom.html",
	FormattedBody: devSitesMessageHTML,
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

	text, err = html2text.FromString(devSitesMessageHTML, html2text.Options{OmitLinks: true})
	if err != nil {
		panic(err)
	}
	devSitesHTMLMessage.Body = text
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
			Path: []string{"devsites"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return devSitesHTMLMessage, nil
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
			Path: []string{"cmc2"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				content, err := e.cmdCMC2(cli, roomID, userID, args)
				log.WithFields(log.Fields{"content": fmt.Sprintf("%v", content), "error": err}).Info("cmc2 response")
				return content, err
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

func (s *Service) OnPoll(cli *gomatrix.Client) time.Time {
	pollingInterval := 10 * time.Minute
	delayAfterReport := 60 * time.Minute

	next := time.Now().Add(pollingInterval)

	//return next   // STUB OFF

	response, err := queryCMC("tezos/")
	if err != nil {
		log.WithError(err).Error("queryCMC failed")
		return next
	}

	var ts []cmcTicker
	err2 := json.Unmarshal(*response, &ts)
	if err2 != nil {
		log.WithError(err).Error("Unmarshal failed")
		return next
	}

	tezosConcernsRoom := "!mOcZCzWBxvtSxNvWzz:matrix.org"
	tezosTraderRoom := "!TUYwzSQkeKBLZlWldJ:matrix.org"
	tezosRandomRoom := "!xDsCezbpSVokOfGwCI:matrix.org"
	tezosRoom := "!KNlqwBRiVdbAwkVpKO:matrix.org"

	// Construct list of rooms in which to report a change in the Tezos CMC rank, along with the minimum rank to ever report for each room.
	rooms := []struct {
		RoomID string
		Limit  int
	}{
		{RoomID: tezosConcernsRoom, Limit: 2000},
		{RoomID: tezosRandomRoom, Limit: 1000},
		{RoomID: tezosTraderRoom, Limit: 100},
		{RoomID: tezosRoom, Limit: 10},
	}

	// We only expect one ticker in the array, but we loop anyway.
	for _, ticker := range ts {
		if ticker.Rank == s.TezosRank {
			//log.WithFields(log.Fields{"rank": ticker.Rank}).Info("rank unchanged")
			continue
		} else {
			log.WithFields(log.Fields{
				"old": s.TezosRank,
				"new": ticker.Rank,
			}).Info("rank changed")
		}
		var messageText string
		if s.TezosRank != "" {
			messageText = fmt.Sprintf("XTZ rank at CMC is <b>%s</b> (was %s)",
				ticker.Rank, s.TezosRank)

			// longer poll after reporting, to reduce thrashing
			next = time.Now().Add(delayAfterReport)
		} else {
			// first time querying (since we don't know prior value)

			//messageText = fmt.Sprintf("XTZ rank at CMC is <b>%s</b>",
			//	         ticker.Rank)
		}
		s.TezosRank = ticker.Rank

		if messageText == "" {
			continue
		}

		message := gomatrix.GetHTMLMessage("m.notice", messageText)

		rankI, err := strconv.Atoi(ticker.Rank)
		if err != nil {
			log.WithError(err).WithField("tickerRank", ticker.Rank).Error("Cannot convert rank to int")
		}

		for _, room := range rooms {
			if rankI > room.Limit {
				log.WithFields(log.Fields{
					"room_id": room.RoomID,
					"rank":    rankI,
				}).Info("ignoring high rank for room")
				continue
			}
			if _, err := cli.SendMessageEvent(room.RoomID, "m.room.message", message); err != nil {
				log.WithError(err).WithField("room_id", room.RoomID).Error("Failed to send to room")
			}
		}
	}

	return next
}

func init() {
	types.RegisterService(func(serviceID, serviceUserID, webhookEndpointURL string) types.Service {
		return &Service{
			DefaultService: types.NewDefaultService(serviceID, serviceUserID, ServiceType),
		}
	})
}
