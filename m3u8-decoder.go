package m3u8_decoder

import (
	"context"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type M3u8Ts struct {
	Url       string
	Duration  time.Duration
	StartTime time.Time
	EndTime   time.Time
}

type M3u8 struct {
	Version        int
	TSList         []M3u8Ts
	MediaSequence  int
	TargetDuration int
}

type M3u8Decoder struct {
	fn      func() (string, error)
	m3u8Url string
	context.Context
}

func NewM3u8Decoder(fn func() (string, error)) *M3u8Decoder {
	m3u8Url, _ := fn()
	return &M3u8Decoder{m3u8Url: m3u8Url, fn: fn}
}

func (decoder *M3u8Decoder) WithContext(ctx context.Context) *M3u8Decoder {
	return &M3u8Decoder{m3u8Url: decoder.m3u8Url, fn: decoder.fn, Context: ctx}
}

func (decoder *M3u8Decoder) refresh() error {
	m3u8Url, err := decoder.fn()
	if err != nil {
		return err
	}
	decoder.m3u8Url = m3u8Url
	return nil
}

func (decoder *M3u8Decoder) Content() (string, error) {
	if decoder.m3u8Url == "" {
		err := decoder.refresh()
		if err != nil {
			return "", err
		}
	}
	resp, err := http.Get(decoder.m3u8Url)
	if err != nil {
		return "", err
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func (decoder *M3u8Decoder) Decode() (M3u8, error) {
	content, err := decoder.Content()
	if err != nil {
		return M3u8{}, err
	}

	if content == "" {
		return M3u8{}, decoder.refresh()
	}
	kvList := strings.Split(content, "#")

	var newKvList []string
	for i := 0; i < len(kvList); i++ {
		if len(kvList[i]) == 0 {
			continue
		}
		tmp := strings.ReplaceAll(kvList[i], "\n", "")
		tmp = strings.ReplaceAll(tmp, "\r", "")
		newKvList = append(newKvList, tmp)
	}

	var m3u8 M3u8
	for _, v := range newKvList {
		kv := strings.Split(v, ":")
		if len(kv) != 2 {
			continue
		}

		switch kv[0] {
		case "EXT-X-VERSION":
			m3u8.Version, _ = strconv.Atoi(kv[1])
		case "EXT-X-MEDIA-SEQUENCE":
			m3u8.MediaSequence, _ = strconv.Atoi(kv[1])
		case "EXT-X-TARGETDURATION":
			m3u8.TargetDuration, _ = strconv.Atoi(kv[1])
		case "EXTINF":
			value := strings.Split(kv[1], ",")
			if len(value) != 2 {
				continue
			}

			var m3u8Ts M3u8Ts
			timeValue, _ := strconv.ParseFloat(value[0], 64)
			m3u8Ts.Duration = time.Second * time.Duration(timeValue)
			urlList := strings.Split(decoder.m3u8Url, "/")
			for i := 0; i < len(urlList); i++ {
				if i != len(urlList)-1 {
					m3u8Ts.Url = m3u8Ts.Url + urlList[i] + "/"
				} else {
					m3u8Ts.Url = m3u8Ts.Url + value[1]
				}
			}
			m3u8.TSList = append(m3u8.TSList, m3u8Ts)
		}
	}
	curTime := time.Now()
	for i := len(m3u8.TSList) - 1; i >= 0; i-- {
		m3u8.TSList[i].EndTime = curTime
		m3u8.TSList[i].StartTime = curTime.Add(-1 * m3u8.TSList[i].Duration)
		curTime = curTime.Add(-1 * m3u8.TSList[i].Duration)
	}
	return m3u8, nil
}

func (decoder *M3u8Decoder) StartDecode(callback func(M3u8Ts) error) error {
	for {
		select {
		case <-decoder.Context.Done():
			return nil
		default:
			m3u8, err := decoder.Decode()
			if err != nil {
				decoder.refresh()
			} else {
				var totalTime time.Duration
				for _, v := range m3u8.TSList {
					go callback(v)
					totalTime = totalTime + v.Duration
				}
				time.Sleep(totalTime)
			}
		}
	}

	return nil
}
