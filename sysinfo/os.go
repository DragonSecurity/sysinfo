package sysinfo

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

type OS struct {
	Name         string `json:"name,omitempty"`
	Vendor       string `json:"vendor,omitempty"`
	Version      string `json:"version,omitempty"`
	Release      string `json:"release,omitempty"`
	Architecture string `json:"architecture,omitempty"`
}

var (
	rePrettyName = regexp.MustCompile(`^PRETTY_NAME=(.*)$`)
	reID         = regexp.MustCompile(`^ID=(.*)$`)
	reVersionID  = regexp.MustCompile(`^VERSION_ID=(.*)$`)
	reUbuntu     = regexp.MustCompile(`[( ]([\d.]+)`)
	reAlma       = regexp.MustCompile(`^AlmaLinux release ([\d\.]+)`)
	reCentOS     = regexp.MustCompile(`^CentOS( Linux)? release ([\d\.]+)`)
	reRocky      = regexp.MustCompile(`^Rocky Linux release ([\d\.]+)`)
	reRedHat     = regexp.MustCompile(`[( ]([\d.]+)`)
)

func (si *SysInfo) getOSInfo() {
	// This seems to be the best and most portable way to detect OS architecture (NOT kernel!)
	if _, err := os.Stat("/lib64/ld-linux-x86-64.so.2"); err == nil {
		si.OS.Architecture = "amd64"
	} else if _, err := os.Stat("/lib/ld-linux.so.2"); err == nil {
		si.OS.Architecture = "i386"
	}

	f, err := os.Open("/etc/os-release")
	if err != nil {
		return
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if m := rePrettyName.FindStringSubmatch(s.Text()); m != nil {
			si.OS.Name = strings.Trim(m[1], `"`)
		} else if m := reID.FindStringSubmatch(s.Text()); m != nil {
			si.OS.Vendor = strings.Trim(m[1], `"`)
		} else if m := reVersionID.FindStringSubmatch(s.Text()); m != nil {
			si.OS.Version = strings.Trim(m[1], `"`)
		}
	}

	switch si.OS.Vendor {
	case "debian":
		si.OS.Release = slurpFile("/etc/debian_version")
	case "ubuntu":
		if m := reUbuntu.FindStringSubmatch(si.OS.Name); m != nil {
			si.OS.Release = m[1]
		}
	case "almalinux":
		if release := slurpFile("/etc/almalinux-release"); release != "" {
			if m := reAlma.FindStringSubmatch(release); m != nil {
				si.OS.Release = m[1]
			}
		}

		si.OS.Version = strings.Split(si.OS.Release, ".")[0]
	case "centos":
		if release := slurpFile("/etc/centos-release"); release != "" {
			if m := reCentOS.FindStringSubmatch(release); m != nil {
				si.OS.Release = m[2]
			}
		}
	case "rocky":
		if release := slurpFile("/etc/rocky-release"); release != "" {
			if m := reRocky.FindStringSubmatch(release); m != nil {
				si.OS.Release = m[1]
			}
		}

		si.OS.Version = strings.Split(si.OS.Release, ".")[0]

	case "rhel":
		if release := slurpFile("/etc/redhat-release"); release != "" {
			if m := reRedHat.FindStringSubmatch(release); m != nil {
				si.OS.Release = m[1]
			}
		}
		if si.OS.Release == "" {
			if m := reRedHat.FindStringSubmatch(si.OS.Name); m != nil {
				si.OS.Release = m[1]
			}
		}
	}
}

func GetOSInfo() {
	var si SysInfo
	si.GetSysInfo()
	data, err := json.MarshalIndent(&si.OS, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))
}