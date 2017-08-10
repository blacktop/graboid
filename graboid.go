package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/apex/log"
	clihander "github.com/apex/log/handlers/cli"
	humanize "github.com/dustin/go-humanize"
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
	// creds
	user   string
	passwd string
)

func init() {
	log.SetHandler(clihander.Default)
	ctx = log.WithFields(log.Fields{
		"file": "something.png",
		"type": "image/png",
		"user": "tobi",
	})
}

func initRegistry(reposName string) (*Registry, string) {
	registry, err := NewRegistry(IndexDomain, RegistryDomain)
	if err != nil {
		ctx.Fatal(err.Error())
	}
	ctx.Infof("Getting token from %s", IndexDomain)
	token, err := registry.GetToken(user, passwd, reposName)
	if err != nil {
		ctx.Fatal(err.Error())
	}
	return registry, token
}

// CmdInfo get docker image metadata info
func CmdInfo(args []string) {
	registry, token := initRegistry(args[0])
	log.WithFields(log.Fields{
		"registry": registry,
		"token":    token,
	}).Info("initRegistry")
	tags, err := registry.ReposTags(token, args[0])
	if err != nil {
		log.Fatal(err.Error())
	}
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Println("- Repository:", args[0])
	fmt.Println("- Tags:")
	for k, v := range tags {
		fmt.Fprintf(w, "\t%s\t%s\n", k, v)
	}
	w.Flush()
}

// CmdLayerInfo gets docker image layer info
func CmdLayerInfo(args []string) {
	registry, token := initRegistry(args[0])
	info, err := registry.LayerJson(token, args[1])
	if err != nil {
		ctx.Fatal(err.Error())
	}
	ancestry, err := registry.LayerAncestry(token, args[1])
	if err != nil {
		ctx.Fatal(err.Error())
	}
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintf(w, "- Id\t%s\n", info.Id)
	fmt.Fprintf(w, "- Parent\t%s\n", info.Parent)
	fmt.Fprintf(w, "- Size\t%s\n", humanize.Bytes(uint64(info.Size)))
	fmt.Fprintf(w, "- Created\t%s\n", info.Created)
	fmt.Fprintf(w, "- DockerVersion\t%s\n", info.DockerVersion)
	fmt.Fprintf(w, "- Author\t%s\n", info.Author)
	fmt.Fprintf(w, "- Ancestry:")
	for _, id := range *ancestry {
		fmt.Fprintf(w, "\t%s\n", id)
	}
	w.Flush()
}

// CmdCurlme outputs curl command to pull image layer
func CmdCurlme(args []string) {
	registry, token := initRegistry(args[0])
	fmt.Printf("curl -i --location-trusted -I -X GET -H \"Authorization: Token %s\" %s/v1/images/%s/layer\n",
		token, registry.RegistryHost, args[1])
}

// DownloadImage downloads docker image
func DownloadImage(args []string) {
	// Create the file
	out, err := os.Create(fmt.Sprintf("%s.tar", args[0]))
	if err != nil {
		log.WithError(err).Error("create file failed")
	}
	defer out.Close()

	registry, token := initRegistry(args[0])
	url := fmt.Sprintf("%s/v1/images/%s/layer", registry.RegistryHost, args[1])

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))

	client := &http.Client{
	// CheckRedirect: redirectPolicyFunc,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer resp.Body.Close()

	log.WithFields(log.Fields{
		"status":   resp.Status,
		"size":     resp.ContentLength,
		"filepath": fmt.Sprintf("%s.tar", args[0]),
	}).Debug("downloading file")

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.WithError(err).Error("writing file failed")
	}
}

var appHelpTemplate = `Usage: {{.Name}} {{if .Flags}}[OPTIONS] {{end}}COMMAND [arg...]
{{.Usage}}
Version: {{.Version}}{{if or .Author .Email}}
Author:{{if .Author}}
  {{.Author}}{{if .Email}} - <{{.Email}}>{{end}}{{else}}
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
		cli.IntFlag{
			Name:   "timeout",
			Value:  60,
			Usage:  "elasticsearch timeout (in seconds)",
			EnvVar: "TIMEOUT",
		},
		cli.StringFlag{
			Name:        "index,i",
			Value:       "https://index.docker.io",
			Usage:       "override index endpoint",
			EnvVar:      "GRABOID_INDEX",
			Destination: &IndexDomain,
		},
		cli.StringFlag{
			Name:        "registry,r",
			Value:       "",
			Usage:       "override registry endpoint",
			EnvVar:      "GRABOID_REGISTRY",
			Destination: &RegistryDomain,
		},
		cli.StringFlag{
			Name:        "user,u",
			Value:       "",
			Usage:       "registry username",
			EnvVar:      "GRABOID_USERNAME",
			Destination: &user,
		},
		cli.StringFlag{
			Name:        "password, p",
			Value:       "",
			Usage:       "registry password",
			EnvVar:      "GRABOID_PASSWORD",
			Destination: &passwd,
		},
	}
	app.Action = func(c *cli.Context) error {

		if c.Bool("verbose") {
			log.SetLevel(log.DebugLevel)
		}

		if c.Args().Present() {
			CmdInfo(c.Args())
			CmdLayerInfo(c.Args())
			CmdCurlme(c.Args())
			DownloadImage(c.Args())
		} else {
			return errors.New("please supply a image to pull")
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		ctx.Fatal(err.Error())
	}
}
