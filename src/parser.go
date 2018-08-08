package main

import (
  "os"
  "fmt"
  "github.com/BurntSushi/toml"
)

type dockerConfig struct {
  Name string
  Tag string
  Dockerfile string
}
  
// TOML configuration.
type tomlConfig struct {
  Images map[string]dockerConfig
}

// Parses configuration file and returns slice of DockerImageOpt
func Parse(filePath string) ([]DockerImageOpt, error) {
  var config tomlConfig
  if _, err := toml.DecodeFile(filePath, &config); err != nil {
    return []DockerImageOpt{}, err
  }

  docks := []DockerImageOpt{}
  for key, dc := range config.Images {
    d, err := NewDockerImageOpt(dc.Name, dc.Tag, dc.Dockerfile)

    if err != nil {
      fmt.Fprintln(os.Stderr, "Warning: error processing", key, "record")
      fmt.Fprintln(os.Stderr, "\t", err)
      continue
    }

    docks = append(docks, d)
  }

  return docks, nil
}

