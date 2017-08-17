package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/apex/log"
	clihander "github.com/apex/log/handlers/cli"
	"github.com/urfave/cli"
)

var (
	ctx *log.Entry
	// Version stores the plugin's version
	Version string
	// BuildTime stores the plugin's build time
	BuildTime string
	// IndexDomain is the index domain
	IndexDomain string
	// RegistryDomain is the registry domain
	RegistryDomain string
	// ImageName is the docker image to pull
	ImageName string
	// ImageTag is the docker image tag to pull
	ImageTag string
	// creds
	user   string
	passwd string
)

// Manifest docker image manifest
type Manifest struct {
	Config   string
	Layers   []string
	RepoTags []string
}

var repositories map[string]map[string]string

func init() {
	log.SetHandler(clihander.Default)
}

func initRegistry(reposName string, insecure bool) *Registry {
	registry, err := NewRegistry(IndexDomain, RegistryDomain, insecure)
	if err != nil {
		ctx.Fatal(err.Error())
	}
	log.Info("getting auth token")
	err = registry.GetToken(user, passwd, reposName)
	if err != nil {
		ctx.Fatal(err.Error())
	}
	return registry
}

// CmdTags get docker image tags
func CmdTags(insecure bool) error {
	ctx.Infof("\033[1m%s\033[0m", "Initialize Registry")
	registry := initRegistry(ImageName, insecure)

	tags, err := registry.ReposTags(ImageName)
	if err != nil {
		return err
	}
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Println("- Repository:", tags.Name)
	fmt.Println("- Tags:")
	for _, v := range tags.Tags {
		fmt.Fprintf(w, "\t%s\n", v)
	}
	w.Flush()
	return nil
}

func createManifest(tempDir, confFile string, layerFiles []string) (string, error) {
	var manifestArray []Manifest
	// Create the file
	tmpfn := filepath.Join(tempDir, "manifest.json")
	out, err := os.Create(tmpfn)
	if err != nil {
		log.WithError(err).Error("create manifest JSON failed")
	}
	defer out.Close()

	m := Manifest{
		Config:   confFile,
		Layers:   layerFiles,
		RepoTags: []string{ImageName + ":" + ImageTag},
	}
	manifestArray = append(manifestArray, m)
	mJSON, err := json.Marshal(manifestArray)
	if err != nil {
		log.WithError(err).Error("marshalling manifest JSON failed")
	}
	// Write the body to JSON file
	_, err = out.Write(mJSON)
	if err != nil {
		log.WithError(err).Error("writing manifest JSON failed")
	}

	return tmpfn, nil
}

// DownloadImage downloads docker image
func DownloadImage(insecure bool) {
	// Get image manifest
	ctx.Infof("\033[1m%s\033[0m", "Initialize Registry")
	registry := initRegistry(ImageName, insecure)

	mF, err := registry.ReposManifests(ImageName, ImageTag)
	if err != nil {
		ctx.Fatal(err.Error())
	}

	dir, err := ioutil.TempDir("", "graboid")
	if err != nil {
		ctx.Fatal(err.Error())
	}
	defer os.RemoveAll(dir) // clean up

	log.Infof("\033[1m%s\033[0m", "GET CONFIG")
	cfile, err := registry.RepoGetConfig(dir, ImageName, mF)
	if err != nil {
		ctx.Fatal(err.Error())
	}

	log.Infof("\033[1m%s\033[0m", "GET LAYERS")
	lfiles, err := registry.RepoGetLayers(dir, ImageName, mF)
	if err != nil {
		ctx.Fatal(err.Error())
	}

	log.Infof("\033[1m%s\033[0m", "CREATE manifest.json")
	_, err = createManifest(dir, cfile, lfiles)
	if err != nil {
		ctx.Fatal(err.Error())
	}

	tarFile := fmt.Sprintf("%s.tar", strings.Replace(ImageName, "/", "_", 1))
	log.Infof("\033[1m%s:\033[0m \033[34m%s\033[0m", "CREATE docker image tarball", tarFile)
	err = tarFiles(dir, tarFile)
	if err != nil {
		ctx.Fatal(err.Error())
	}
	log.Infof("\033[1mSUCCESS!\033[0m")
}

var appHelpTemplate = `Usage: {{.Name}} {{if .Flags}}[OPTIONS] {{end}}COMMAND [arg...]

{{.Usage}}

Version: {{.Version}}{{if or .Author .Email}}
Author:{{if .Author}} {{.Author}}{{if .Email}} - <{{.Email}}>{{end}}{{else}}
  {{.Email}}{{end}}{{end}}
{{if .Flags}}
Options:
  {{range .Flags}}{{.}}
  {{end}}{{end}}
Commands:
  {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
  {{end}}
Run '{{.Name}} COMMAND --help' for more information on a command.
`

func main() {

	cli.AppHelpTemplate = appHelpTemplate
	app := cli.NewApp()

	app.Name = "graboid"
	app.Author = "blacktop"
	app.Email = "https://github.com/blacktop"
	app.Version = Version + ", BuildTime: " + BuildTime
	app.Compiled, _ = time.Parse("20060102", BuildTime)
	app.Usage = "Docker Image Downloader"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "verbose output",
		},
		cli.StringFlag{
			Name:        "index",
			Value:       "https://index.docker.io",
			Usage:       "override index endpoint",
			EnvVar:      "GRABOID_INDEX",
			Destination: &IndexDomain,
		},
		cli.StringFlag{
			Name:        "registry",
			Value:       "",
			Usage:       "override registry endpoint",
			EnvVar:      "GRABOID_REGISTRY",
			Destination: &RegistryDomain,
		},
		cli.BoolFlag{
			Name:  "insecure",
			Usage: "do not verify ssl certs",
		},
		cli.StringFlag{
			Name:        "user",
			Value:       "",
			Usage:       "registry username",
			EnvVar:      "GRABOID_USERNAME",
			Destination: &user,
		},
		cli.StringFlag{
			Name:        "password",
			Value:       "",
			Usage:       "registry password",
			EnvVar:      "GRABOID_PASSWORD",
			Destination: &passwd,
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "tags",
			Usage: "List image tags",
			Action: func(c *cli.Context) error {
				if c.Bool("verbose") {
					log.SetLevel(log.DebugLevel)
				}

				if c.Args().Present() {
					if strings.Contains(c.Args().First(), ":") {
						imageParts := strings.Split(c.Args().First(), ":")
						ImageName = imageParts[0]
						ImageTag = imageParts[1]
					} else {
						ImageName = c.Args().First()
						ImageTag = "latest"
					}

					ctx = log.WithFields(log.Fields{
						"domain": IndexDomain,
						"image":  ImageName,
						"tag":    ImageTag,
					})
					return CmdTags(c.Bool("insecure"))
				}
				return errors.New("please supply a image:tag to pull")
			},
		},
	}
	app.Action = func(c *cli.Context) error {

		if c.Bool("verbose") {
			log.SetLevel(log.DebugLevel)
		}

		if c.Args().Present() {
			if strings.Contains(c.Args().First(), ":") {
				imageParts := strings.Split(c.Args().First(), ":")
				ImageName = imageParts[0]
				ImageTag = imageParts[1]
			} else {
				ImageName = c.Args().First()
				ImageTag = "latest"
			}

			ctx = log.WithFields(log.Fields{
				"domain": IndexDomain,
				"image":  ImageName,
				"tag":    ImageTag,
			})
			// downlad docker image as a tarball
			DownloadImage(c.Bool("insecure"))
		} else {
			cli.ShowAppHelp(c)
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err.Error())
	}
}
