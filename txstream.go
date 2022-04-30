package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/wire"
)

/*
var (
	blockDir      = "/media/hdd1/bitcoin/blocks"
	blockIndexDir = "/media/hdd1/bitcoin/blocks/index"
	chainStateDir = "/media/hdd1/bitcoin/chainstate"
	indexFile     = "blockPositionIndex"
)
*/
func txStream() {
	idxFile, err := os.Open(indexFile)
	if err != nil {
		panic(err)
	}

	info, err := idxFile.Stat()
	if err != nil {
		panic(err)
	}
	maxHeight := info.Size() / 6 // max height based on index file size

	curBlockFile := new(os.File)
	curBlockFileNum := uint16(1 << 15)

	var fileNum uint16
	var offset uint32
	var idxChunk [6]byte

	for height := int64(1); height < maxHeight; height++ { // loop through all blocks
		_, err = idxFile.ReadAt(idxChunk[:], height*6)
		if err != nil {
			panic(err)
		}
		fileNum = binary.BigEndian.Uint16(idxChunk[0:2])
		offset = binary.BigEndian.Uint32(idxChunk[2:6])

		if fileNum != curBlockFileNum { // need to switch files
			if curBlockFile != nil { // close if it's open
				curBlockFile.Close()
			}
			fileName := fmt.Sprintf("blk%05d.dat", fileNum)
			filePath := filepath.Join(blockDir, fileName)
			curBlockFile, err = os.Open(filePath)
			if err != nil {
				panic(err)
			}
		}
		_, err := curBlockFile.Seek(int64(offset), 0)
		if err != nil {
			panic(err)
		}
		block := new(wire.MsgBlock)
		err = block.Deserialize(curBlockFile)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", block.BlockHash().String())
	}

}
