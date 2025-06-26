//go:build libvirt
// +build libvirt

package libvirt

import (
	"encoding/xml"
	"fmt"
	"math/rand"
	"strings"
	"text/template"
	"time"

	"github.com/sirupsen/logrus"
	"libvirt.org/go/libvirt"
)

// DomainManager handles VM (domain) operations
type DomainManager struct {
	client *Client
	logger *logrus.Entry
}

// NewDomainManager creates a new domain manager
func NewDomainManager(client *Client, logger *logrus.Entry) *DomainManager {
	return &DomainManager{
		client: client,
		logger: logger.WithField("component", "domain-manager"),
	}
}

// DomainXML represents the libvirt domain XML structure
type DomainXML struct {
	XMLName    xml.Name    `xml:"domain"`
	Type       string      `xml:"type,attr"`
	Name       string      `xml:"name"`
	UUID       string      `xml:"uuid,omitempty"`
	Memory     MemoryXML   `xml:"memory"`
	CurrentMem MemoryXML   `xml:"currentMemory"`
	VCPU       VCPUXML     `xml:"vcpu"`
	OS         OSXML       `xml:"os"`
	Features   FeaturesXML `xml:"features"`
	CPU        CPUXML      `xml:"cpu"`
	Clock      ClockXML    `xml:"clock"`
	OnPoweroff string      `xml:"on_poweroff"`
	OnReboot   string      `xml:"on_reboot"`
	OnCrash    string      `xml:"on_crash"`
	Devices    DevicesXML  `xml:"devices"`
}

type MemoryXML struct {
	Unit  string `xml:"unit,attr"`
	Value string `xml:",chardata"`
}

type VCPUXML struct {
	Placement string `xml:"placement,attr"`
	Value     string `xml:",chardata"`
}

type OSXML struct {
	Type     OSTypeXML   `xml:"type"`
	Boot     []BootXML   `xml:"boot"`
	BootMenu BootMenuXML `xml:"bootmenu"`
}

type OSTypeXML struct {
	Arch    string `xml:"arch,attr"`
	Machine string `xml:"machine,attr"`
	Value   string `xml:",chardata"`
}

type BootXML struct {
	Dev string `xml:"dev,attr"`
}

type BootMenuXML struct {
	Enable string `xml:"enable,attr"`
}

type FeaturesXML struct {
	ACPI struct{} `xml:"acpi"`
	APIC struct{} `xml:"apic"`
}

type CPUXML struct {
	Mode  string `xml:"mode,attr"`
	Check string `xml:"check,attr"`
}

type ClockXML struct {
	Offset string `xml:"offset,attr"`
}

type DevicesXML struct {
	Emulator   string         `xml:"emulator"`
	Disks      []DiskXML      `xml:"disk"`
	Interfaces []InterfaceXML `xml:"interface"`
	Console    ConsoleXML     `xml:"console"`
	Graphics   GraphicsXML    `xml:"graphics"`
	Video      VideoXML       `xml:"video"`
	MemBalloon MemBalloonXML  `xml:"memballoon"`
}

type DiskXML struct {
	Type   string       `xml:"type,attr"`
	Device string       `xml:"device,attr"`
	Driver DriverXML    `xml:"driver"`
	Source SourceXML    `xml:"source"`
	Target TargetXML    `xml:"target"`
	Boot   BootOrderXML `xml:"boot"`
}

type DriverXML struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
}

type SourceXML struct {
	File   string `xml:"file,attr,omitempty"`
	Pool   string `xml:"pool,attr,omitempty"`
	Volume string `xml:"volume,attr,omitempty"`
}

type TargetXML struct {
	Dev string `xml:"dev,attr"`
	Bus string `xml:"bus,attr"`
}

type BootOrderXML struct {
	Order string `xml:"order,attr"`
}

type InterfaceXML struct {
	Type   string      `xml:"type,attr"`
	Source IfSourceXML `xml:"source"`
	Model  ModelXML    `xml:"model"`
	MAC    MACXML      `xml:"mac"`
}

type IfSourceXML struct {
	Network string `xml:"network,attr,omitempty"`
	Bridge  string `xml:"bridge,attr,omitempty"`
}

type ModelXML struct {
	Type string `xml:"type,attr"`
}

type MACXML struct {
	Address string `xml:"address,attr"`
}

type ConsoleXML struct {
	Type   string           `xml:"type,attr"`
	Target ConsoleTargetXML `xml:"target"`
}

type ConsoleTargetXML struct {
	Type string `xml:"type,attr"`
	Port string `xml:"port,attr"`
}

type GraphicsXML struct {
	Type     string `xml:"type,attr"`
	Autoport string `xml:"autoport,attr"`
	Listen   string `xml:"listen,attr"`
	KeyMap   string `xml:"keymap,attr,omitempty"`
}

type VideoXML struct {
	Model VideoModelXML `xml:"model"`
}

type VideoModelXML struct {
	Type string `xml:"type,attr"`
}

type MemBalloonXML struct {
	Model string `xml:"model,attr"`
}

// CreateDomain creates a new VM domain from XML
func (dm *DomainManager) CreateDomain(domainXML string) (*libvirt.Domain, error) {
	conn, err := dm.client.GetConnection()
	if err != nil {
		return nil, err
	}

	domain, err := conn.DomainDefineXML(domainXML)
	if err != nil {
		return nil, fmt.Errorf("failed to define domain: %w", err)
	}

	dm.logger.Info("Domain created successfully")
	return domain, nil
}

// GetDomain retrieves a domain by name
func (dm *DomainManager) GetDomain(name string) (*libvirt.Domain, error) {
	conn, err := dm.client.GetConnection()
	if err != nil {
		return nil, err
	}

	domain, err := conn.LookupDomainByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup domain %s: %w", name, err)
	}

	return domain, nil
}

// GetDomainByUUID retrieves a domain by UUID
func (dm *DomainManager) GetDomainByUUID(uuid string) (*libvirt.Domain, error) {
	conn, err := dm.client.GetConnection()
	if err != nil {
		return nil, err
	}

	domain, err := conn.LookupDomainByUUIDString(uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup domain by UUID %s: %w", uuid, err)
	}

	return domain, nil
}

// StartDomain starts a VM
func (dm *DomainManager) StartDomain(domain *libvirt.Domain) error {
	if err := domain.Create(); err != nil {
		return fmt.Errorf("failed to start domain: %w", err)
	}

	name, _ := domain.GetName()
	dm.logger.WithField("domain", name).Info("Domain started")
	return nil
}

// StopDomain stops a VM gracefully
func (dm *DomainManager) StopDomain(domain *libvirt.Domain) error {
	if err := domain.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown domain: %w", err)
	}

	name, _ := domain.GetName()
	dm.logger.WithField("domain", name).Info("Domain shutdown initiated")
	return nil
}

// ForceStopDomain forcefully stops a VM
func (dm *DomainManager) ForceStopDomain(domain *libvirt.Domain) error {
	if err := domain.Destroy(); err != nil {
		return fmt.Errorf("failed to destroy domain: %w", err)
	}

	name, _ := domain.GetName()
	dm.logger.WithField("domain", name).Info("Domain destroyed")
	return nil
}

// DeleteDomain removes a VM permanently
func (dm *DomainManager) DeleteDomain(domain *libvirt.Domain) error {
	// First ensure it's stopped
	state, _, err := domain.GetState()
	if err != nil {
		return fmt.Errorf("failed to get domain state: %w", err)
	}

	if state == libvirt.DOMAIN_RUNNING {
		if err := dm.ForceStopDomain(domain); err != nil {
			return err
		}
	}

	// Undefine the domain
	if err := domain.Undefine(); err != nil {
		return fmt.Errorf("failed to undefine domain: %w", err)
	}

	name, _ := domain.GetName()
	dm.logger.WithField("domain", name).Info("Domain deleted")
	return nil
}

// GetDomainState returns the current state of a VM
func (dm *DomainManager) GetDomainState(domain *libvirt.Domain) (libvirt.DomainState, error) {
	state, _, err := domain.GetState()
	if err != nil {
		return 0, fmt.Errorf("failed to get domain state: %w", err)
	}
	return state, nil
}

// GetDomainInfo returns domain information
func (dm *DomainManager) GetDomainInfo(domain *libvirt.Domain) (*libvirt.DomainInfo, error) {
	info, err := domain.GetInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get domain info: %w", err)
	}
	return info, nil
}

// ListDomains returns all domains
func (dm *DomainManager) ListDomains() ([]libvirt.Domain, error) {
	return dm.client.ListAllDomains(0) // List all domains (active and inactive)
}

// GenerateDomainXML creates VM XML configuration
func (dm *DomainManager) GenerateDomainXML(name, uuid string, memoryMB, vcpus int, diskPath, networkName, macAddr string) (string, error) {
	domain := DomainXML{
		Type: "kvm",
		Name: name,
		UUID: uuid,
		Memory: MemoryXML{
			Unit:  "MiB",
			Value: fmt.Sprintf("%d", memoryMB),
		},
		CurrentMem: MemoryXML{
			Unit:  "MiB",
			Value: fmt.Sprintf("%d", memoryMB),
		},
		VCPU: VCPUXML{
			Placement: "static",
			Value:     fmt.Sprintf("%d", vcpus),
		},
		OS: OSXML{
			Type: OSTypeXML{
				Arch:    "x86_64",
				Machine: "pc-i440fx-2.9",
				Value:   "hvm",
			},
			Boot: []BootXML{
				{Dev: "hd"},
				{Dev: "cdrom"},
			},
			BootMenu: BootMenuXML{Enable: "yes"},
		},
		Features: FeaturesXML{},
		CPU: CPUXML{
			Mode:  "host-model",
			Check: "partial",
		},
		Clock: ClockXML{
			Offset: "utc",
		},
		OnPoweroff: "destroy",
		OnReboot:   "restart",
		OnCrash:    "restart",
		Devices: DevicesXML{
			Emulator: "/usr/bin/qemu-system-x86_64",
			Disks: []DiskXML{
				{
					Type:   "file",
					Device: "disk",
					Driver: DriverXML{
						Name: "qemu",
						Type: "qcow2",
					},
					Source: SourceXML{
						File: diskPath,
					},
					Target: TargetXML{
						Dev: "vda",
						Bus: "virtio",
					},
					Boot: BootOrderXML{Order: "1"},
				},
			},
			Interfaces: []InterfaceXML{
				{
					Type: "network",
					Source: IfSourceXML{
						Network: networkName,
					},
					Model: ModelXML{Type: "virtio"},
					MAC:   MACXML{Address: macAddr},
				},
			},
			Console: ConsoleXML{
				Type: "pty",
				Target: ConsoleTargetXML{
					Type: "serial",
					Port: "0",
				},
			},
			Graphics: GraphicsXML{
				Type:     "vnc",
				Autoport: "yes",
				Listen:   "127.0.0.1",
				KeyMap:   "en-us",
			},
			Video: VideoXML{
				Model: VideoModelXML{Type: "cirrus"},
			},
			MemBalloon: MemBalloonXML{Model: "virtio"},
		},
	}

	xmlData, err := xml.MarshalIndent(domain, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal domain XML: %w", err)
	}

	// Add XML declaration
	xmlString := `<?xml version="1.0" encoding="UTF-8"?>` + "\n" + string(xmlData)

	dm.logger.WithField("domain", name).Debug("Generated domain XML")
	return xmlString, nil
}

// StateToString converts libvirt domain state to string
func StateToString(state libvirt.DomainState) string {
	switch state {
	case libvirt.DOMAIN_NOSTATE:
		return "no_state"
	case libvirt.DOMAIN_RUNNING:
		return "running"
	case libvirt.DOMAIN_BLOCKED:
		return "blocked"
	case libvirt.DOMAIN_PAUSED:
		return "paused"
	case libvirt.DOMAIN_SHUTDOWN:
		return "shutdown"
	case libvirt.DOMAIN_SHUTOFF:
		return "shutoff"
	case libvirt.DOMAIN_CRASHED:
		return "crashed"
	case libvirt.DOMAIN_PMSUSPENDED:
		return "suspended"
	default:
		return "unknown"
	}
}

// CreateDomainWithCloudInit creates a new VM domain with cloud-init support
func (dm *DomainManager) CreateDomainWithCloudInit(name, uuid string, memoryMB, vcpus int, diskPath, networkName, cloudInitISOPath string) (*libvirt.Domain, error) {
	// Generate MAC address
	macAddr := dm.generateMACAddress()

	// Generate domain XML with cloud-init
	domainXML, err := dm.GenerateDomainXMLWithCloudInit(name, uuid, memoryMB, vcpus, diskPath, networkName, macAddr, cloudInitISOPath)
	if err != nil {
		return nil, fmt.Errorf("failed to generate domain XML: %w", err)
	}

	// Create domain
	domain, err := dm.CreateDomain(domainXML)
	if err != nil {
		return nil, fmt.Errorf("failed to create domain: %w", err)
	}

	dm.logger.WithFields(logrus.Fields{
		"name": name,
		"uuid": uuid,
		"mac":  macAddr,
	}).Info("Domain created with cloud-init support")

	return domain, nil
}

// GenerateDomainXMLWithCloudInit generates domain XML with cloud-init CD-ROM
func (dm *DomainManager) GenerateDomainXMLWithCloudInit(name, uuid string, memoryMB, vcpus int, diskPath, networkName, macAddr, cloudInitISOPath string) (string, error) {
	xmlTemplate := `<?xml version="1.0" encoding="UTF-8"?>
<domain type="kvm">
  <name>{{.Name}}</name>
  <uuid>{{.UUID}}</uuid>
  <memory unit="MiB">{{.MemoryMB}}</memory>
  <currentMemory unit="MiB">{{.MemoryMB}}</currentMemory>
  <vcpu placement="static">{{.VCPUs}}</vcpu>
  <os>
    <type arch="x86_64" machine="pc-q35-2.12">hvm</type>
    <boot dev="hd"/>
    <boot dev="cdrom"/>
  </os>
  <features>
    <acpi/>
    <apic/>
    <vmx state="on"/>
  </features>
  <cpu mode="host-model" check="partial"/>
  <clock offset="utc"/>
  <on_poweroff>destroy</on_poweroff>
  <on_reboot>restart</on_reboot>
  <on_crash>destroy</on_crash>
  <devices>
    <emulator>/usr/bin/qemu-system-x86_64</emulator>
    <disk type="file" device="disk">
      <driver name="qemu" type="qcow2"/>
      <source file="{{.DiskPath}}"/>
      <target dev="vda" bus="virtio"/>
      <boot order="1"/>
    </disk>
    <disk type="file" device="cdrom">
      <driver name="qemu" type="raw"/>
      <source file="{{.CloudInitISOPath}}"/>
      <target dev="sda" bus="sata"/>
      <readonly/>
      <boot order="2"/>
    </disk>
    <interface type="network">
      <source network="{{.NetworkName}}"/>
      <model type="virtio"/>
      <mac address="{{.MACAddr}}"/>
    </interface>
    <console type="pty">
      <target type="serial" port="0"/>
    </console>
    <graphics type="vnc" autoport="yes" listen="0.0.0.0"/>
    <video>
      <model type="qxl"/>
    </video>
    <memballoon model="virtio"/>
    <rng model="virtio">
      <backend model="random">/dev/urandom</backend>
    </rng>
  </devices>
</domain>`

	data := struct {
		Name             string
		UUID             string
		MemoryMB         int
		VCPUs            int
		DiskPath         string
		NetworkName      string
		MACAddr          string
		CloudInitISOPath string
	}{
		Name:             name,
		UUID:             uuid,
		MemoryMB:         memoryMB,
		VCPUs:            vcpus,
		DiskPath:         diskPath,
		NetworkName:      networkName,
		MACAddr:          macAddr,
		CloudInitISOPath: cloudInitISOPath,
	}

	tmpl, err := template.New("domain").Parse(xmlTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse domain template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute domain template: %w", err)
	}

	return buf.String(), nil
}

// generateMACAddress generates a unique MAC address
func (dm *DomainManager) generateMACAddress() string {
	// Generate a random MAC address in the QEMU range (52:54:00:xx:xx:xx)
	// This is a simple implementation - in production you'd want better uniqueness
	rand.Seed(time.Now().UnixNano())

	// QEMU MAC prefix: 52:54:00
	// Generate random bytes for the last 3 octets
	octet1 := rand.Intn(256)
	octet2 := rand.Intn(256)
	octet3 := rand.Intn(256)

	return fmt.Sprintf("52:54:00:%02x:%02x:%02x", octet1, octet2, octet3)
}
