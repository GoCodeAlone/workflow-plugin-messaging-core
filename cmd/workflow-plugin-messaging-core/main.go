// Command workflow-plugin-messaging-core hosts the messaging-core manifest over
// the workflow external-plugin protocol.
package main

import (
	"os"

	pluginpkg "github.com/GoCodeAlone/workflow/plugin"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

var Version = "dev"

type manifestProvider struct {
	manifest *pluginpkg.PluginManifest
}

func (p manifestProvider) Manifest() sdk.PluginManifest {
	return sdk.PluginManifest{
		Name:        p.manifest.Name,
		Version:     p.manifest.Version,
		Author:      p.manifest.Author,
		Description: p.manifest.Description,
	}
}

func main() {
	manifestJSON, err := os.ReadFile("plugin.json")
	if err != nil {
		panic(err)
	}
	manifest := sdk.MustEmbedManifest(manifestJSON)
	sdk.Serve(
		manifestProvider{manifest: manifest},
		sdk.WithManifestProvider(manifest),
		sdk.WithBuildVersion(sdk.ResolveBuildVersion(Version)),
	)
}
