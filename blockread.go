package main

import (
	"bytes"
	"fmt"
	"io"

	"github.com/btcsuite/btcd/wire"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

/*
https://bitcoin.stackexchange.com/questions/28168/what-are-the-keys-used-in-the-blockchain-leveldb-ie-what-are-the-keyvalue-pair
*/

var (
	blockDir      = "/media/hdd1/bitcoin/blocks/index"
	blockIndexDir = "/media/hdd1/bitcoin/blocks/index"
	chainStateDir = "/media/hdd1/bitcoin/chainstate"
)

func blockread() error {
	chainstateDb, err := OpenDB(chainStateDir)
	if err != nil {
		return err
	}
	tip := getLastHash(chainstateDb)
	chainstateDb.Close()

	blockIndexDb, err := OpenDB(blockIndexDir)
	if err != nil {
		return err
	}
	hmap := dumpDBAllHeaders(blockIndexDb)
	blockIndexDb.Close()

	fmt.Printf("hmap len %d, tip %x\n", len(hmap), tip)
	fval := hmap[tip]
	fmt.Printf("last val %x\n", fval)

	buildChainBackwards(tip, hmap)
	return nil
}

func buildChainBackwards(tip [32]byte, hmap map[[32]byte][]byte) {

	h := tip
	height := int64(9)
	for height > 1 { // stop when you get to height=1
		v := hmap[h]
		buf := bytes.NewBuffer(v[len(v)-80:])
		hed := new(wire.BlockHeader)
		err := hed.Deserialize(buf)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
		}
		// fmt.Printf("%s xtra: %x\n", hed.BlockHash().String(), v[:len(v)-80])
		xbuf := bytes.NewBuffer(v[:len(v)-80])
		_, _ = varIntToInt(xbuf)
		height, _ = varIntToInt(xbuf)
		_, _ = varIntToInt(xbuf)
		_, _ = varIntToInt(xbuf)
		filenum, _ := varIntToInt(xbuf)
		fileoffset, _ := varIntToInt(xbuf)
		fmt.Printf("%x h:%d\tfile:%d\toff:%d\n", h, height, filenum, fileoffset)

		h = hed.PrevBlock
	}
}

func varIntToInt(r io.Reader) (int64, int) {
	var n int64
	var size int
	for {
		var val [1]byte
		r.Read(val[:])
		size++
		n = (n << 7) | int64(val[0]&0x7f)
		if val[0]&0x80 != 0x80 {
			break
		}
		n++
	}

	return n, size
}

// getLastHash gets the last hash, or tip, of the blockchain.
// this is 1-shot so the tip can never get reorg'd out.
// it assumes everyting works, blockchain-wise
func getLastHash(lvdb *leveldb.DB) (blockHash [32]byte) {
	// the chainstate DB uses an obfuscation key, so get that
	obkey, err := lvdb.Get([]byte{0x0e, 0x00,
		'o', 'b', 'f', 'u', 's', 'c', 'a', 't', 'e', '_',
		'k', 'e', 'y'}, nil)
	if err != nil {
		panic(err)
	}
	obkey = obkey[1:]
	// fmt.Printf("obkey %x len %d\n", obkey, len(obkey))
	val, err := lvdb.Get([]byte{'B'}, nil)
	if err != nil {
		panic(err)
	}
	deObs(obkey, val) // deobfuscate tip hash
	copy(blockHash[:], val)
	return
}

// deobfuscate a byte string in place
func deObs(obkey, input []byte) {
	oblen := len(obkey)
	inlenMultiples := len(input) / oblen
	for i := 0; i < inlenMultiples; i++ {
		for j := 0; j < oblen; j++ {
			input[(i*oblen)+j] ^= obkey[j]
		}
	}
	return
}

// dumpDBAllHeaders reads all the headers from the blocks/index DB and returns them
// a map
// Actually the orders is sorted by the *LSB* of the block hash.  Which is silly.
func dumpDBAllHeaders(lvdb *leveldb.DB) (hmap map[[32]byte][]byte) {

	hmap = make(map[[32]byte][]byte)
	var curHash [32]byte
	fmt.Printf("iterating db\n")
	iter := lvdb.NewIterator(util.BytesPrefix([]byte{'b'}), nil)
	for iter.Next() {
		copy(curHash[:], iter.Key()[1:]) // <- you couldn't do that before right?
		v := make([]byte, len(iter.Value()))
		copy(v[:], iter.Value()) // blarg have to copy everything because leveldb
		hmap[curHash] = v
	}
	/*
			k := iter.Key()

			var revhash [32]byte
			for i, _ := range revhash {
				revhash[i] = k[32-i]
			}

			v := iter.Value()
			buf := bytes.NewBuffer(v[len(v)-80:])
			hed := new(wire.BlockHeader)
			err := hed.Deserialize(buf)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
			}

			// v = v[80:]
			// fmt.Printf("%x\n", revhash)
			fmt.Printf("%s xtra: %x\n", hed.BlockHash().String(), v[:len(v)-80])
		}
	*/
	iter.Release()
	err := iter.Error()
	if err != nil {
		panic(err)
	}
	fmt.Printf("got %d headers\n", len(hmap))
	return
}

// OpenDB returns the db with only read only option enabled
func OpenDB(dataDir string) (*leveldb.DB, error) {
	o := opt.Options{ReadOnly: true, Compression: opt.NoCompression}
	lvdb, err := leveldb.OpenFile(dataDir, &o)
	if err != nil {
		return nil, fmt.Errorf("can't open %s. err:%s", dataDir, err)
	}
	return lvdb, nil
}
