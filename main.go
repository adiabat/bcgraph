package main

import (
	"errors"
	"flag"
	"os"

	"github.com/btcsuite/btcd/wire"
)

var (
	blockDir      = "/media/hdd1/bitcoin/blocks"
	blockIndexDir = "/media/hdd1/bitcoin/blocks/index"
	chainStateDir = "/media/hdd1/bitcoin/chainstate"
	indexFile     = "blockPositionIndex"

	clique = false
	dec    = false
)

func main() {

	cliqueFlag := flag.Bool("c", false,
		"print clique per line instead of edge per line")
	decFlag := flag.Bool("d", false, "print decimal instead of hex")
	flag.Parse()
	clique = *cliqueFlag
	dec = *decFlag

	_, err := os.Stat(indexFile)
	if errors.Is(err, os.ErrNotExist) {
		buildBlockIndex()
	}

	txChan := make(chan *wire.MsgTx, 8)
	go graphGenerate(txChan)
	txStream(txChan)
}
