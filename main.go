// vim: set tabstop=2 expandtab autoindent smartindent:

package main

import (
  "os"
  "time"
  "net/http"
  "crypto/tls"
  "github.com/sylr/alertmanager-splunkbot/splunkbot"
  "github.com/jessevdk/go-flags"
  log "github.com/sirupsen/logrus"
)

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

var (
  opts Options
  parser  = flags.NewParser(&opts, flags.Default)
  version = "v0.0.0-dev"
)

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
  log.Infof("Version: %s", version)
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
  sbot := &splunkbot.Splunkbot{
    HttpClient:       client,
    ListeningAddress: opts.ListeningAddress,
    ListeningPort:    opts.ListeningPort,
    SplunkSourcetype: opts.SplunkSourcetype,
    SplunkIndex:      opts.SplunkIndex,
    SplunkUrl:        opts.SplunkUrl,
    SplunkToken:      opts.SplunkToken,
  }

  // Serving
  err := sbot.Serve()

  if err != nil {
    log.Fatal(err)
    os.Exit(1)
  }
}
