package main

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/btcsuite/btcd/wire"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

/*
https://bitcoin.stackexchange.com/questions/28168/what-are-the-keys-used-in-the-blockchain-leveldb-ie-what-are-the-keyvalue-pair
*/

var blocksDir = "/media/hdd1/bitcoin/blocks"

func blockread() error {
	lvdb, err := OpenIndexFile(blocksDir)
	if err != nil {
		return err
	}
	readDB(lvdb)
	lvdb.Close()
	return nil
}

// BufferDB buffers the leveldb key values into map in memory
func readDB(lvdb *leveldb.DB) {
	fmt.Printf("iterating db\n")
	iter := lvdb.NewIterator(util.BytesPrefix([]byte{'b'}), nil)
	for iter.Next() {
		k := iter.Key()

		var revhash [32]byte
		for i, _ := range revhash {
			revhash[i] = k[32-i]
		}

		v := iter.Value()
		buf := bytes.NewBuffer(v)
		hed := new(wire.BlockHeader)
		err := hed.Deserialize(buf)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
		}

		// v = v[80:]
		fmt.Printf("%x\t val %x\n", revhash, v)
		fmt.Printf("%s\n", hed.PrevBlock.String())
	}

	iter.Release()
	err := iter.Error()
	if err != nil {
		panic(err)
	}

	return
}

// OpenIndexFile returns the db with only read only option enabled
func OpenIndexFile(dataDir string) (*leveldb.DB, error) {
	indexDir := filepath.Join(dataDir, "/index")
	o := opt.Options{ReadOnly: true, Compression: opt.NoCompression}
	lvdb, err := leveldb.OpenFile(indexDir, &o)
	if err != nil {
		return nil, fmt.Errorf("can't open %s. err:%s", indexDir, err)
	}

	return lvdb, nil
}
