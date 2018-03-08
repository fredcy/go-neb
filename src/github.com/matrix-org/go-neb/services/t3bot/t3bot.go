// Package t3bot implements a Service for Tezos room !commands.
package t3bot

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/matrix-org/go-neb/types"
	"github.com/matrix-org/gomatrix"
)

// ServiceType of the T3bot service
const ServiceType = "t3bot"

// Service represents the T3bot service. It has no Config fields.
type Service struct {
	types.DefaultService
}

var roomsMessage = `
Tezos General Chat: #tezos:matrix.org
Tezos Price Chat: #tezostrader:matrix.org
Tezos Media Chat: #tezosmedia:matrix.org
Tezos Governance Chat: #tezosgovernance:matrix.org
Tezos Ideas and Collaboration Chat: #tezosfoundry:matrix.org
Tezos IRC Tech Chat: #freenode_#tezos:matrix.org
Tezos Random Chat: #tezosrandom:matrix.org
`

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
	Body:          roomsMessage,
	Format:        "org.matrix.custom.html",
	FormattedBody: roomsMessageHTML,
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
	}
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
