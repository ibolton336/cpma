package ocp

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/fusor/cpma/env"
	"github.com/fusor/cpma/internal/io"
	"github.com/fusor/cpma/pkg/ocp3"
	"github.com/fusor/cpma/pkg/ocp4"
	"github.com/sirupsen/logrus"
)

type ConfigFile struct {
	Hostname string
	Path     string
	Content  []byte
}

type ConfigMaster struct {
	ConfigFile
	OCP3 ocp3.Master
	OCP4 ocp4.Master
}

type ConfigNode struct {
	ConfigFile
	OCP3 ocp3.Node
	OCP4 ocp4.Node
}

type Translator interface {
	Add(string)
	Decode()
	Fetch(string)
	GenYAML() ocp4.Manifests
	Translate()
}

// GetFile allows to mock file retrieval
var GetFile = io.GetFile

func (config *ConfigMaster) Add(hostname string) {
	masterf := env.Config().GetString("MasterConfigFile")

	if masterf == "" {
		masterf = "/etc/origin/master/master-config.yaml"
	}
	config.ConfigFile.Hostname = hostname
	config.ConfigFile.Path = masterf
}

func (config *ConfigNode) Add(hostname string) {
	nodef := env.Config().GetString("NodeConfigFile")

	if nodef == "" {
		nodef = "/etc/origin/node/node-config.yaml"
	}

	config.ConfigFile.Hostname = hostname
	config.ConfigFile.Path = nodef
}

func (config *ConfigMaster) Decode() {
	config.OCP3.Decode(config.ConfigFile.Content)
}

func (config *ConfigNode) Decode() {
	config.OCP3.Decode(config.ConfigFile.Content)
}

// DumpManifests creates OCDs files
func DumpManifests(outputDir string, manifests ocp4.Manifests) {
	for _, manifest := range manifests {
		maniftestfile := filepath.Join(outputDir, "manifests", manifest.Name)
		os.MkdirAll(path.Dir(maniftestfile), 0755)
		err := ioutil.WriteFile(maniftestfile, manifest.CRD, 0644)
		logrus.Printf("CRD:Added: %s", maniftestfile)
		if err != nil {
			logrus.Panic(err)
		}
	}
}

func (config *ConfigMaster) Fetch(outputDir string) {
	localF := filepath.Join(outputDir, config.Hostname, config.Path)
	config.ConfigFile.Content = GetFile(config.Hostname, config.Path, localF)
	logrus.Printf("File:Loaded: %s", localF)
}

func (config *ConfigNode) Fetch(outputDir string) {
	localF := filepath.Join(outputDir, config.Hostname, config.Path)
	config.ConfigFile.Content = GetFile(config.Hostname, config.Path, localF)
	logrus.Printf("File:Loaded: %s", localF)
}

// GenYAML returns the list of translated CRDs
func (config *ConfigMaster) GenYAML() ocp4.Manifests {
	var manifests ocp4.Manifests

	masterManifests := config.OCP4.GenYAML()

	for _, manifest := range masterManifests {
		manifests = append(manifests, manifest)
	}
	return manifests
}

// GenYAML returns the list of translated CRDs
func (config *ConfigNode) GenYAML() ocp4.Manifests {
	var manifests ocp4.Manifests

	nodeManifests := config.OCP4.GenYAML()

	for _, manifest := range nodeManifests {
		manifests = append(manifests, manifest)
	}
	return manifests
}

// Translate OCP3 to OCP4
func (config *ConfigMaster) Translate() {
	config.OCP4.Translate(config.OCP3.Config)
}

// Translate OCP3 to OCP4
func (config *ConfigNode) Translate() {
	config.OCP4.Translate(config.OCP3.Config)
}
