package main

import (
  "encoding/json"
  "fmt"
  "github.com/Mischanix/applog"
  "net"
  "net/http"
)

var serverListener *net.Listener
var currentUrl = ""
var currentPort = -1

type postReceiveData struct {
  Repository struct {
    Name string `json:"name"`
  } `json:"repository"`
}

func unhandler(w http.ResponseWriter, req *http.Request) {
  w.Write([]byte("nok"))
}

func handler(w http.ResponseWriter, req *http.Request) {
  applog.Info(
    "server.handler: request: source=%s, method=%s, length=%d",
    req.RemoteAddr, req.Method, req.ContentLength,
  )
  if req.Method != "POST" {
    w.Write([]byte("nok: method not supported"))
    return
  }
  d := json.NewDecoder(req.Body)
  var hookData postReceiveData
  err := d.Decode(&hookData)
  if err != nil {
    applog.Error("server.handler: json Decode error: %v", err)
    w.Write([]byte("nok: invalid json"))
    return
  }
  if hookData.Repository.Name != "" {
    go func() {
      repoUpdate <- hookData.Repository.Name
    }()
    w.Write([]byte("ok"))
  } else {
    applog.Warn("server.handler: invalid data")
    w.Write([]byte("nok: invalid data"))
  }
}

func updateServer() {
  if config.HookUrl == "" {
    applog.Error("server: hook_url cannot be empty string")
    currentUrl = ""
    return
  }
  if currentUrl == config.HookUrl && currentPort == config.Port {
    return
  }
  if currentPort != config.Port && serverListener != nil {
    (*serverListener).Close()
    serverListener = nil
  }
  if serverListener == nil {
    l, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
    serverListener = &l
    currentPort = config.Port
    if err != nil {
      applog.Error(
        "server: net.Listen(:%d) failed: %v", config.Port, err)
      return
    }
    go func() {
      http.Serve(*serverListener, nil)
    }()
  }

  // Reset handlers
  http.DefaultServeMux = http.NewServeMux()
  http.HandleFunc(config.HookUrl, handler)
  currentUrl = config.HookUrl

  applog.Info("server: listening at :%d%s", currentPort, currentUrl)
}
