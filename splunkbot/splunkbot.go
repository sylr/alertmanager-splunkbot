// vim: set tabstop=4 expandtab autoindent smartindent:

package splunkbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type SpunkHECMessage struct {
	Time       string      `json:"time,omitempty"`
	Host       string      `json:"host,omitempty"`
	Source     string      `json:"source,omitempty"`
	Sourcetype string      `json:"sourcetype,omitempty"`
	Index      string      `json:"index,omitempty"`
	Event      interface{} `json:"event"`
}

type Splunkbot struct {
	HttpClient       *http.Client
	ListeningAddress string
	ListeningPort    uint
	SplunkSourcetype string
	SplunkIndex      string
	SplunkUrl        string
	SplunkToken      string
}

func (sbot Splunkbot) Serve() error {
	http.HandleFunc("/", sbot.alert)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", sbot.ListeningAddress, sbot.ListeningPort), nil)

	return err
}

func (sbot Splunkbot) alert(w http.ResponseWriter, r *http.Request) {
	log.Debugf("New request: %v", r)

	var alert map[string]interface{}
	var message SpunkHECMessage

	// Decode input
	buf, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(buf, &alert)

	// if buf is not valid json we cast it as string
	if err != nil {
		message.Event = interface{}(string(buf))
	} else {
		message.Event = alert
	}

	// Splunk Message
	message.Sourcetype = sbot.SplunkSourcetype
	message.Index = sbot.SplunkIndex

	if value, ok := alert["externalURL"]; ok {
		u, _ := url.Parse(value.(string))
		message.Host = u.Hostname()
		message.Source = strings.TrimLeft(u.Path, "/")
	}

	// HTTP Splunk request
	j, _ := json.Marshal(message)
	jr := bytes.NewReader(j)

	splunkReq, _ := http.NewRequest("POST", sbot.SplunkUrl, jr)
	splunkReq.Header.Set("Authorization", "Splunk "+sbot.SplunkToken)

	// Do request
	resp, err := sbot.HttpClient.Do(splunkReq)

	if err != nil {
		log.Errorf("Failed to send request to splunk: %+v", err)

		if resp != nil {
			buf, _ := ioutil.ReadAll(resp.Body)
			w.WriteHeader(resp.StatusCode)
			w.Write(buf)
		} else {
			w.WriteHeader(503)
			w.Write([]byte(fmt.Sprintf("Something went wrong:\n\n%+v\n", err)))
		}
	} else {
		buf, _ := ioutil.ReadAll(resp.Body)
		w.WriteHeader(resp.StatusCode)
		w.Write(buf)
	}

	log.Debugf("End of request")
}
