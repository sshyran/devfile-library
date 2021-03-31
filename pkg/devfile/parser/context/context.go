package parser

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"context"

	"github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/devfile/library/pkg/util"
	"k8s.io/klog"
)

// DevfileCtx stores context info regarding devfile
type DevfileCtx struct {

	// devfile ApiVersion
	apiVersion string

	// absolute path of devfile
	absPath string

	// relative path of devfile
	relPath string

	// raw content of the devfile
	rawContent []byte

	// devfile json schema
	jsonSchema string

	//url path of the devfile
	url string

	// filesystem for devfile
	fs filesystem.Filesystem

	// trace of all url referenced
	uriMap map[string]bool

	// registry URLs list
	registryURLs []string

	// default namespace to resolve kubernetes import reference
	defaultNameSpace string

	// kubeContext is the context used for making Kubernetes requests
	kubeContext context.Context

	// k8sClient is the Kubernetes client instance used for interacting with a cluster
	k8sClient client.Client
}

// NewDevfileCtx returns a new DevfileCtx type object
func NewDevfileCtx(path string) DevfileCtx {
	return DevfileCtx{
		relPath: path,
		fs:      filesystem.DefaultFs{},
	}
}

// NewURLDevfileCtx returns a new DevfileCtx type object
func NewURLDevfileCtx(url string) DevfileCtx {
	return DevfileCtx{
		url: url,
	}
}

// populateDevfile checks the API version is supported and returns the JSON schema for the given devfile API Version
func (d *DevfileCtx) populateDevfile() (err error) {

	// Get devfile APIVersion
	if err := d.SetDevfileAPIVersion(); err != nil {
		return err
	}

	// Read and save devfile JSON schema for provided apiVersion
	return d.SetDevfileJSONSchema()
}

// Populate fills the DevfileCtx struct with relevant context info
func (d *DevfileCtx) Populate() (err error) {
	if !strings.HasSuffix(d.relPath, ".yaml") {
		if _, err := os.Stat(filepath.Join(d.relPath, "devfile.yaml")); os.IsNotExist(err) {
			if _, err := os.Stat(filepath.Join(d.relPath, ".devfile.yaml")); os.IsNotExist(err) {
				return fmt.Errorf("the provided path is not a valid yaml filepath, and devfile.yaml or .devfile.yaml not found in the provided path : %s", d.relPath)
			} else {
				d.relPath = filepath.Join(d.relPath, ".devfile.yaml")
			}
		} else {
			d.relPath = filepath.Join(d.relPath, "devfile.yaml")
		}
	}
	if err := d.SetAbsPath(); err != nil {
		return err
	}
	klog.V(4).Infof("absolute devfile path: '%s'", d.absPath)
	if d.uriMap == nil {
		d.uriMap = make(map[string]bool)
	}
	if d.uriMap[d.absPath] {
		return fmt.Errorf("URI %v is recursively referenced", d.absPath)
	}
	d.uriMap[d.absPath] = true
	// Read and save devfile content
	if err := d.SetDevfileContent(); err != nil {
		return err
	}
	return d.populateDevfile()
}

// PopulateFromURL fills the DevfileCtx struct with relevant context info
func (d *DevfileCtx) PopulateFromURL() (err error) {
	u, err := url.ParseRequestURI(d.url)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(d.url, ".yaml") {
		u.Path = path.Join(u.Path, "devfile.yaml")
		if _, err = util.DownloadFileInMemory(u.String()); err != nil {
			u.Path = path.Join(path.Dir(u.Path), ".devfile.yaml")
			if _, err = util.DownloadFileInMemory(u.String()); err != nil {
				return fmt.Errorf("the provided url is not a valid yaml filepath, and devfile.yaml or .devfile.yaml not found in the provided path : %s", d.url)
			}
		}
		d.url = u.String()
	}
	if d.uriMap == nil {
		d.uriMap = make(map[string]bool)
	}
	if d.uriMap[d.url] {
		return fmt.Errorf("URI %v is recursively referenced", d.url)
	}
	d.uriMap[d.url] = true
	// Read and save devfile content
	if err := d.SetDevfileContent(); err != nil {
		return err
	}
	return d.populateDevfile()
}

// PopulateFromRaw fills the DevfileCtx struct with relevant context info
func (d *DevfileCtx) PopulateFromRaw() (err error) {
	return d.populateDevfile()
}

// Validate func validates devfile JSON schema for the given apiVersion
func (d *DevfileCtx) Validate() error {

	// Validate devfile
	return d.ValidateDevfileSchema()
}

// GetAbsPath func returns current devfile absolute path
func (d *DevfileCtx) GetAbsPath() string {
	return d.absPath
}

// GetURL func returns current devfile absolute URL address
func (d *DevfileCtx) GetURL() string {
	return d.url
}

// SetAbsPath sets absolute file path for devfile
func (d *DevfileCtx) SetAbsPath() (err error) {
	// Set devfile absolute path
	if d.absPath, err = util.GetAbsPath(d.relPath); err != nil {
		return err
	}
	klog.V(2).Infof("absolute devfile path: '%s'", d.absPath)

	return nil

}

// GetURIMap func returns current devfile uri map
func (d *DevfileCtx) GetURIMap() map[string]bool {
	return d.uriMap
}

// SetURIMap set uri map in the devfile ctx
func (d *DevfileCtx) SetURIMap(uriMap map[string]bool) {
	d.uriMap = uriMap
}

// GetRegistryURLs func returns current devfile registry URLs
func (d *DevfileCtx) GetRegistryURLs() []string {
	return d.registryURLs
}

// SetRegistryURLs set registry URLs in the devfile ctx
func (d *DevfileCtx) SetRegistryURLs(registryURLs []string) {
	d.registryURLs = registryURLs
}

// GetDefaultNameSpace func returns current devfile default namespace
func (d *DevfileCtx) GetDefaultNameSpace() string {
	return d.defaultNameSpace
}

// SetDefaultNameSpace set default namespace in the devfile ctx
func (d *DevfileCtx) SetDefaultNameSpace(defaultNameSpace string) {
	d.defaultNameSpace = defaultNameSpace
}

// SetKubeContext set context in the devfile ctx
func (d *DevfileCtx) SetKubeContext(kubeContext context.Context) {
	d.kubeContext = kubeContext
}

// GetKubeContext func returns current devfile context to make Kubernetes request
func (d *DevfileCtx) GetKubeContext() context.Context {
	return d.kubeContext
}

// SetKubeContext set Kubernetes client instance in the devfile ctx
func (d *DevfileCtx) SetK8sClient(k8sClient client.Client) {
	d.k8sClient = k8sClient
}

// GetK8sClient func returns current devfile Kubernetes client instance to interact with a cluster
func (d *DevfileCtx) GetK8sClient() client.Client {
	return d.k8sClient
}