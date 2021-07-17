package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	configDir         = flag.String("configDir", "./out/configs", "The directory containing all configs")
	raspberryNodes    = flag.String("raspberryNodes", "", "A comma seperated list of node names that will get raspberry pi images created for them. Example: node-1,node-2")
	outDir            = flag.String("outDir", "./out/images", "The directory images are generated into")
	imageGeneratorDir = flag.String("imageGenerator", "", "The directory containing https://github.com/kwiesmueller/picl-k3os-image-generator")
)

func main() {
	flag.Parse()

	if len(*configDir) < 1 {
		logErrAndExit(errors.New("configDir required"))
	}

	if len(*imageGeneratorDir) < 1 {
		logErrAndExit(errors.New("imageGenerator required"))
	}

	outDir, err := filepath.Abs(*outDir)
	if err != nil {
		logErrAndExit(err)
	}

	if _, err := os.Stat(outDir); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(outDir, 0770); err != nil {
			logErrAndExit(err)
		}
	}

	nodes := strings.Split(*raspberryNodes, ",")

	configs := make(map[string]string)
	if err := filepath.WalkDir(*configDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		filename := filepath.Base(path)
		parts := strings.Split(filename, ".")
		if len(parts) < 1 {
			return errors.New("invalid filename: " + filename)
		}
		configs[parts[0]] = path
		return nil
	}); err != nil {
		logErrAndExit(err)
	}

	targetNodes := make(map[string]string)
	for _, node := range nodes {
		path, found := configs[node]
		if !found {
			logErrAndExit(errors.New("couldn't find config for node " + node))
		}
		targetNodes[node] = path
	}

	imageConfigDir := filepath.Join(*imageGeneratorDir, "config")
	for node, path := range targetNodes {
		inFile, err := os.Open(path)
		if err != nil {
			logErrAndExit(err)
		}

		outFile, err := os.OpenFile(filepath.Join(imageConfigDir, node+".raspberrypi.yaml"), os.O_CREATE|os.O_RDWR, 0770)
		if err != nil {
			logErrAndExit(err)
		}

		if _, err := io.Copy(outFile, inFile); err != nil {
			logErrAndExit(err)
		}
	}

	buildImages := exec.Command("./build-images.sh")
	buildImages.Dir = *imageGeneratorDir
	buildImages.Stderr = os.Stderr
	buildImages.Stdout = os.Stdout

	defer func() {
		cleanup := exec.Command("./cleanup-tempfiles.sh")
		cleanup.Dir = *imageGeneratorDir
		cleanup.Stderr = os.Stderr
		cleanup.Stdout = os.Stdout
		fmt.Println("cleaning up temporary files")
		if err := cleanup.Run(); err != nil {
			fmt.Println(err)
		}
	}()

	if err := buildImages.Run(); err != nil {
		logErrAndExit(err)
	}

	if err := filepath.WalkDir(filepath.Join(*imageGeneratorDir, "out"), func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if ext := filepath.Ext(path); ext != ".img" {
			fmt.Println("invalid extension " + ext)
			return nil
		}
		fmt.Printf("moving image %s to %s\n", path, outDir)
		return os.Rename(path, filepath.Join(outDir, filepath.Base(path)))
	}); err != nil {
		logErrAndExit(err)
	}
}

func logErrAndExit(err error) {
	fmt.Println(err)
	os.Exit(1)
}
