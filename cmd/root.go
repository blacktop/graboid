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
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/apex/log"
	clihander "github.com/apex/log/handlers/cli"
	"github.com/blacktop/graboid/pkg/image"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var config struct {
	File string
	// Verbose boolean flag for verbose logging
	Verbose bool
	// IndexUrl is the index domain
	IndexUrl string
	// RegistryDomain is the registry domain
	RegistryDomain string
	// ImageName is the docker image to pull
	ImageName string
	// ImageTag is the docker image tag to pull
	ImageTag string

	// Registry Authentication
	Username string
	Password string
}

func getFmtStr() string {
	if runtime.GOOS == "windows" {
		return "%s"
	}
	return "\033[1m%s\033[0m"
}

func createManifest(tempDir, confFile string, layerFiles []string) (string, error) {
	var manifestArray []image.Manifest
	// Create the file
	tmpfn := filepath.Join(tempDir, "manifest.json")
	out, err := os.Create(tmpfn)
	if err != nil {
		log.WithError(err).Error("create manifest JSON failed")
	}
	defer out.Close()

	m := image.Manifest{
		Config:   confFile,
		Layers:   layerFiles,
		RepoTags: []string{config.ImageName + ":" + config.ImageTag},
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

func tarFiles(srcDir, tarName string) error {
	tarfile, err := os.Create(tarName)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	gw := gzip.NewWriter(tarfile)
	defer gw.Close()
	tarball := tar.NewWriter(gw)
	defer tarball.Close()

	return filepath.Walk(srcDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}
			if err = tarball.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			log.WithField("path", path).Debug("taring file")
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarball, file)
			return err
		})
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "graboid",
	Short: "Docker Image Downloader",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		if config.Verbose {
			log.SetLevel(log.DebugLevel)
		}
		insecure, _ := cmd.Flags().GetBool("insecure")

		if strings.Contains(args[0], ":") {
			imageParts := strings.Split(args[0], ":")
			config.ImageName = imageParts[0]
			config.ImageTag = imageParts[1]
		} else {
			config.ImageName = args[0]
			config.ImageTag = "latest"
		}
		// test for official image name (Docker Registry specific)
		if !strings.Contains(config.ImageName, "/") && config.RegistryDomain == "" {
			config.ImageName = "library/" + config.ImageName
		}

		// Get image manifest
		log.WithFields(log.Fields{
			"image": config.ImageName,
		}).Infof(getFmtStr(), "Querying Registry")
		registry := initRegistry(config.ImageName, insecure, config.Username, config.Password)

		mF, err := registry.ReposManifests(config.ImageName, config.ImageTag)
		if err != nil {
			log.Fatal(err.Error())
		}

		dir, err := ioutil.TempDir("", "graboid")
		if err != nil {
			log.Fatal(err.Error())
		}
		defer os.RemoveAll(dir) // clean up

		log.Infof(getFmtStr(), "GET CONFIG")
		cfile, err := registry.RepoGetConfig(dir, config.ImageName, mF)
		if err != nil {
			log.Fatal(err.Error())
		}

		log.Infof(getFmtStr(), "GET LAYERS")
		lfiles, err := registry.RepoGetLayers(dir, config.ImageName, mF)
		if err != nil {
			log.Fatal(err.Error())
		}

		log.Infof(getFmtStr(), "CREATE manifest.json")
		_, err = createManifest(dir, cfile, lfiles)
		if err != nil {
			log.Fatal(err.Error())
		}

		tarFile := fmt.Sprintf("%s_%s.tar.gz",
			strings.Replace(config.ImageName, "/", "_", 1),
			config.ImageTag,
		)

		if runtime.GOOS == "windows" {
			log.Infof("%s: %s", "CREATE docker image tarball", tarFile)
		} else {
			log.Infof("\033[1m%s:\033[0m \033[34m%s\033[0m", "CREATE docker image tarball", tarFile)
		}
		err = tarFiles(dir, tarFile)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Infof("\033[1mSUCCESS!\033[0m")
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	log.SetHandler(clihander.Default)
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&config.File, "config", "", "config file (default is $HOME/.graboid.yaml)")
	rootCmd.PersistentFlags().StringVar(&config.IndexUrl, "index", "", "override index endpoint")
	rootCmd.PersistentFlags().StringVar(&config.RegistryDomain, "registry", "", "override registry endpoint")
	rootCmd.PersistentFlags().StringVar(&config.Username, "username", "", "Username")
	rootCmd.PersistentFlags().StringVar(&config.Password, "password", "", "Password")
	rootCmd.PersistentFlags().BoolVarP(&config.Verbose, "verbose", "V", false, "verbose output")

	rootCmd.Flags().String("proxy", "", "HTTP/HTTPS proxy")
	rootCmd.Flags().Bool("insecure", false, "do not verify ssl certs")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if config.File != "" {
		// Use config file from the flag.
		viper.SetConfigFile(config.File)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".graboid.{yml,json,yaml,...}".
		viper.AddConfigPath(home)
		viper.SetConfigName(".graboid")
	}

	// If a config file is found, read it in.
	err := viper.ReadInConfig(); if err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}


	// Set config entries from file, or defaults
	if viper.GetString("index") == "" {
		config.IndexUrl = "https://index.docker.io"
	} else {
		config.IndexUrl = viper.GetString("index")
	}

	if viper.GetString("registry") == "" {
		config.RegistryDomain = "https://registry-1.docker.com"
	} else {
		config.RegistryDomain = viper.GetString("registry")
	}

	if viper.GetString("username") != "" {
		config.Username = viper.GetString("username")
	}

	if viper.GetString("password") != "" {
		config.Password = viper.GetString("password")
	}

}
