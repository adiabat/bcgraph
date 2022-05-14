package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/wire"
)

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
	if dec {
		i := new(big.Int)
		return i.SetBytes(p[:]).String()
	}
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
	for {
		tx := <-txChan
		hash := tx.TxHash()
		fromNodes := make([]pksHash, 0)
		toNodes := make([]pksHash, 0)
		for row := 0; row < len(tx.TxIn) || row < len(tx.TxOut); row++ {
			if row < len(tx.TxIn) {
				prevAddress, pres :=
					m[hashOutpoint(tx.TxIn[row].PreviousOutPoint)]
				if pres {
					fromNodes = append(fromNodes, prevAddress)
				}
			}
			if row < len(tx.TxOut) {
				op := hashOutpoint(*wire.NewOutPoint(&hash, uint32(row)))

				m[op] = hashPKS(tx.TxOut[row].PkScript)
				toNodes = append(toNodes, m[op])
			}
		}
		// print out in clique format
		if len(fromNodes)+len(toNodes) == 1 {
			// fmt.Printf("%s has 1 node\n", tx.TxHash().String())
			continue
		}
		if clique {
			for _, node := range fromNodes {
				fmt.Printf("%s ", node.toString())
			}
			for _, node := range toNodes {
				fmt.Printf("%s ", node.toString())
			}
			fmt.Printf("\n")
		} else if len(fromNodes) != 0 { // quadratic blowup comes from this nested loop...
			for i := 0; i < len(fromNodes); i++ {
				for j := 0; j < len(toNodes); j++ {
					fmt.Println(
						fromNodes[i].toString() + " " + toNodes[j].toString())
				}
			}
		}
	}
}
