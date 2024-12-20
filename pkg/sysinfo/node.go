package sysinfo

import (
	"bufio"
	"os"
	"strings"
)

// Node information.
type Node struct {
	Hostname   string `json:"hostname,omitempty"`
	MachineID  string `json:"machineid,omitempty"`
	Hypervisor string `json:"hypervisor,omitempty"`
	Timezone   string `json:"timezone,omitempty"`
}

func (si *SysInfo) getHostname() {
	si.Node.Hostname = slurpFile("/proc/sys/kernel/hostname")
}

func (si *SysInfo) getSetMachineID() {
	const pathSystemdMachineID = "/etc/machine-id"
	const pathDbusMachineID = "/var/lib/dbus/machine-id"

	systemdMachineID := slurpFile(pathSystemdMachineID)
	dbusMachineID := slurpFile(pathDbusMachineID)

	if systemdMachineID != "" && dbusMachineID != "" {
		// All OK, just return the machine id.
		if systemdMachineID == dbusMachineID {
			si.Node.MachineID = systemdMachineID
			return
		}

		// They both exist, but they don't match! Copy systemd machine id to DBUS machine id.
		spewFile(pathDbusMachineID, systemdMachineID, 0444)
		si.Node.MachineID = systemdMachineID
		return
	}

	// Copy DBUS machine id to non-existent systemd machine id.
	if systemdMachineID == "" && dbusMachineID != "" {
		spewFile(pathSystemdMachineID, dbusMachineID, 0444)
		si.Node.MachineID = dbusMachineID
		return
	}

	// Copy systemd machine id to non-existent DBUS machine id.
	if systemdMachineID != "" && dbusMachineID == "" {
		spewFile(pathDbusMachineID, systemdMachineID, 0444)
		si.Node.MachineID = systemdMachineID
		return
	}
}

func (si *SysInfo) getTimezone() {
	const zoneInfoPrefix = "/usr/share/zoneinfo/"

	if fi, err := os.Lstat("/etc/localtime"); err == nil {
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			if tzfile, err := os.Readlink("/etc/localtime"); err == nil {
				tzfile = strings.TrimPrefix(tzfile, "..")
				if strings.HasPrefix(tzfile, zoneInfoPrefix) {
					si.Node.Timezone = strings.TrimPrefix(tzfile, zoneInfoPrefix)
					return
				}
			}
		}
	}

	if timezone := slurpFile("/etc/timezone"); timezone != "" {
		si.Node.Timezone = timezone
		return
	}

	if f, err := os.Open("/etc/sysconfig/clock"); err == nil {
		defer f.Close()
		s := bufio.NewScanner(f)
		for s.Scan() {
			if sl := strings.Split(s.Text(), "="); len(sl) == 2 {
				if sl[0] == "ZONE" {
					si.Node.Timezone = strings.Trim(sl[1], `"`)
					return
				}
			}
		}
	}
}

func (si *SysInfo) getNodeInfo() {
	si.getHostname()
	si.getSetMachineID()
	si.getHypervisor()
	si.getTimezone()
}
