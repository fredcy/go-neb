package t3bot

import (
	"fmt"
	"github.com/matrix-org/go-neb/types"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"maunium.net/go/mautrix/id"
	"net/http"
)

type ticker struct {
	Last string `json:"last"`
}

func (s *Service) cmdHitBTC(client types.MatrixClient, roomID id.RoomID, userID id.UserID, query string) (*[]byte, error) {
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
