package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rancher/k3os/pkg/config"
)

var (
	masterConfigPath = flag.String("masterConfig", "./master.yaml", "The config file describing the master")
	nodeConfigDir    = flag.String("nodeConfigDir", "./nodes", "The directory containing node config YAML files")
	outDir           = flag.String("outDir", "./out", "The directory cloud-config files are generated into")
)

func main() {
	flag.Parse()

	fmt.Println("processing master at", *masterConfigPath)
	masterConfig, err := DecodeMasterConfig(*masterConfigPath)
	if err != nil {
		logErrAndExit(err)
	}

	master, err := masterConfig.Finalize()
	if err != nil {
		logErrAndExit(err)
	}

	fmt.Println("processing nodes in", *nodeConfigDir)
	if err := filepath.WalkDir(*nodeConfigDir, func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}
		nodeConfig, err := DecodeNodeConfig(path)
		if err != nil {
			return err
		}

		master, err = master.AddNode(nodeConfig)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {

		logErrAndExit(err)
	}

	generatedMaster, err := master.Generate()
	if err != nil {
		logErrAndExit(err)
	}

	if err := encodeCloudConfig("master.yaml", generatedMaster); err != nil {
		logErrAndExit(err)
	}

	for _, node := range master.Nodes {
		generatedNode, err := node.Generate()
		if err != nil {
			logErrAndExit(err)
		}
		if encodeCloudConfig(
			fmt.Sprintf("node-%d.yaml", node.ID),
			generatedNode,
		); err != nil {
			logErrAndExit(err)
		}
	}
}

func encodeCloudConfig(filename string, cfg config.CloudConfig) error {
	file, err := os.OpenFile(
		filepath.Join(*outDir, filename),
		os.O_CREATE|os.O_RDWR|os.O_TRUNC,
		0777,
	)
	if err != nil {
		return err
	}

	if err := config.Write(cfg, file); err != nil {
		return err
	}
	return nil
}

func logErrAndExit(err error) {
	fmt.Println(err)
	os.Exit(1)
}
