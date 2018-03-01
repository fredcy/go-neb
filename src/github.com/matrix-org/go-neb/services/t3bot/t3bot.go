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

var topicMessage = "This would be the <i>topic</i> message."

var topicHTMLMessage = gomatrix.HTMLMessage{
    MsgType: "m.notice",
    Body: "This would be the topic message.",
    Format: "org.matrix.custom.html",
    FormattedBody: "This would be the <strong>topic message</strong>.",
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
			Path: []string{"topic"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				//return &gomatrix.TextMessage{"m.notice", topicMessage}, nil
				return topicHTMLMessage, nil
			},
		},
	}
}


// Match message with bad words. Constuct pattern that it matches only
// once per message so that respond only once.
var badwordsRegex = regexp.MustCompile(`(?i:^.*\b(gevers|rumplestiltskin)\b.*$)`)

var badwordsExpand = types.Expansion{
    Regexp: badwordsRegex,

    Expand: func(roomID, userID string, matches []string) interface{} {
    	 log.WithFields(log.Fields{"room_id": roomID, "user_id": userID, "matches": matches}).Print("badwords matched")
	 return &gomatrix.TextMessage{
	 	"m.notice",
		fmt.Sprintf("%s used bad words", userID),
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
