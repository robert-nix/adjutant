package main

import (
  "github.com/Mischanix/applog"
  "os"
  "os/exec"
  "strings"
  "time"
)

type cmdLine struct {
  path string
  args []string
}

var repositories map[string]cmdLine
var daycare map[string]*exec.Cmd

func updateDaycare() {
  if repositories == nil {
    repositories = make(map[string]cmdLine)
  }
  if daycare == nil {
    daycare = make(map[string]*exec.Cmd)
  }
  newRepos := make(map[string]cmdLine)
  changed := make([]string, 0)
  // Add new/updated repos
  for _, repoData := range config.Repositories {
    if line, ok := repositories[repoData.Name]; !ok ||
      (ok && line.path != repoData.CmdPath ||
        strings.Join(line.args, " ") != strings.Join(repoData.Args, " ")) {

      changed = append(changed, repoData.Name)
    }
    newRepos[repoData.Name] = cmdLine{repoData.CmdPath, repoData.Args}
  }
  // Remove repos no longer in config
  for name, _ := range repositories {
    if _, ok := newRepos[name]; !ok {
      go undeploy(name)
    }
  }
  repositories = newRepos // Copy new repositories map
  for _, name := range changed {
    if name == "" {
      applog.Error("empty string name\n%v", config)
    }
    go redeploy(name)
  }
}

// Called from main thread
func redeploy(name string) {
  if _, ok := daycare[name]; ok {
    undeploy(name)
  }

  if line, ok := repositories[name]; ok {
    cmd := exec.Command(line.path, line.args...)
    err := cmd.Start()
    if err != nil {
      applog.Error("daycare: start failed for %s: %v", name, err)
    } else {
      applog.Info("daycare: starting %s", name)
      daycare[name] = cmd

      go func() {
        err := cmd.Wait()
        if err != nil {
          applog.Warn("daycare: %s cmd.Wait: %v", name, err)
        } else {
          applog.Info("daycare: %s completed successfully", name)
        }
      }()
    }
  } else {
    applog.Error("daycare: no deploy was found for %s", name)
  }
}

func undeploy(name string) {
  killProcess(name)
}

func isPresent(cmd *exec.Cmd) bool {
  if cmd.Process == nil ||
    cmd.ProcessState != nil && cmd.ProcessState.Exited() {

    return false
  } else {
    return true
  }
}

func killProcess(name string) {
  timeout := time.Duration(config.KillTimeout) * time.Millisecond

  if cmd, ok := daycare[name]; ok {
    applog.Info("daycare: killing %s", name)
    delete(daycare, name)

    if isPresent(cmd) {
      cmd.Process.Signal(os.Interrupt)
    } else {
      // Warning since although the process could have gracefully terminated,
      // adjutant is meant for daemons where that behavior is unexpected.
      applog.Warn("daycare.kill: process for %s nil\ncmd: %v",
        name, cmd,
      )
      return
    }
    <-time.After(timeout)
    if isPresent(cmd) {
      applog.Warn("daycare.kill: force killing process for %s after %v",
        name, timeout,
      )
      cmd.Process.Kill()
    }
  } else {
    applog.Error("daycare.kill: could not find Cmd to kill %s", name)
  }
}
