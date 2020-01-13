package m3u8_decoder_test

import (
	"context"
	"encoding/json"
	"fmt"
	m3u8_decoder "github.com/changxiliu/m3u8-decoder"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

type LiveUrl struct {
	LiveUrl string `json:"liveUrl"`
}

func GetM3u8Url() (string, error) {
	resp, err := http.Get("") // TODO fix
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var liveUrl LiveUrl
	err = json.Unmarshal(body, &liveUrl)
	if err != nil {
		return "", err
	}

	m3u8Url := liveUrl.LiveUrl

	urlUrl, err := url.Parse(m3u8Url)
	if err != nil {
		return "", err
	}

	port := strings.Split(urlUrl.Host, ":")[1]
	urlUrl.Host = "119.3.175.106" + ":" + port
	return urlUrl.String(), err
}

func m3u8DecodeCallback(tsUrl string) error {
	fmt.Println(tsUrl)
	return nil
}

func TestNewM3u8Decoder(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		err := m3u8_decoder.NewM3u8Decoder(GetM3u8Url).WithContext(ctx).StartDecode(m3u8DecodeCallback)
		require.NoError(t, err)
		fmt.Println("StartDecode end")
	}()
	time.Sleep(time.Second * 30)
	cancelFunc()

	for {
		time.Sleep(time.Second)
	}
}
