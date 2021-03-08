package t3bot

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/matrix-org/gomatrix"
	"io/ioutil"
	"net/http"
)

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
