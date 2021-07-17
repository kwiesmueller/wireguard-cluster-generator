# (Wireguard) Cluster Generator

This repository is meant to help with generating [K3OS](https://github.com/rancher/k3os) configs
for setting up a wireguard connected Kubernetes cluster.

**NOTE:** This repository is based on my own experimentation and more a sandbox project than anything else.
Feel free to open issues or get in touch if you are trying similar things. I'm happy to hear feedback and ideas.

## Quickstart

Place your configs in a custom repository (e.g. `cluster-configs`). You can copy the [examples](./examples) folder for a reference directory structure.

```sh
export BASE_DIR=<path-to-cluster-configs>
export CLUSTER_NAME=<your-cluster-name>
make configs
```

To also generate Raspberry Pi images, checkout the [picl-k3os-image-generator fork](https://github.com/kwiesmueller/picl-k3os-image-generator)
and set it's path as `IMAGE_GENERATOR`

```sh
export IMAGE_GENERATOR=${HOME}/git/kwiesmueller/picl-k3os-image-generator
# select which of your nodes will need Raspberry Pi images
export RASPBERRY_NODES=node-1,node-2

make images
```

Find your images and configs in `<cluster-configs>/out/<your-cluster-name>`.

## Configuration

The helper tool in [/cmd/generate.go](./cmd/generate.go) takes a master configuration and a folder
containing node configurations from which it generates a single k3os config per machine.

You can refer to [examples](./examples) for a reference on how to build your configs.

## Generate

To generate your machine configs, run:

```sh
go run cmd/*.go --masterConfig=examples/master.yaml --nodeConfigDir=examples/nodes --outDir=examples/out
```

## Use configs

Now you can start providing the resulting configs to your nodes.

**Keep in mind that those configs contain Wireguard secrets, so handle them with care.**

A good way to install K3OS from a config, starting with your master, is to mount the ISO file provided on [GitHub](https://github.com/rancher/k3os/releases).
Then login to the machine using the `rancher` account and enable SSH for your username by running:

```sh
curl https://github.com/<username>.keys | tee -a ~/.ssh/authorized_keys
```

Then you can simply upload the config to your machine and run the installer:

```sh
scp examples/out/master.cloudconfig.yaml rancher@<master-ip>:
ssh rancher@<master-ip>
> sudo k3os install
```

When asked select the install to disk option and provide the config you just uploaded.

### Raspberry Pi Nodes

Running K3OS on a Raspberry Pi requires a bit more work, but is still easy.
First you need https://github.com/kwiesmueller/picl-k3os-image-generator which is a fork with some modifications supporting this repo.

To generate images for your nodes, copy all configs supposed to run on Raspberry Pis into the config folder in [picl-k3os-image-generator](https://github.com/kwiesmueller/picl-k3os-image-generator).

Then build the imagebuilder `docker build . -t picl-builder:latest`.

And generate an image for every node `./build-images.sh`.

Those images can then be flashed onto your respective device.

#### Flashing (eMMC edition)

Prepare the Raspberry Pi for flashing.
1. Get https://github.com/raspberrypi/usbboot.
2. Run `sudo ./rpiboot`.
3. Find your device `lsblk`.
4. Flash to your device `sudo dd bs=4M if=out/picl-k3os-node-1-v0.21.1-k3s1r0-raspberrypi.img | pv | sudo dd of=/dev/<yourraspberry>`

## Using

Now switch your Pi back into boot mode and just start.
Ideally I didn't make any mistake writing this and it just works.
It should prepare its disks and then just connect to your master via wireguard.

Right now I only tested this with a amd64 master running centrally in the cloud and a single Raspberry Pi Compute Module 4
placed on either an IO Board or right now a https://hackaday.io/project/177626-mirkopc-cm4-carrier-board which I can highly recommend.

