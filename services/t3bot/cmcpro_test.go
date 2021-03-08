package t3bot

import (
	"testing"
)

func TestQueryPro(t *testing.T) {
	bytes, err := queryCMCPro("cryptocurrency/listings/latest")
	if err != nil {
		t.Fatalf("queryCMCPro returned error: %v", err)
	}

	if len(*bytes) == 0 {
		t.Fatal("queryCMCPro returned empty bytes")
	}

	//t.Logf("bytes: %v", string(*bytes))
}

func TestQueryProMap(t *testing.T) {
	idmap, err := getCmcProMap()
	if err != nil {
		t.Fatalf("getCmcProMap returned error: %v", err)
	}

	//t.Logf("map: %v", *idmap)

	if len(*idmap) == 0 {
		t.Fatalf("Empty idmap")
	}

	id, err := findCoinIDPro("btc", idmap)
	if err != nil || id != 1 {
		t.Errorf("failed to find btc: %v, %v", id, err)
	}

	id, err = findCoinIDPro("xtz", idmap)
	if err != nil || id != 2011 {
		t.Errorf("failed to find xtz: %v, %v", id, err)
	}
}

func TestCmcProListings(t *testing.T) {
	listings, err := getCmcProListings()
	if err != nil {
		t.Fatalf("getCmcProListings: %v", err)
	}

	if len(*listings) == 0 {
		t.Fatalf("No listings")
	}

	for _, listing := range *listings {
		if listing.Id == 2011 {
			t.Logf("listing: %v", listing)
		}
	}
}
