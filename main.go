package main

import (
	"errors"
	"os"
)

var (
	blockDir      = "/media/hdd1/bitcoin/blocks"
	blockIndexDir = "/media/hdd1/bitcoin/blocks/index"
	chainStateDir = "/media/hdd1/bitcoin/chainstate"
	indexFile     = "blockPositionIndex"
)

func main() {
	_, err := os.Stat(indexFile)
	if errors.Is(err, os.ErrNotExist) {
		buildBlockIndex()
	}
	txStream()
}
