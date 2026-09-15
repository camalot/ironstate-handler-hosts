// Command stagelocal copies a locally built plugin binary into ironstate's
// per-user plugin cache (and removes it again), so a dev build can be
// exercised via `ironstate plugin test`/`bench` or a playbook without
// publishing a release - see docs/plugins.md's "Isolated testing" section
// in the ironstate repo for the store layout this mirrors.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type manifest struct {
	Organization    string    `json:"organization"`
	Name            string    `json:"name"`
	Version         string    `json:"version"`
	Source          string    `json:"source"`
	Checksum        string    `json:"checksum"`
	InstalledAt     time.Time `json:"installed_at"`
	HandlerNames    []string  `json:"handler_names"`
	ProtocolVersion int       `json:"protocol_version"`
	License         string    `json:"license,omitempty"`
	DocSummary      string    `json:"doc_summary,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "stage":
		stage(os.Args[2:])
	case "unstage":
		unstage(os.Args[2:])
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: stagelocal stage --org <organization> --name <plugin> --version <version> --binary <path> --handlers <name,...>")
	fmt.Fprintln(os.Stderr, "       stagelocal unstage --org <organization> --name <plugin> --version <version>")
	os.Exit(2)
}

func stage(args []string) {
	set := flag.NewFlagSet("stage", flag.ExitOnError)
	org := set.String("org", "", "plugin organization")
	name := set.String("name", "", "plugin name")
	version := set.String("version", "v0.0.0-dev", "dev version to stage as")
	binary := set.String("binary", "", "path to the locally built plugin binary")
	handlers := set.String("handlers", "", "comma-separated handler names the plugin serves")
	license := set.String("license", "", "optional license identifier")
	summary := set.String("summary", "", "optional one-line doc summary")
	_ = set.Parse(args)
	if *org == "" || *name == "" || *binary == "" {
		usage()
	}

	destDir := versionDir(*org, *name, *version)
	if err := os.MkdirAll(destDir, 0o755); err != nil { //nolint:gosec // matches ironstate's own plugin store directories
		fatal(err)
	}
	destBinary := filepath.Join(destDir, binaryName(*name))
	if err := copyFile(*binary, destBinary); err != nil {
		fatal(err)
	}
	checksum, err := sha256File(destBinary)
	if err != nil {
		fatal(err)
	}

	var handlerNames []string
	if *handlers != "" {
		for _, h := range strings.Split(*handlers, ",") {
			if h = strings.TrimSpace(h); h != "" {
				handlerNames = append(handlerNames, h)
			}
		}
	}
	man := manifest{
		Organization:    *org,
		Name:            *name,
		Version:         *version,
		Source:          "github.com/" + *org + "/ironstate-handler-" + *name,
		Checksum:        checksum,
		InstalledAt:     time.Now().UTC(),
		HandlerNames:    handlerNames,
		ProtocolVersion: 1,
		License:         *license,
		DocSummary:      *summary,
	}
	encoded, err := json.MarshalIndent(man, "", "  ")
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destDir, "manifest.json"), append(encoded, '\n'), 0o600); err != nil {
		fatal(err)
	}
	fmt.Printf("staged %s.%s@%s -> %s\n", *org, *name, *version, destDir)
}

func unstage(args []string) {
	set := flag.NewFlagSet("unstage", flag.ExitOnError)
	org := set.String("org", "", "plugin organization")
	name := set.String("name", "", "plugin name")
	version := set.String("version", "v0.0.0-dev", "dev version to remove")
	_ = set.Parse(args)
	if *org == "" || *name == "" {
		usage()
	}

	destDir := versionDir(*org, *name, *version)
	if err := os.RemoveAll(destDir); err != nil {
		fatal(err)
	}
	fmt.Printf("removed %s.%s@%s\n", *org, *name, *version)
}

func versionDir(org, name, version string) string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		fatal(err)
	}
	return filepath.Join(cacheDir, "ironstate", "plugins", org, name, version)
}

func binaryName(name string) string {
	result := "ironstate-handler-" + name
	if runtime.GOOS == "windows" {
		return result + ".exe"
	}
	return result
}

func copyFile(source, destination string) error {
	input, err := os.Open(source) //nolint:gosec // source is a caller-provided local build output
	if err != nil {
		return fmt.Errorf("open plugin binary: %w", err)
	}
	defer func() { _ = input.Close() }()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755) //nolint:gosec // plugin binaries must be executable
	if err != nil {
		return fmt.Errorf("create staged plugin binary: %w", err)
	}
	if _, err := io.Copy(output, input); err != nil {
		_ = output.Close()
		return fmt.Errorf("copy plugin binary: %w", err)
	}
	return output.Close()
}

func sha256File(path string) (string, error) {
	file, err := os.Open(path) //nolint:gosec // path is the just-staged plugin binary
	if err != nil {
		return "", fmt.Errorf("open staged plugin binary: %w", err)
	}
	defer func() { _ = file.Close() }()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("checksum staged plugin binary: %w", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
