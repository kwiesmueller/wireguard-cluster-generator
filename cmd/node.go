package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"text/template"

	"github.com/rancher/k3os/pkg/config"
	"gopkg.in/yaml.v2"
)

type NodeConfig struct {
	// ID of the node used to set its internal IP etc.
	ID uint `yaml:"id"`
	// Region the master is located in
	Region string `yaml:"region,omitempty"`

	DNSNameservers []string `yaml:"dns_nameservers,omitempty"`
	NTPServers     []string `yaml:"ntp_servers,omitempty"`
	Disks          []Disk   `yaml:"disks,omitempty"`

	WireGuard WireGuardConfig `yaml:"wireguard,omitempty"`
}

type Disk struct {
	Path string `yaml:"path,omitempty"`
	UUID string `yaml:"uuid,omitempty"`
}

func DecodeNodeConfig(path string) (NodeConfig, error) {
	nodeConfigFile, err := os.Open(path)
	if err != nil {
		return NodeConfig{}, err
	}

	var nodeConfig NodeConfig
	if err := yaml.NewDecoder(nodeConfigFile).Decode(&nodeConfig); err != nil {
		return NodeConfig{}, err
	}
	return nodeConfig, nil
}

func (cfg NodeConfig) Finalize(master Master) (Node, error) {
	node := Node{
		NodeConfig: cfg,
		IP: &net.IPNet{
			IP:   getNextIP(master.Network, 100+cfg.ID),
			Mask: master.NodeIPMask,
		},
		RouteNetwork:     master.RouteNetwork,
		MasterExternalIP: master.ExternalIP,
		MasterInternalIP: master.InternalIP.String(),
		WireguardPort:    master.ListenPort,
		MasterPublicKey:  master.WireGuard.PublicKey,
		SharedConfig:     master.SharedConfig,
	}

	if cfg.DNSNameservers == nil {
		node.DNSNameservers = master.DNSNameservers
	}
	if cfg.NTPServers == nil {
		node.NTPServers = master.NTPServers
	}

	if len(node.WireGuard.PrivateKey) < 1 {
		var err error
		node.WireGuard.PrivateKey, node.WireGuard.PublicKey, err = generateKeyPair()
		if err != nil {
			return Node{}, err
		}
	}

	return node, nil
}

type Node struct {
	NodeConfig

	MasterExternalIP string
	MasterInternalIP string
	MasterPublicKey  string
	SharedConfig

	IP            *net.IPNet
	RouteNetwork  *net.IPNet
	WireguardPort string
}

func (node Node) Generate() (config.CloudConfig, error) {
	wireguardConf, err := node.generateWireguardConf()
	if err != nil {
		return config.CloudConfig{}, err
	}
	wireguardInit, err := node.generateWireguardInit()
	if err != nil {
		return config.CloudConfig{}, err
	}

	files := []config.File{
		{
			Content:            wireguardConf,
			RawFilePermissions: "0555",
			Path:               "/etc/wireguard/wg0.conf",
		},
		{
			Content:            wireguardInit,
			RawFilePermissions: "0555",
			Path:               "/etc/wireguard/init-server",
		},
	}

	files = append(files, node.generateFstabEntries()...)

	return config.CloudConfig{
		SSHAuthorizedKeys: node.SSHAuthorizedKeys,
		Hostname:          fmt.Sprintf("node-%d", node.ID),
		WriteFiles:        files,
		Runcmd:            []string{"/etc/wireguard/init-server"},
		K3OS: config.K3OS{
			Modules: []string{
				"nvme",
				"wireguard",
			},
			Sysctls: map[string]string{
				"net.ipv4.ip_forward": "1",
			},
			DNSNameservers: node.DNSNameservers,
			NTPServers:     node.NTPServers,
			ServerURL:      fmt.Sprintf("https://%s:6443", node.MasterInternalIP),
			Token:          node.ClusterToken,
			K3sArgs: []string{
				"--token=" + node.ClusterToken,
				"--flannel-iface=wg0",
			},
			Labels: map[string]string{
				"region": node.Region,
			},
		},
	}, nil
}

func (node Node) generateWireguardConf() (string, error) {
	return node.executeNodeTemplate("./data/node/wg0.conf.template")
}

func (node Node) generateWireguardInit() (string, error) {
	return node.executeNodeTemplate("./data/node/init-server.template")
}

func (node Node) generateFstabContent() (string, error) {
	return node.executeNodeTemplate("./data/node/fstab.template")
}

func (node Node) generateFstabEntries() []config.File {
	var files []config.File
	if len(node.Disks) > 0 {
		content, err := node.generateFstabContent()
		if err != nil {
			return files
		}
		files = append(files, config.File{
			Content:            content,
			Owner:              "root:root",
			Path:               "/etc/fstab",
			RawFilePermissions: "0644",
		})
	}
	return files
}

func (node Node) executeNodeTemplate(path string) (string, error) {
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	if err := tmpl.Execute(&output, node); err != nil {
		return "", err
	}
	return output.String(), nil
}
