package app

import (
	"bytes"

	"github.com/dgraph-io/badger"

	abcitypes "github.com/tendermint/tendermint/abci/types"
)

// KVStoreApplication KVStoreApplication
type KVStoreApplication struct {
	db           *badger.DB
	currentBatch *badger.Txn
}

var _ abcitypes.Application = (*KVStoreApplication)(nil)

// NewKVStoreApplication NewKVStoreApplication
func NewKVStoreApplication(db *badger.DB) *KVStoreApplication {
	return &KVStoreApplication{
		db: db,
	}
}

// SetOption SetOption
func (KVStoreApplication) SetOption(req abcitypes.RequestSetOption) abcitypes.ResponseSetOption {
	return abcitypes.ResponseSetOption{}
}

// CheckTx Mempool Connection
func (app *KVStoreApplication) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	code := app.isValid(req.Tx)
	return abcitypes.ResponseCheckTx{Code: code, GasWanted: 1}
}

// Info Return application info
func (KVStoreApplication) Info(abcitypes.RequestInfo) abcitypes.ResponseInfo {
	return abcitypes.ResponseInfo{}
}

// Query for state
func (app KVStoreApplication) Query(reqQUery abcitypes.RequestQuery) (resQuery abcitypes.ResponseQuery) {
	resQuery.Key = reqQUery.Data
	err := app.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(reqQUery.Data)
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		if err == badger.ErrKeyNotFound {
			resQuery.Log = "does not exist"
		} else {
			return item.Value(func(val []byte) error {
				resQuery.Log = "exists"
				resQuery.Value = val
				return nil
			})
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return
}

// InitChain Initialize blockchain w validators/other info from TendermintCore
func (KVStoreApplication) InitChain(abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	return abcitypes.ResponseInitChain{}
}

// BeginBlock Signals the beginning of a block
func (app KVStoreApplication) BeginBlock(abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	app.currentBatch = app.db.NewTransaction(true)
	return abcitypes.ResponseBeginBlock{}
}

// DeliverTx Deliver a tx for full processing
func (app KVStoreApplication) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	code := app.isValid(req.Tx)
	if code != 0 {
		return abcitypes.ResponseDeliverTx{Code: code}
	}
	parts := bytes.Split(req.Tx, []byte("="))
	key, value := parts[0], parts[1]
	err := app.currentBatch.Set(key, value)
	if err != nil {
		panic(err)
	}
	return abcitypes.ResponseDeliverTx{Code: 0}
}

// EndBlock Signals the end of a block, returns changes to the validator set
func (KVStoreApplication) EndBlock(abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	return abcitypes.ResponseEndBlock{}
}

// Commit the state and return the application Merkle root hash
func (app KVStoreApplication) Commit() abcitypes.ResponseCommit {
	app.currentBatch.Commit()
	return abcitypes.ResponseCommit{Data: []byte{}}
}

// ListSnapshots available snapshots
func (KVStoreApplication) ListSnapshots(abcitypes.RequestListSnapshots) abcitypes.ResponseListSnapshots {
	return abcitypes.ResponseListSnapshots{}
}

// OfferSnapshot Offer a snapshot to the application
func (KVStoreApplication) OfferSnapshot(abcitypes.RequestOfferSnapshot) abcitypes.ResponseOfferSnapshot {
	return abcitypes.ResponseOfferSnapshot{}
}

// LoadSnapshotChunk Load a snapshot chunk
func (KVStoreApplication) LoadSnapshotChunk(abcitypes.RequestLoadSnapshotChunk) abcitypes.ResponseLoadSnapshotChunk {
	return abcitypes.ResponseLoadSnapshotChunk{}
}

// ApplySnapshotChunk Apply a shapshot chunk
func (KVStoreApplication) ApplySnapshotChunk(abcitypes.RequestApplySnapshotChunk) abcitypes.ResponseApplySnapshotChunk {
	return abcitypes.ResponseApplySnapshotChunk{}
}

func (app *KVStoreApplication) isValid(tx []byte) (code uint32) {
	// check format
	parts := bytes.Split(tx, []byte("="))
	if len(parts) != 2 {
		return 1
	}

	key, value := parts[0], parts[1]

	// check if the same key=value already exists
	err := app.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		if err == nil {
			return item.Value(func(val []byte) error {
				if bytes.Equal(val, value) {
					code = 2
				}
				return nil
			})
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	return code
}
