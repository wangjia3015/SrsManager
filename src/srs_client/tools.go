package srs_client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	URL_STREAMS_PATH = "api/v1/streams"
	URL_CLIENTS_PATH = "api/v1/clients"

	HTTP_GET    = "GET"
	HTTP_PUT    = "PUT"
	HTTP_DELETE = "DELETE"
)

func sendRequest(method, url string) (int, []byte, error) {
	code := -1
	var body []byte
	var err error
	client := http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return code, body, err
	}

	//req.Header.Add("Authorization", AuthString)
	rsp, err := client.Do(req)
	if err != nil {
		return code, body, err
	}
	defer rsp.Body.Close()
	code = rsp.StatusCode
	body, err = ioutil.ReadAll(rsp.Body)
	return code, body, err
}

/*
streams: [
{
	id: 8186,
	name: "kanwo",
	vhost: 8185,
	app: "live",
	live_ms: 1465296710455,
	clients: 2,
	send_bytes: 120168054,
	recv_bytes: 118344183,
	kbps: {
		recv_30s: 96,
		send_30s: 99
	},
	publish: {
		active: true,
		cid: 129
	},
	video: null,
	audio: null
}
]
*/

type KbpsInfo struct {
	Recv30s int `json:recv_30s`
	Send30s int `json:send_30s`
}

type Publisher struct {
	Active bool `json:active` // 是否工作
	CID    int  `json:cid`    // publisher ID
}

type Stream struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	VHost     int       `json:"vhost"`
	AppName   string    `json:"app"`
	LiveMs    int64     `json:"live_ms"`
	ClientNum int       `json:"clients"`
	SendBytes int64     `json:"send_bytes"`
	RecvBytes int64     `json:"recv_bytes"`
	Kbps      KbpsInfo  `json:"kshbps"`
	Publish   Publisher `json:"publish"`
}

type RspStream struct {
	Code     int      `json:"code"`
	ServerID int      `json:"server"`
	Streams  []Stream `json:"streams"`
}

func GetStreams(host string) (RspStream, error) {
	url := fmt.Sprintf("http://%s/%s", host, URL_STREAMS_PATH)
	code, body, err := sendRequest(HTTP_GET, url)

	var stream RspStream
	if code == http.StatusOK {
		err = json.Unmarshal(body, &stream)
	}
	return stream, err
}

type RspBase struct {
	Code int `json:"code"`
}

func KickOffClient(host string, clientID int64) (RspBase, error) {
	url := fmt.Sprintf("http://%s/%s/%d", host, URL_CLIENTS_PATH, clientID)
	fmt.Println(url)
	code, body, err := sendRequest(HTTP_DELETE, url)
	var rsp RspBase
	if code == http.StatusOK {
		err = json.Unmarshal(body, &rsp)
	}
	return rsp, err
}

/*
	ok: true,
	sample_time: 1465369557159,
	percent: 0.00334448,
	user: 9796,
	nice: 3,
	sys: 7234,
	idle: 1751474,
	iowait: 201,
	irq: 0,
	softirq: 1619,
	steal: 0,
	guest: 0
*/
//type RspSystemProcStats struct {
//}
//
//func GetSystemProcStats(host string) {
//
//}
