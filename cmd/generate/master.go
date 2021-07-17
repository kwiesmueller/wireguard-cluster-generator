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

type MasterConfig struct {
	ExternalIP string `yaml:"external_ip,omitempty"`
	// Network is the general base IP the cluster operates on (10.222.0.0)
	Network net.IP `yaml:"network,omitempty"`
	// MTU set on the Wireguard server
	MTU string `yaml:"mtu,omitempty"`
	// ListenPort of the Wireguard server
	ListenPort string `yaml:"listen_port,omitempty"`
	// Region the master is located in
	Region string `yaml:"region,omitempty"`

	DNSNameservers []string `yaml:"dns_nameservers,omitempty"`
	NTPServers     []string `yaml:"ntp_servers,omitempty"`

	WireGuard WireGuardConfig `yaml:"wireguard,omitempty"`

	SharedConfig `yaml:"shared_config,inline"`
}

type WireGuardConfig struct {
	// PrivateKey can be set to avoid regenerating new ones on every run
	PrivateKey string `yaml:"private_key,omitempty"`
	PublicKey  string `yaml:"public_key,omitempty"`
}

func DecodeMasterConfig(path string) (MasterConfig, error) {
	masterConfigFile, err := os.Open(*masterConfigPath)
	if err != nil {
		return MasterConfig{}, err
	}

	var masterConfig MasterConfig
	if err := yaml.NewDecoder(masterConfigFile).Decode(&masterConfig); err != nil {
		return MasterConfig{}, err
	}
	return masterConfig, nil
}

type SharedConfig struct {
	ClusterToken string `yaml:"cluster_token,omitempty"`

	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys,omitempty"`
}

func (cfg MasterConfig) Finalize() (Master, error) {
	master := Master{
		MasterConfig: cfg,
		RouteNetwork: &net.IPNet{},
		NodeIPMask:   net.CIDRMask(32, 32),
	}

	master.RouteNetwork.IP = cfg.Network
	master.RouteNetwork.Mask = net.CIDRMask(24, 32)
	internalIP := getNextIP(cfg.Network, 1)
	master.InternalIP = &internalIP

	if len(cfg.WireGuard.PrivateKey) < 1 {
		var err error
		master.WireGuard.PrivateKey, master.WireGuard.PublicKey, err = generateKeyPair()
		if err != nil {
			return Master{}, err
		}
	}

	return master, nil
}

type Master struct {
	MasterConfig
	// RouteNetwork is the /24 range of the NetworkIP
	RouteNetwork *net.IPNet
	// InternalIP generated from the NetworkIP
	InternalIP *net.IP

	NodeIPMask net.IPMask
	Nodes      []Node
}

func (master Master) AddNode(cfg NodeConfig) (Master, error) {
	node, err := cfg.Finalize(master)
	if err != nil {
		return master, err
	}
	master.Nodes = append(master.Nodes, node)
	return master, nil
}

func (master Master) Generate() (config.CloudConfig, error) {
	wireguardConf, err := master.generateWireguardConf()
	if err != nil {
		return config.CloudConfig{}, err
	}
	wireguardInit, err := master.generateWireguardInit()
	if err != nil {
		return config.CloudConfig{}, err
	}
	return config.CloudConfig{
		SSHAuthorizedKeys: master.SSHAuthorizedKeys,
		Hostname:          "master-0",
		WriteFiles: []config.File{
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
		},
		Runcmd: []string{"/etc/wireguard/init-server"},
		K3OS: config.K3OS{
			Modules: []string{
				"kvm",
				"nvme",
				"wireguard",
			},
			Sysctls: map[string]string{
				"net.ipv4.ip_forward": "1",
			},
			DNSNameservers: master.DNSNameservers,
			NTPServers:     master.NTPServers,
			Token:          master.ClusterToken,
			K3sArgs: []string{
				fmt.Sprintf("--advertise-address=%s", master.InternalIP),
				"--flannel-iface=wg0",
			},
			Labels: map[string]string{
				"region": master.Region,
			},
		},
	}, nil
}

func (master Master) generateWireguardConf() (string, error) {
	return master.executeMasterTemplate("./data/master/wg0.conf.template")
}

func (master Master) generateWireguardInit() (string, error) {
	return master.executeMasterTemplate("./data/master/init-server.template")
}

func (master Master) executeMasterTemplate(path string) (string, error) {
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	if err := tmpl.Execute(&output, master); err != nil {
		return "", err
	}
	return output.String(), nil
}
