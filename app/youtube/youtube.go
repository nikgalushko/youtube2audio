package youtube

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	videoInfoURL = "http://www.youtube.com/get_video_info?&video_id="
)

// Video info
type Video struct {
	ID, Title, Author, Keywords string
	AvgRating                   float32
	ViewCount                   int
	Duration                    time.Duration
	Formats                     []Format
}

// Format is info about video type
type Format struct {
	Itag               int
	Type, Quality, URL string
}

// NewFromURL return info about video by link
func NewFromURL(link *url.URL) (Video, error) {
	//parse url
	videoID := link.Query().Get("v")
	if len(videoID) < 10 {
		return Video{}, fmt.Errorf("youtube link %s is invalid", link.String())
	}
	resp, err := http.Get(videoInfoURL + videoID)

	if err != nil {
		return Video{}, err
	}
	defer resp.Body.Close()

	rawData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Video{}, err
	}
	resp.Body.Close()

	data := string(rawData)
	u, err := url.Parse("?" + data)
	if err != nil {
		return Video{}, err
	}

	query := u.Query()

	if query.Get("errorcode") != "" || query.Get("status") == "fail" {
		return Video{}, errors.New(query.Get("reason"))
	}

	video := Video{
		ID:       videoID,
		Title:    query.Get("title"),
		Author:   query.Get("author"),
		Keywords: query.Get("keywords"),
	}

	v, _ := strconv.Atoi(query.Get("view_count"))
	video.ViewCount = v

	r, _ := strconv.ParseFloat(query.Get("avg_rating"), 32)
	video.AvgRating = float32(r)

	seconds, _ := strconv.Atoi(query.Get("length_seconds"))
	video.Duration = time.Duration(seconds) * time.Second

	formatParams := strings.Split(query.Get("url_encoded_fmt_stream_map"), ",")

	// every video has multiple format choices. collate the list.
	for _, f := range formatParams {
		furl, _ := url.Parse("?" + f)
		fquery := furl.Query()

		itag, _ := strconv.Atoi(fquery.Get("itag"))

		video.Formats = append(video.Formats, Format{
			Itag:    itag,
			Type:    fquery.Get("type"),
			Quality: fquery.Get("quality"),
			URL:     fquery.Get("url") + "&signature=" + fquery.Get("sig"),
		})
	}

	return video, nil
}
