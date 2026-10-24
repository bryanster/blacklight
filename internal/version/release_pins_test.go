package version_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// TestDeployPinsMatchLatestRelease keeps the deployment examples on the newest
// release in CHANGELOG.md. The Deploy to Azure button deploys
// deploy/azure-arm/azuredeploy.json straight from main, so a stale default
// there gives every new one-click install an old image. Cutting a release adds
// a `## [x.y.z]` section to the changelog, and this test then fails until every
// pin below names vx.y.z.
func TestDeployPinsMatchLatestRelease(t *testing.T) {
	root := filepath.Join("..", "..")
	read := func(rel string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		return string(b)
	}

	m := regexp.MustCompile(`(?m)^## \[(\d+\.\d+\.\d+)\]`).FindStringSubmatch(read("CHANGELOG.md"))
	if m == nil {
		t.Fatal("CHANGELOG.md has no released `## [x.y.z]` section")
	}
	want := "v" + m[1]

	var tmpl struct {
		Parameters struct {
			ImageTag struct {
				DefaultValue string `json:"defaultValue"`
			} `json:"imageTag"`
		} `json:"parameters"`
	}
	if err := json.Unmarshal([]byte(read("deploy/azure-arm/azuredeploy.json")), &tmpl); err != nil {
		t.Fatalf("parse azuredeploy.json: %v", err)
	}

	var params struct {
		Parameters struct {
			ImageTag struct {
				Value string `json:"value"`
			} `json:"imageTag"`
		} `json:"parameters"`
	}
	if err := json.Unmarshal([]byte(read("deploy/azure-arm/azuredeploy.parameters.json")), &params); err != nil {
		t.Fatalf("parse azuredeploy.parameters.json: %v", err)
	}

	tfDefault := regexp.MustCompile(`(?s)variable "image_tag" \{.*?default\s*=\s*"([^"]+)"`)
	tfExample := regexp.MustCompile(`(?m)^image_tag\s*=\s*"([^"]+)"`)
	readmeStatus := regexp.MustCompile("Blacklight `(v[^`]+)` is released")

	first := func(re *regexp.Regexp, s string) string {
		if m := re.FindStringSubmatch(s); m != nil {
			return m[1]
		}
		return "(not found)"
	}

	pins := []struct{ where, got string }{
		{"deploy/azure-arm/azuredeploy.json imageTag default (the Deploy to Azure button)", tmpl.Parameters.ImageTag.DefaultValue},
		{"deploy/azure-arm/azuredeploy.parameters.json imageTag", params.Parameters.ImageTag.Value},
		{"deploy/azure-terraform/variables.tf image_tag default", first(tfDefault, read("deploy/azure-terraform/variables.tf"))},
		{"deploy/azure-terraform/terraform.tfvars.example image_tag", first(tfExample, read("deploy/azure-terraform/terraform.tfvars.example"))},
		{"README.md status line", first(readmeStatus, read("README.md"))},
	}
	for _, p := range pins {
		if p.got != want {
			t.Errorf("%s = %q, want %q (the newest release in CHANGELOG.md)", p.where, p.got, want)
		}
	}
}
