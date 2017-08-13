package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/apex/log"
)

// Registry registry object
type Registry struct {
	URL          *url.URL
	Host         string
	RegistryHost string
	client       *http.Client
	Token        string
}

type auth struct {
	Token       string    `json:"token,omitempty"`
	AccessToken string    `json:"access_token,omitempty"`
	ExpiresIn   int       `json:"expires_in,omitempty"`
	IssuedAt    time.Time `json:"issued_at,omitempty"`
}

// Tags is the image tags struct
type Tags struct {
	Name string   `json:"name,omitempty"`
	Tags []string `json:"tags,omitempty"`
}

// Manifests is the image manifest struct
type Manifests struct {
	Config        manifestConfig  `json:"config,omitempty"`
	Layers        []manifestLayer `json:"layers,omitempty"`
	MediaType     string          `json:"mediaType,omitempty"`
	SchemaVersion int             `json:"schemaVersion,omitempty"`
}

type manifestConfig struct {
	Digest    string `json:"digest,omitempty"`
	MediaType string `json:"mediaType,omitempty"`
	Size      int    `json:"size,omitempty"`
}

type manifestLayer struct {
	Digest    string `json:"digest,omitempty"`
	MediaType string `json:"mediaType,omitempty"`
	Size      int    `json:"size,omitempty"`
}

// NewRegistry creates a new Registry object
func NewRegistry(endpoint string, registryDomain string) (*Registry, error) {
	u, e := url.Parse(endpoint)
	if e != nil {
		return nil, e
	}
	if u.Host == "" {
		u.Host = u.Path
	}
	origURL := u
	host := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	registryHost := ""
	if registryDomain != "" {
		u, e = url.Parse(registryDomain)
		if e != nil {
			return nil, e
		}
		if u.Host == "" {
			u.Host = u.Path
		}
		registryHost = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	}
	client := &http.Client{}
	return &Registry{
		URL:          origURL,
		Host:         host,
		RegistryHost: registryHost,
		client:       client}, nil
}

// GetToken retrives a docker registry API pull token
func (reg *Registry) GetToken(username string, password string, reposName string) error {
	u := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", reposName)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}
	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}
	res, err := reg.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("HTTP Error: %s", res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var a = new(auth)
	err = json.Unmarshal(body, &a)
	if err != nil {
		return err
	}

	reg.Token = a.Token
	log.WithField("token", a.Token).Debugf("got token")

	return nil
}

func (reg *Registry) doGet(url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", reg.Token))
	// add additional headers
	if headers != nil {
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}
	res, err := reg.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		res.Body.Close()
		return nil, fmt.Errorf("HTTP Error: %s", res.Status)
	}
	return res, nil
}

// ReposTags gets a list of the docker image tags
func (reg *Registry) ReposTags(reposName string) (*Tags, error) {
	url := fmt.Sprintf("https://index.docker.io/v2/%s/tags/list", reposName)

	log.WithFields(log.Fields{
		"url": url,
	}).Debug("downloading tags")
	res, err := reg.doGet(url, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	rawJSON, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	t := new(Tags)
	if err := json.Unmarshal(rawJSON, &t); err != nil {
		return nil, err
	}

	return t, nil
}

// ReposManifests gets docker image manifest for name:tag
func (reg *Registry) ReposManifests(reposName, repoTag string) (*Manifests, error) {
	headers := make(map[string]string)
	url := fmt.Sprintf("%s/v2/%s/manifests/%s", reg.Host, reposName, repoTag)
	headers["Accept"] = "application/vnd.docker.distribution.manifest.v2+json"
	log.WithFields(log.Fields{
		"url":     url,
		"headers": headers,
		"image":   reposName,
		"tag":     repoTag,
	}).Debug("get manifests")
	res, err := reg.doGet(url, headers)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	rawJSON, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	m := new(Manifests)
	if err := json.Unmarshal(rawJSON, &m); err != nil {
		return nil, err
	}

	return m, nil
}

// RepoGetConfig gets docker image config JSON
func (reg *Registry) RepoGetConfig(tempDir, reposName string, manifest *Manifests) (string, error) {
	// Create the file
	tmpfn := filepath.Join(tempDir, fmt.Sprintf("%s.json", strings.TrimPrefix(manifest.Config.Digest, "sha256:")))
	out, err := os.Create(tmpfn)
	if err != nil {
		log.WithError(err).Error("create config file failed")
	}
	defer out.Close()
	// Download config
	headers := make(map[string]string)
	url := fmt.Sprintf("%s/v2/%s/blobs/%s", reg.Host, reposName, manifest.Config.Digest)
	headers["Accept"] = manifest.Config.MediaType
	log.WithField("url", url).Debug("downloading config")
	res, err := reg.doGet(url, headers)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, res.Body)
	if err != nil {
		log.WithError(err).Error("writing config file failed")
	}

	return tmpfn, nil
}

// RepoGetLayers gets docker image layer tarballs
func (reg *Registry) RepoGetLayers(tempDir, reposName string, manifest *Manifests) ([]string, error) {
	var layerFiles []string

	for _, layer := range manifest.Layers {
		// Create the TAR file
		tmpfn := filepath.Join(tempDir, fmt.Sprintf("%s.tar", strings.TrimPrefix(layer.Digest, "sha256:")))
		layerFiles = append(layerFiles, tmpfn)
		out, err := os.Create(tmpfn)
		if err != nil {
			log.WithError(err).Error("create tar file failed")
		}
		defer out.Close()

		// Download config
		headers := make(map[string]string)
		url := fmt.Sprintf("%s/v2/%s/blobs/%s", reg.Host, reposName, layer.Digest)
		headers["Accept"] = layer.MediaType
		log.WithField("url", url).Debug("downloading layer")
		res, err := reg.doGet(url, headers)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		// Write the body to file
		_, err = io.Copy(out, res.Body)
		if err != nil {
			log.WithError(err).Error("writing tar file failed")
		}
	}

	return layerFiles, nil
}

// func (reg *Registry) LayerJson(layerId string) (*LayerJson, error) {
// 	url := fmt.Sprintf("%s/v1/images/%s/json", reg.RegistryHost, layerId)
// 	res, e := reg.doGet(url)
// 	if e != nil {
// 		return nil, e
// 	}
// 	defer res.Body.Close()
// 	rawJSON, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	result := &LayerJson{}
// 	if err := json.Unmarshal(rawJSON, &result); err != nil {
// 		return nil, err
// 	}
// 	result.Size, _ = strconv.Atoi(res.Header.Get("X-Docker-Size"))
// 	return result, nil
// }

// func (reg *Registry) LayerAncestry(layerId string) (*[]string, error) {
// 	url := fmt.Sprintf("%s/v1/images/%s/ancestry", reg.RegistryHost, layerId)
// 	res, e := reg.doGet(url)
// 	if e != nil {
// 		return nil, e
// 	}
// 	defer res.Body.Close()
// 	rawJSON, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	result := new([]string)
// 	if err := json.Unmarshal(rawJSON, &result); err != nil {
// 		return nil, err
// 	}
// 	return result, nil
// }
