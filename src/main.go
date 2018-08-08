package main

import (
  "fmt"
  "flag"  
  "os"  
//  "./updater"
)

// Prints usage on stdout
func usage() {
  fmt.Println("Usage of", os.Args[0])
  fmt.Println()
  flag.PrintDefaults()
}

// Entry point of the program
func main() {
  fmt.Println("Docker image updater")
  forcePtr := flag.Bool("force", false, "Force rebuild")
  filePtr := flag.String("file", "", "Path to the configuration file")
  apiPtr := flag.String("api", "", "Docker API version")

  flag.Parse();

  //fmt.Println("Configuration file path", *filePtr)
  //fmt.Println("Force rebuild: ", *forcePtr)

  if len(*filePtr) == 0 {
    fmt.Fprintln(os.Stderr, "Error: missing \"file\" parameter")
    usage()
    os.Exit(1)
  }

  if len(*apiPtr) == 0 {
    os.Setenv("DOCKER_API_VERSION", "1.30")
  } else {
    os.Setenv("DOCKER_API_VERSION", *apiPtr)
  }

  docks, err := Parse(*filePtr)
  
  if err != nil {
    fmt.Fprintln(os.Stderr, "Error: ", err)
    os.Exit(1)
  }

  tree := GetDockerImageOptMap(docks)
  err = UpdateDockerImage(tree, *forcePtr)

  if err != nil {
    fmt.Fprintln(os.Stderr, "Error occurred:", err);
  } else {
    fmt.Println("Docker image updater finished")
  }
}

