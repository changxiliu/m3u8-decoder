package m3u8_decoder_test

import (
	"context"
	"fmt"
	m3u8_decoder "github.com/changxiliu/m3u8-decoder"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func GetM3u8Url() (string, error) {
	return "https://bp1.dkkomo.com/stream/full/japan/3000/heyzo_2166.m3u8", nil
}

func m3u8DecodeCallback(m3u8Ts m3u8_decoder.M3u8Ts) error {
	fmt.Println(m3u8Ts.Url)
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
