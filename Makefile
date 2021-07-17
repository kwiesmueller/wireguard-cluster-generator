generate:

BASE_DIR ?= .
CONFIG_DIR ?= ${BASE_DIR}
OUT_DIR ?= ${BASE_DIR}/out
CLUSTER_NAME ?= cluster-1
IMAGE_GENERATOR ?= ${HOME}/git/sgielen/picl-k3os-image-generator
RASPBERRY_NODES ?= node-1

configs:
	go run cmd/generate/*.go \
	--masterConfig=${CONFIG_DIR}/${CLUSTER_NAME}/master.yaml \
	--nodeConfigDir=${CONFIG_DIR}/${CLUSTER_NAME}/nodes \
	--outDir=${OUT_DIR}/${CLUSTER_NAME}/configs

images:
	go run cmd/image/*.go \
	--configDir=${OUT_DIR}/${CLUSTER_NAME}/configs \
	--raspberryNodes=${RASPBERRY_NODES} \
	--imageGenerator=${IMAGE_GENERATOR} \
	--outDir=${OUT_DIR}/${CLUSTER_NAME}/images

.PHONY: configs images