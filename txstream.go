package main

import (
	"crypto/md5"
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
	maxHeight := 100000
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

type outpointHash [16]byte
type pksHash [16]byte

func (p pksHash) toString() string {
	return fmt.Sprintf("%x", p[:])
}

func hashOutpoint(op wire.OutPoint) (oh outpointHash) {
	var opBytes [36]byte
	copy(opBytes[:32], op.Hash[:])
	binary.BigEndian.PutUint32(opBytes[32:], op.Index)
	return md5.Sum(opBytes[:])
}
func hashPKS(pks []byte) pksHash {
	return md5.Sum(pks)
}

func graphGenerate(txChan chan *wire.MsgTx) {
	m := make(map[outpointHash]pksHash)
	idxFile, err := os.Open(indexFile)
	if err != nil {
		panic(err)
	}

	//info, err := idxFile.Stat()
	if err != nil {
		panic(err)
	}
	//maxHeight := info.Size() / 6 // max height based on index file size
	maxHeight := 100000

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
			hash := tx.TxHash()
			fromNodes := make([]pksHash, 0)
			toNodes := make([]pksHash, 0)
			for row := 0; row < len(tx.TxIn) || row < len(tx.TxOut); row++ {
				//fmt.Println(m)
				if row < len(tx.TxIn) {
					//index := fmt.Sprintf(",%d",row);
					//s := hash + index
					//fmt.Println("looking at in")
					//fmt.Println(tx.TxIn[row].PreviousOutPoint.String())
					prevAddress, pres :=
						m[hashOutpoint(tx.TxIn[row].PreviousOutPoint)]
					/*if pres {
						fmt.Println(inAddress)
					}*/
					if pres {
						//fmt.Println(inAddress)
						fromNodes = append(fromNodes, prevAddress)
					}
				} else {
					//fmt.Printf("\t\t\t\t\t\t\t\t -> ")
				}
				if row < len(tx.TxOut) {
					op := hashOutpoint(*wire.NewOutPoint(&hash, uint32(row)))
					// address := fmt.Sprintf("%x", tx.TxOut[row].PkScript)
					// amt := fmt.Sprintf("%d", tx.TxOut[row].Value)
					//fmt.Println(s)
					m[op] = hashPKS(tx.TxOut[row].PkScript)
					toNodes = append(toNodes, m[op])
				} else {
					//fmt.Printf("\n")
				}
			}
			/*fmt.Println("edges from ")
			fmt.Println(fromNodes)
			fmt.Println(" to ")
			fmt.Println(toNodes)*/
			/*if len(fromNodes) != 0 {
				fmt.Println(len(fromNodes))
				fmt.Println(len(toNodes))
				fmt.Println()
			}*/
			if len(fromNodes) != 0 {
				for i := 0; i < len(fromNodes); i++ {
					for j := 0; j < len(toNodes); j++ {
						//fmt.Println(fromNodes[i])
						fmt.Println(
							fromNodes[i].toString() + " " + toNodes[j].toString())
					}
					//fmt.Println(fromNodes[i])
				}
			}
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
