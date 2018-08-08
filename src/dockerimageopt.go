package main

import (
  "fmt"
  "strings"
  "os"
  "bufio"
  "path/filepath"

  "github.com/docker/docker/client"
  "github.com/docker/docker/api/types"
  "github.com/jhoonb/archivex"
  "golang.org/x/net/context"
)

// DockerImageOpt structure
type DockerImageOpt struct {
  name string
  tag string
  fromImage string
  dockerfile string
  childrenImages []DockerImageOpt
}

// Constructs a new DockerImageOpt object.
func NewDockerImageOpt(name string, tag string, path string) (DockerImageOpt, error) {
  file, err := os.Open(path)

  if err != nil {
    return DockerImageOpt{}, err
  }

  var from string
  reader := bufio.NewReader(file)

  for {
    line, _, err := reader.ReadLine()

    ln := strings.Split(string(line), " ")

    if ln[0] == "FROM" && len(ln) > 1 {
      from = ln[1]
      break;
    }

    if err != nil {
      break
    }
  }

  file.Close()

  d := DockerImageOpt{
    name: name,
    tag: tag,
    fromImage: from,
    dockerfile: path,
  }

  return d,nil
}

// Returns image name.
func (d* DockerImageOpt) GetName() string {
  return d.name
}

// Returns image tag.
func (d* DockerImageOpt) GetTag() string {
  return d.tag
}

// Return image and tag.
func (d* DockerImageOpt) GetImage() string {
  return d.name + ":" + d.tag
}

// Return parent image (i.e. FROM).
func (d* DockerImageOpt) GetFromImage() string {
  return d.fromImage
}

// Returns DockerImageOpt path.
func (d* DockerImageOpt) GetDockerImageOpt() string {
  return d.dockerfile
}

// Returns children list.
func (d* DockerImageOpt) GetChildren() []DockerImageOpt {
  return d.childrenImages
}

// Returns whether or not the object is a children of another.
func (d* DockerImageOpt) IsChildrenOf(p DockerImageOpt) bool {
  return d.fromImage == p.GetImage()
}

// Adds children to object.
func (d* DockerImageOpt) AddChildren(child DockerImageOpt) {
  // check if it already contains the children image
  for _, dc := range d.childrenImages {
    if dc.fromImage == child.fromImage {
      return
    }
  }

  d.childrenImages = append(d.childrenImages, child)
}

func (d* DockerImageOpt) Print() {
  fmt.Println("FROM:", d.fromImage, "Children:", d.childrenImages)
}

func addChildren(docks []DockerImageOpt, idx int) {
  for i, _ := range docks {
    if idx != i && docks[idx].GetImage() == docks[i].GetFromImage() {
      // before adding it to children list, perform the same stuff for the 
      // children
      addChildren(docks, i)

      docks[idx].AddChildren(docks[i])
    }
  }
}

func GetDockerImageOptMap(docks []DockerImageOpt) map[string][]DockerImageOpt {
  // check if image depends on others
  for i, _ := range docks {
    addChildren(docks, i)
  }

  tree := make(map[string][]DockerImageOpt)

  // sort it as a tree (i.e. remove element if present as a child)
  for i, _ := range docks {
    isChild := false

    for j, _ := range docks {
      if i != j && docks[i].IsChildrenOf(docks[j]) {
        isChild = true
        break
      }
    }
    if isChild == false {
      tree[docks[i].GetFromImage()] = append(tree[docks[i].GetFromImage()],
        docks[i])
    }
  }

  return tree
}

// Builds an image (i.e. docker build).
func buildImage(d DockerImageOpt, cli *client.Client) (err error) {
  buildOptions := types.ImageBuildOptions {
    Tags: []string{d.GetImage()},
    SuppressOutput: true,
  }

  tar := new(archivex.TarFile)
  tarFileName := "./" + d.GetName() + "-" + d.GetTag() + ".tar"
  tar.Create(tarFileName)
  //fmt.Println(filepath.Dir(d.dockerfile))
  tar.AddAll("./" + filepath.Dir(d.dockerfile), false)
  tar.Close()

  buildContext, err := os.Open(tarFileName)
  if err != nil {
    return err
  }

  defer func() {
    fmt.Println("Remove tar file", buildContext.Name())
    buildContext.Close()
    os.Remove(buildContext.Name())
  }()

  fmt.Println("Build image:", d.GetImage())
  build, err := cli.ImageBuild(context.Background(), buildContext, buildOptions)

  if err != nil {
    return err
  }

  build.Body.Close()
  fmt.Println("Build finished:", d.GetImage())

  for _, c := range d.GetChildren() {
    // now build the children
    return buildImage(c, cli)
  }

  return nil
}

// Checks if children has to be rebuilt
func checkUpdate(children []DockerImageOpt, lastBaseLayer string,
cli *client.Client, images []types.ImageSummary, force bool) error {
  ctx := context.Background()
  var err error

  for i, _ := range children {
    isFound := false
    mustRebuild := true 
    var imgBaseLayer string

    fmt.Println("Should", children[i].GetImage(), "be rebuilt?")

    // check if image is already present in local registry
    for _, img := range images {
      for _, t := range img.RepoTags {
        if t == children[i].GetImage() {
          isFound = true

          // check now if image has last layer from its base
          inspect, _, err := cli.ImageInspectWithRaw(ctx, img.ID)

          if err != nil {
            return err
          }
          
          imgBaseLayer = inspect.RootFS.Layers[len(inspect.RootFS.Layers) - 1]

          for _, l := range inspect.RootFS.Layers {
            if l == lastBaseLayer {
              // got it so useless to rebuild
              mustRebuild = false
            }
          }
          break;
        }
      }
    }

    if !isFound || mustRebuild || force {
      err = buildImage(children[i], cli)

      if err != nil {
        fmt.Fprintln(os.Stderr, "Error during", children[i].GetImage())
        fmt.Fprintln(os.Stderr, "\t", err)
      }

    } else {
      fmt.Println("No need to rebuild", children[i].GetImage())
      checkUpdate(children[i].GetChildren(), imgBaseLayer, cli, images, force)
    }
  }

  return nil
}

// Updates docker images if their base image has changed.
func UpdateDockerImage(tree map[string][]DockerImageOpt, force bool) (err error ) {
  ctx := context.Background()
  cli, err := client.NewEnvClient()

  if err != nil {
    return err
  }

  var images []types.ImageSummary
  var lastBaseLayer string

  // - Pull each base image (key of the map)
  // - Children of each base image will be rebuilt if they have not the last
  // layer of their base image
  // - Each children of a "rebuilt" child will be rebuilt too
  // - And so on
  for k, v := range tree {
    // special case "scratch" is not an image to pull
    if k != "scratch" {
      fmt.Println("Pull image", k)

      out, err := cli.ImagePull(ctx, k, types.ImagePullOptions{})

      if err != nil {
        return err
      }
      out.Close()

      fmt.Println("Image", k, "pull finished")

      images, _ = cli.ImageList(ctx, types.ImageListOptions{})
      for _, img := range images {
        for _, t := range img.RepoTags {
          if t == k {
            inspectBase, _, err := cli.ImageInspectWithRaw(ctx, img.ID)

            if err != nil {
              return err
            }

            lastBaseLayer =
              inspectBase.RootFS.Layers[len(inspectBase.RootFS.Layers) - 1]
            break;
          }
        }
      }
    } else {
      // TODO should we need to rebuild scratch image ?
      continue
    }

    err = checkUpdate(v, lastBaseLayer, cli, images, force) 
    if err != nil {
      return err
    }
  }

  return nil 
}
