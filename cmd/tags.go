/*
Copyright © 2019 blacktop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"strings"

	"github.com/apex/log"
	"github.com/blacktop/graboid/pkg/registry"
	"github.com/blacktop/ipsw/utils"
	"github.com/spf13/cobra"
)

func initRegistry(reposName string, insecure bool) *registry.Registry {
	config := registry.Config{
		Endpoint:       IndexDomain,
		RegistryDomain: RegistryDomain,
		Proxy:          "",
		// Proxy:          Proxy,
		Insecure: insecure,
		RepoName: reposName,
		Username: "",
		Password: "",
		// Username:       user,
		// Password:       passwd,
	}
	registry, err := registry.New(config)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Debug("getting auth token")
	err = registry.GetToken()
	if err != nil {
		log.Fatal(err.Error())
	}
	return registry
}

// tagsCmd represents the tags command
var tagsCmd = &cobra.Command{
	Use:   "tags [docker/image]",
	Short: "List image tags",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		if Verbose {
			log.SetLevel(log.DebugLevel)
		}
		insecure, _ := cmd.Flags().GetBool("insecure")

		registry := initRegistry(strings.Split(args[0], ":")[0], insecure)

		tags, err := registry.ReposTags(strings.Split(args[0], ":")[0])
		if err != nil {
			return err
		}

		log.WithFields(log.Fields{
			"image": tags.Name,
		}).Infof(getFmtStr(), "Querying Registry")

		utils.Indent(log.Info, 1)("Tags:")
		for _, v := range tags.Tags {
			utils.Indent(log.Info, 2)(v)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tagsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tagsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tagsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	tagsCmd.Flags().String("proxy", "", "HTTP/HTTPS proxy")
	tagsCmd.Flags().Bool("insecure", false, "do not verify ssl certs")
}
