package main

import (
  "github.com/Mischanix/applog"
  "github.com/Mischanix/evconf"
  "os"
)

var config struct {
  Repositories []struct {
    Name    string   `json:"name"`
    CmdPath string   `json:"cmd_path"`
    Args    []string `json:"args"`
  } `json:"repositories"`
  // Must not be empty string
  HookUrl     string `json:"hook_url"`
  Port        int    `json:"port"`
  KillTimeout int    `json:"kill_timeout"` // in milliseconds
}

func defaultConfig() {
  config.HookUrl = "/"
  config.Port = 9171
  config.KillTimeout = 5000
}

const logStdout = true

var stop = make(chan bool)
var configUpdate = make(chan bool)
var repoUpdate = make(chan string)

func main() {
  // Log setup
  applog.Level = applog.InfoLevel
  if logStdout {
    applog.SetOutput(os.Stdout)
  } else {
    if logFile, err := os.OpenFile(
      "adjutant.log",
      os.O_WRONLY|os.O_CREATE|os.O_APPEND,
      os.ModeAppend|0666,
    ); err != nil {
      applog.SetOutput(os.Stdout)
      applog.Error("Unable to open log file: %v", err)
    } else {
      applog.SetOutput(logFile)
    }
  }

  // Config setup
  conf := evconf.New("adjutant.json", &config)
  conf.OnLoad(func() {
    configUpdate <- true
  })
  defaultConfig()
  go func() {
    conf.Ready()
  }()

  // Event loop
  for {
    select {
    case <-stop:
      return
    case <-configUpdate:
      updateServer()
      updateDaycare()
    case repoName := <-repoUpdate:
      redeploy(repoName)
    }
  }
}
