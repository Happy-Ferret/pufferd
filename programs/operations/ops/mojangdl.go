package ops

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pufferpanel/apufferi/common"
	"github.com/pufferpanel/apufferi/logging"
	"github.com/pufferpanel/pufferd/environments"
	"net/http"
)

const VERSION_JSON = "https://launchermeta.mojang.com/mc/game/version_manifest.json"

type MojangDl struct {
	Version string
	Target  string
}

func (op MojangDl) Run(env environments.Environment) error {
	client := &http.Client{}

	response, err := client.Get(VERSION_JSON)
	if err != nil {
		return err
	}

	var data MojangLauncherJson
	json.NewDecoder(response.Body).Decode(&data)
	response.Body.Close()

	var targetVersion string
	switch op.Version {
		case "release":
			targetVersion = data.Latest.Release
		case "latest":
			targetVersion = data.Latest.Release
		case "snapshot":
			targetVersion = data.Latest.Snapshot
		default:
			targetVersion = op.Version
	}

	for _, version := range data.Versions {
		if version.Id == targetVersion {
			logging.Debugf("Version %s json located, downloading from %s", version.Id, version.Url)
			env.DisplayToConsole(fmt.Sprintf("Version %s json located, downloading from %s\n", version.Id, version.Url))
			//now, get the version json for this one...
			return downloadServerFromJson(version.Url, op.Target, env)
		}
	}

	env.DisplayToConsole("Could not locate version " + targetVersion + "\n")

	return errors.New("Version not located: " + op.Version)
}

func downloadServerFromJson(url, target string, env environments.Environment) error {
	client := &http.Client{}
	response, err := client.Get(url)
	if err != nil {
		return err
	}

	var data MojangVersionJson
	json.NewDecoder(response.Body).Decode(&data)
	response.Body.Close()

	serverBlock := data.Downloads["server"]

	logging.Debugf("Version jar located, downloading from %s", serverBlock.Url)
	env.DisplayToConsole(fmt.Sprintf("Version jar located, downloading from %s\n", serverBlock.Url))

	return downloadFile(serverBlock.Url, target, env)
}

type MojangDlOperationFactory struct {
}

func (of MojangDlOperationFactory) Create(op CreateOperation) Operation {
	version := op.OperationArgs["version"].(string)
	target := op.OperationArgs["target"].(string)

	version = common.ReplaceTokens(version, op.DataMap)
	target = common.ReplaceTokens(target, op.DataMap)

	return MojangDl{Version: version, Target: target}
}

func (of MojangDlOperationFactory) Key() string {
	return "mojangdl"
}

type MojangLauncherJson struct {
	Versions []MojangLauncherVersion `json:"versions"`
	Latest MojangLatest `json:"latest"`
}

type MojangLatest struct {
	Release string `json:"release"`
	Snapshot string `json:"snapshot"`
}

type MojangLauncherVersion struct {
	Id   string `json:"id"`
	Url  string `json:"url"`
	Type string `json:"type"`
}

type MojangVersionJson struct {
	Downloads map[string]MojangDownloadType `json:"downloads"`
}

type MojangDownloadType struct {
	Sha1 string `json:"sha1"`
	Size uint64 `json:"size"`
	Url  string `json:"url"`
}