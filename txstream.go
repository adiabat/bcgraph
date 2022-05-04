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
func txStream(txChan chan *wire.MsgTx) {
	idxFile, err := os.Open(indexFile)
	if err != nil {
		panic(err)
	}

	//info, err := idxFile.Stat()
	if err != nil {
		panic(err)
	}
	//maxHeight := info.Size() / 6 // max height based on index file size
	maxHeight := 100
	curBlockFile := new(os.File)
	curBlockFileNum := uint16(1 << 15)

	var fileNum uint16
	var offset uint32
	var idxChunk [6]byte

	for height := int64(1); height < int64(maxHeight); height++ { // loop through all blocks
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

		// fmt.Printf("%s\n", block.BlockHash().String())
		for _, tx := range block.Transactions {
			txChan <- tx
		}
	}
}

func graphGenerate(txChan chan *wire.MsgTx) {
	m := make(map[string]string)
	idxFile, err := os.Open(indexFile)
	if err != nil {
		panic(err)
	}

	//info, err := idxFile.Stat()
	if err != nil {
		panic(err)
	}
	//maxHeight := info.Size() / 6 // max height based on index file size
	maxHeight := 8000

	curBlockFile := new(os.File)
	curBlockFileNum := uint16(1 << 15)

	var fileNum uint16
	var offset uint32
	var idxChunk [6]byte

	for height := int64(1); height < int64(maxHeight); height++ { // loop through all blocks
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
		// fmt.Printf("%s\n", block.BlockHash().String())
		for _, tx := range block.Transactions {
			hash := tx.TxHash().String()
			fromNodes := make([]string, len(block.Transactions))
			toNodes := make([]string, len(block.Transactions))
			for row := 0; row < len(tx.TxIn) || row < len(tx.TxOut); row++ {
				//fmt.Println(m)
				if row < len(tx.TxIn) {
					//index := fmt.Sprintf(",%d",row);
					//s := hash + index
					//fmt.Println("looking at in")
					//fmt.Println(tx.TxIn[row].PreviousOutPoint.String())
					inAddress, _ := m[tx.TxIn[row].PreviousOutPoint.String()]
					fromNodes = append(fromNodes, inAddress)
					//fmt.Printf("%s -> ", tx.TxIn[row].PreviousOutPoint.String())
				} else {
					//fmt.Printf("\t\t\t\t\t\t\t\t -> ")
				}
				if row < len(tx.TxOut) {
					//fmt.Printf("%x:%d\n", tx.TxOut[row].PkScript, tx.TxOut[row].Value)
					index := fmt.Sprintf("%d", row)
					s := hash + ":" + index
					address := fmt.Sprintf("%x", tx.TxOut[row].PkScript)
					//fmt.Println(s)
					m[s] = address
					toNodes = append(toNodes, address)
				} else {
					//fmt.Printf("\n")
				}
			}
			fmt.Println("edges from ")
			fmt.Println(fromNodes)
			fmt.Println(" to ")
			fmt.Println(toNodes)
		}
	}
}
func txPrinter(txChan chan *wire.MsgTx) {
	for {
		tx := <-txChan
		fmt.Printf("--------%s\t%d in %d out\n", tx.TxHash().String(), len(tx.TxIn), len(tx.TxOut))
		for row := 0; row < len(tx.TxIn) || row < len(tx.TxOut); row++ {
			if row < len(tx.TxIn) {
				fmt.Printf("%s -> ", tx.TxIn[row].PreviousOutPoint.String())
			} else {
				fmt.Printf("\t\t\t\t\t\t\t\t -> ")
			}
			if row < len(tx.TxOut) {
				fmt.Printf("%x:%d\n", tx.TxOut[row].PkScript, tx.TxOut[row].Value)
			} else {
				fmt.Printf("\n")
			}
		}
		fmt.Printf("\n")
	}
}
