package ledger

import (
	. "github.com/Ontology/common"
	"github.com/Ontology/core/account"
	. "github.com/Ontology/core/asset"
	tx "github.com/Ontology/core/transaction"
	"github.com/Ontology/core/transaction/utxo"
	"github.com/Ontology/crypto"
)

// ILedgerStore provides func with store package.
type ILedgerStore interface {
	//TODO: define the state store func
	SaveBlock(b *Block, ledger *Ledger) error
	GetBlock(hash Uint256) (*Block, error)
	BlockInCache(hash Uint256) bool
	GetBlockHash(height uint32) (Uint256, error)
	InitLedgerStore(ledger *Ledger) error
	IsDoubleSpend(tx *tx.Transaction) bool

	//SaveHeader(header *Header,ledger *Ledger) error
	AddHeaders(headers []Header, ledger *Ledger) error
	GetHeader(hash Uint256) (*Header, error)

	GetTransaction(hash Uint256) (*tx.Transaction, error)

	SaveAsset(assetid Uint256, asset *Asset) error
	GetAsset(hash Uint256) (*Asset, error)

	GetAccount(programHash Uint160) (*account.AccountState, error)

	GetCurrentBlockHash() Uint256
	GetCurrentHeaderHash() Uint256
	GetHeaderHeight() uint32
	GetHeight() uint32
	GetHeaderHashByHeight(height uint32) Uint256

	GetBookKeeperList() ([]*crypto.PubKey, []*crypto.PubKey, error)
	InitLedgerStoreWithGenesisBlock(genesisblock *Block, defaultBookKeeper []*crypto.PubKey) (uint32, error)

	GetQuantityIssued(assetid Uint256) (Fixed64, error)

	GetUnspent(txid Uint256, index uint16) (*utxo.TxOutput, error)
	CcntmainsUnspent(txid Uint256, index uint16) (bool, error)
	GetUnspentFromProgramHash(programHash Uint160, assetid Uint256) ([]*utxo.UTXOUnspent, error)
	GetUnspentsFromProgramHash(programHash Uint160) (map[Uint256][]*utxo.UTXOUnspent, error)
	GetAssets() map[Uint256]*Asset

	IsTxHashDuplicate(txhash Uint256) bool
	IsBlockInStore(hash Uint256) bool
	Close()
}
