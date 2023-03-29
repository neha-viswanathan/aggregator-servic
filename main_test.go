package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestMultiShops(t *testing.T) {
	defer gock.Off()

	url := "http://localhost:8080"
	gock.New(url).
		Get("/shops").
		Reply(200).
		BodyString(fmt.Sprintf(`shop1,%s/shops/shop1/flavors
shop2,%s/shops/shop2/flavors
`, url, url))
	gock.New(url).
		Get("/shops/shop1/flavors").
		Reply(200).
		BodyString("rum raisin,5\ncherry garcia,2\n")
	gock.New(url).
		Get("/shops/shop2/flavors").
		Reply(200).
		BodyString("rum raisin,2\nchunky monkey,1\n")

	testServer := start(":8081", url+"/shops")
	defer testServer.Shutdown(context.Background())
	time.Sleep(100 * time.Millisecond) // Make sure server has enough time to do an initial fetch

	c := http.Client{Transport: &http.Transport{}} // Do not use the default http Transport, as `gock` will intercept it
	res, _ := c.Get("http://localhost:8081/flavors")
	defer res.Body.Close()
	payload, _ := ioutil.ReadAll(res.Body)
	require.ElementsMatch(t, []string{"chunky monkey,1", "cherry garcia,2", "rum raisin,7", ""}, strings.Split(string(payload), "\n"), "failed to get all flavors")
}
