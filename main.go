package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/build"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/google/licenseclassifier"
	"github.com/google/licenseclassifier/licenses"
	"github.com/rogpeppe/go-internal/module"
	"github.com/sirupsen/logrus"
	"golang.org/x/mod/modfile"
)

func init() {
	http.DefaultClient.Timeout = 1 * time.Minute
}

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, "usage: %s go.mod [...]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	log.SetFlags(0) // no timestamp

	// parse configuration
	flag.Usage = usage
	configFile := flag.String("config", "", "Configuration file (optional)")
	flag.Parse()
	config := ReadConfig(*configFile)
	if flag.NArg() > 0 {
		config.GoModFiles = flag.Args()
	}
	if len(config.GoModFiles) == 0 {
		config.GoModFiles = []string{"go.mod"}
	}

	// read go.mods
	var modInfos []module.Version
	for _, arg := range config.GoModFiles {
		infos, err := loadGoMod(arg)
		if err != nil {
			logrus.Fatalf("error processing module %s: %v", arg, err)
		}
		modInfos = append(modInfos, infos...)
	}

	// prepare classifier
	classifier, err := licenseclassifier.New(licenseclassifier.DefaultConfidenceThreshold)
	if err != nil {
		log.Fatalf("cannot create license classifier: %v", err)
	}

	// read license files and classify
	for _, info := range modInfos {
		var license string
		var err error
		license, err = GetLicenseFromOverrides(config.LicenseOverrides, info)
		if err != nil {
			license, err = GetLicenseFromDisk(info)
			if err != nil {
				license, err = GetLicenseOnline(info)
				if err != nil {
					logrus.Fatalf("error reading license file for module %s: %v", info.Path, err)
				}
			}
		}
		fmt.Printf("License for %s:\n\n", info.Path)
		fmt.Printf("%s\n\n\n", license)

		// classify license
		m := classifier.NearestMatch(licenseclassifier.TrimExtraneousTrailingText(license))
		if !slices.Contains(config.AllowedLicenses, m.Name) {
			fmt.Printf("Module: %s %s, has forbidden license type: %s\n", info.Path, info.Version, m.Name)
			os.Exit(2)
		}
	}
}

func loadGoMod(filename string) ([]module.Version, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	f, err := modfile.ParseLax(filename, data, nil)
	if err != nil {
		return nil, err
	}

	var mvs []module.Version
	for _, req := range f.Require {
		mod := req.Mod
		mvs = append(mvs, mod)
	}
	for _, rep := range f.Replace {
		mvs = append(mvs, rep.New)
	}
	return mvs, nil
}

func GetLicenseFromOverrides(licenseOverrides []LicenseOverride, mod module.Version) (string, error) {
	for _, o := range licenseOverrides {
		if o.ModuleVersion == mod {
			return GetLicenseBySPDX(o.SPDX)
		}
	}
	return "", errors.New("not found")
}

func GetLicenseOnline(mod module.Version) (string, error) {
	target := fmt.Sprintf("https://proxy.golang.org/%s/@v/%s.zip", strings.ToLower(mod.Path), mod.Version)
	resp, err := http.Get(target)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GET %v: %v", target, resp.Status)
	}
	zippedModule, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading module %s (read all): %w", mod.Path, err)
	}
	zipReader, err := zip.NewReader(bytes.NewReader(zippedModule), int64(len(zippedModule)))
	if err != nil {
		return "", fmt.Errorf("error reading module %s zip: %w", mod.Path, err)
	}
	var filenames []string
	for _, zipFile := range zipReader.File {
		if !IsLicenseFilename(mod, zipFile.Name) {
			filenames = append(filenames, zipFile.Name)
			continue
		}
		licenseFile, err := zipFile.Open()
		if err != nil {
			return "", fmt.Errorf("error opening module %s license file: %w", mod.Path, err)
		}
		defer licenseFile.Close()
		licenseBytes, err := io.ReadAll(licenseFile)
		if err != nil {
			return "", fmt.Errorf("error reading module %s license file: %w", mod.Path, err)
		}
		return string(licenseBytes), nil
	}
	return "", fmt.Errorf("license file not found in module %s. Files found: %s", mod.Path, filenames)
}

func GetLicenseBySPDX(spdx string) (string, error) {
	license, err := licenses.ReadLicenseFile(spdx + ".txt")
	if err != nil {
		return "", err
	}
	return string(license), nil
}

func GetLicenseFromDisk(mod module.Version) (string, error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	p := path.Join(gopath, "pkg", "mod", fmt.Sprintf("%s@%s", strings.ToLower(mod.Path), mod.Version))
	entries, err := os.ReadDir(p)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		fullpath := p + "/" + e.Name()
		if IsLicenseFilename(mod, fullpath) {
			bytes, err := os.ReadFile(fullpath)
			if err != nil {
				return "", err
			}
			return string(bytes), nil
		}
	}
	return "", errors.New("not found")
}

func IsLicenseFilename(mod module.Version, filename string) bool {
	return strings.ToLower(fmt.Sprintf("%s@%s/%s", mod.Path, mod.Version, "license")) == strings.ToLower(filename) ||
		strings.ToLower(fmt.Sprintf("%s@%s/%s", mod.Path, mod.Version, "license.txt")) == strings.ToLower(filename) ||
		strings.ToLower(fmt.Sprintf("%s@%s/%s", mod.Path, mod.Version, "license.md")) == strings.ToLower(filename)
}
