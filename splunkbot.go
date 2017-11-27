// vim: set tabstop=2 expandtab autoindent smartindent:

package main

import (
  "os"
  "fmt"
  "time"
  "bytes"
  "io/ioutil"
  "strings"
  "net/url"
  "net/http"
  "crypto/tls"
  "encoding/json"
  "github.com/jessevdk/go-flags"
  log "github.com/sirupsen/logrus"
)

type SpunkHECMessage struct {
  Time                string          `json:"time,omitempty"`
  Host                string          `json:"host,omitempty"`
  Source              string          `json:"source,omitempty"`
  Sourcetype          string          `json:"sourcetype,omitempty"`
  Index               string          `json:"index,omitempty"`
  Event               interface{}     `json:"event"`
}

type Options struct {
  Verbose             []bool          `short:"v" long:"verbose" description:"Show verbose debug information"`
  ListeningAddress    string          `short:"a" long:"address" description:"Listening address" env:"SPLUNKBOT_LISTENING_ADDRESS" default:"127.0.0.1"`
  ListeningPort       uint            `short:"p" long:"port" description:"Listening port" env:"SPLUNKBOT_LISTENING_PORT" default:"44553"`
  SplunkUrl           string          `short:"u" long:"splunk-url" description:"Splunk HEC endpoint" env:"SPLUNKBOT_SPLUNK_URL" required:"true"`
  SplunkHTTPTimeout   uint            `short:"n" long:"splunk-timeout" description:"Splunk HEC timeout (seconds)" env:"SPLUNKBOT_SPLUNK_HTTP_TIMEOUT" default:"5"`
  SplunkToken         string          `short:"t" long:"splunk-token" description:"Splunk HEC token" env:"SPLUNKBOT_SPLUNK_TOKEN" required:"true"`
  SplunkIndex         string          `short:"i" long:"splunk-index" description:"Splunk index" env:"SPLUNKBOT_SPLUNK_INDEX"`
  SplunkSourcetype    string          `short:"s" long:"splunk-sourcetype" description:"Splunk event sourcetype" env:"SPLUNKBOT_SPLUNK_SOURCETYPE" required:"true" default:"alertmanager"`
  SplunkTLSInsecure   bool            `short:"k" long:"insecure" description:"Do not check Splunk TLS certificate"`
}

type Splunkbot struct {
  httpClient          *http.Client
}

func (sbot Splunkbot) serve() error {
  http.HandleFunc("/", sbot.alert)
  err := http.ListenAndServe(fmt.Sprintf("%s:%d", opts.ListeningAddress, opts.ListeningPort), nil)

  return err
}

func (s Splunkbot) alert(w http.ResponseWriter, r *http.Request) {
  log.Debugf("New request: %v", r)

  var alert map[string]interface{}
  var message SpunkHECMessage

  // Decode input
  buf, _  := ioutil.ReadAll(r.Body)
  err     := json.Unmarshal(buf, &alert)

  // if buf is not valid json we cast it as string
  if err != nil {
    message.Event = interface{}(string(buf))
  } else {
    message.Event = alert
  }

  // Splunk Message 
  message.Sourcetype  = opts.SplunkSourcetype
  message.Index       = opts.SplunkIndex

  if value, ok := alert["externalURL"]; ok {
    u, _ := url.Parse(value.(string))
    message.Host    = u.Hostname()
    message.Source  = strings.TrimLeft(u.Path, "/")
  }

  // HTTP Splunk request
  j, _  := json.Marshal(message)
  jr    := bytes.NewReader(j)

  splunkReq, _ := http.NewRequest("POST", opts.SplunkUrl, jr)
  splunkReq.Header.Set("Authorization", "Splunk " + opts.SplunkToken)

  // Do request
  resp, err := s.httpClient.Do(splunkReq)

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

var opts Options
var parser = flags.NewParser(&opts, flags.Default)

func init() {
  // Log as JSON instead of the default ASCII formatter.
  //log.SetFormatter(&log.JSONFormatter{})
  log.SetFormatter(&log.TextFormatter{})

  // Output to stdout instead of the default stderr
  // Can be any io.Writer, see below for File example
  log.SetOutput(os.Stdout)

  // Only log the warning severity or above.
  log.SetLevel(log.DebugLevel)
}

// main
func main() {
  if _, err := parser.Parse(); err != nil {
    if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
      os.Exit(0)
    } else {
      log.Fatal(err)
      os.Exit(1)
    }
  }

  // Update logging level
  switch  {
    case len(opts.Verbose) >= 1:
      log.SetLevel(log.DebugLevel)
    default:
      log.SetLevel(log.InfoLevel)
  }

  // Loggin options
  log.Debugf("Options: %+v", opts)

  // Starting server
  log.Infof("Starting server at http://%s:%v", opts.ListeningAddress, opts.ListeningPort)

  // HTTP Transport
  tr := &http.Transport{
    TLSClientConfig:  &tls.Config{
      InsecureSkipVerify: opts.SplunkTLSInsecure,
    },
  }

  // HTTP Client
  client := &http.Client{
    Timeout:          time.Duration(time.Duration(opts.SplunkHTTPTimeout) * time.Second),
    Transport:        tr,
  }

  // Splunkbot
  sbot := Splunkbot{
    httpClient:       client,
  }

  // Serving
  err := sbot.serve()

  if err != nil {
    log.Fatal(err)
    os.Exit(1)
  }
}
