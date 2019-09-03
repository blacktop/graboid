package registry

import (
	"crypto/tls"
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

	pb "gopkg.in/cheggaaa/pb.v1"

	"github.com/apex/log"
)

// Config registry config struct
type Config struct {
	Endpoint       string
	RegistryDomain string
	Proxy          string
	Insecure       bool
	Username       string
	Password       string
	RepoName       string
}

// Registry registry object
type Registry struct {
	URL          *url.URL
	Host         string
	RegistryHost string
	client       *http.Client
	Auth         auth
	Config       Config
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

func getProxy(proxy string) func(*http.Request) (*url.URL, error) {
	if len(proxy) > 0 {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			log.WithError(err).Error("bad proxy url")
		}
		return http.ProxyURL(proxyURL)
	}
	return http.ProxyFromEnvironment
}

// TokenExpired returns wheither or not an auth token has expired
func (reg *Registry) TokenExpired() bool {
	duration := time.Since(reg.Auth.IssuedAt)
	if int(duration.Seconds()) > reg.Auth.ExpiresIn {
		log.Warn("auth token expired")
		return true
	}
	return false
}

// New creates a new Registry object
func New(rc Config) (*Registry, error) {
	u, e := url.Parse(rc.Endpoint)
	if e != nil {
		return nil, e
	}
	if u.Host == "" {
		u.Host = u.Path
	}
	origURL := u
	host := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	registryHost := ""
	if rc.RegistryDomain != "" {
		u, e = url.Parse(rc.RegistryDomain)
		if e != nil {
			return nil, e
		}
		if u.Host == "" {
			u.Host = u.Path
		}
		registryHost = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy:           getProxy(rc.Proxy),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: rc.Insecure},
		},
	}
	return &Registry{
		URL:          origURL,
		Host:         host,
		RegistryHost: registryHost,
		client:       client,
		Config:       rc}, nil
}

// GetToken retrives a docker registry API pull token
func (reg *Registry) GetToken() error {
	u := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", reg.Config.RepoName)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}
	if reg.Config.Username != "" && reg.Config.Password != "" {
		req.SetBasicAuth(reg.Config.Username, reg.Config.Password)
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

	reg.Auth = *a
	log.WithField("token", a.Token).Debugf("got token")

	return nil
}

func (reg *Registry) doGet(url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", reg.Auth.Token))
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
	url := fmt.Sprintf("https://registry-1.docker.io/v2/%s/tags/list", reposName)

	if reg.TokenExpired() {
		reg.GetToken()
	}

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

	if reg.TokenExpired() {
		reg.GetToken()
	}

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

	if reg.TokenExpired() {
		reg.GetToken()
	}

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

	return filepath.Base(tmpfn), nil
}

// RepoGetLayers gets docker image layer tarballs
func (reg *Registry) RepoGetLayers(tempDir, reposName string, manifest *Manifests) ([]string, error) {
	var layerFiles []string

	for _, layer := range manifest.Layers {
		// Create the TAR file
		tmpfn := filepath.Join(tempDir, fmt.Sprintf("%s.tar", strings.TrimPrefix(layer.Digest, "sha256:")))
		layerFiles = append(layerFiles, filepath.Base(tmpfn))
		out, err := os.Create(tmpfn)
		if err != nil {
			log.WithError(err).Error("create tar file failed")
		}
		defer out.Close()

		// Download layer
		headers := make(map[string]string)
		url := fmt.Sprintf("%s/v2/%s/blobs/%s", reg.Host, reposName, layer.Digest)
		headers["Accept"] = layer.MediaType
		log.WithField("url", url).Debug("downloading layer")

		if reg.TokenExpired() {
			reg.GetToken()
		}

		res, err := reg.doGet(url, headers)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		// create progressbar
		bar := pb.New(layer.Size).SetUnits(pb.U_BYTES)
		bar.SetWidth(90)
		bar.Start()
		reader := bar.NewProxyReader(res.Body)
		// Write the body to file
		_, err = io.Copy(out, reader)
		if err != nil {
			log.WithError(err).Error("writing tar file failed")
		}
		bar.Finish()
	}

	return layerFiles, nil
}
