package account

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	. "github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/common/password"
	"github.com/Ontology/core/ccntmract"
	ct "github.com/Ontology/core/ccntmract"
	"github.com/Ontology/core/ledger"
	sig "github.com/Ontology/core/signature"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/errors"
	"github.com/Ontology/net/protocol"
)

const (
	DefaultBookKeeperCount = 4
	WalletFileName         = "wallet.dat"
)

type Client interface {
	Sign(ccntmext *ct.CcntmractCcntmext) bool
	CcntmainsAccount(pubKey *crypto.PubKey) bool
	GetAccount(pubKey *crypto.PubKey) (*Account, error)
	GetDefaultAccount() (*Account, error)
	GetBookKeepers() ([]*crypto.PubKey, error)
}

type ClientImpl struct {
	mu sync.Mutex

	path      string
	iv        []byte
	masterKey []byte

	accounts  map[Uint160]*Account
	ccntmracts map[Uint160]*ct.Ccntmract

	watchOnly     []Uint160
	currentHeight uint32

	FileStore
	isrunning bool
}

//TODO: adjust ccntmract folder structure
func Create(path string, passwordKey []byte) *ClientImpl {
	cl := NewClient(path, passwordKey, true)

	_, err := cl.CreateAccount()
	if err != nil {
		fmt.Println(err)
	}

	return cl
}

func Open(path string, passwordKey []byte) *ClientImpl {
	cl := NewClient(path, passwordKey, false)
	if cl == nil {
		log.Error("Alloc new client failure")
		return nil
	}

	cl.accounts = cl.LoadAccount()
	if cl.accounts == nil {
		log.Error("Load accounts failure")
	}
	cl.ccntmracts = cl.LoadCcntmracts()
	if cl.ccntmracts == nil {
		log.Error("Load ccntmracts failure")
	}
	return cl
}

func NewClient(path string, password []byte, create bool) *ClientImpl {
	newClient := &ClientImpl{
		path:      path,
		accounts:  map[Uint160]*Account{},
		ccntmracts: map[Uint160]*ct.Ccntmract{},
		FileStore: FileStore{path: path},
		isrunning: true,
	}

	passwordKey := crypto.ToAesKey(password)
	if create {
		//create new client
		newClient.iv = make([]byte, 16)
		newClient.masterKey = make([]byte, 32)
		newClient.watchOnly = []Uint160{}
		newClient.currentHeight = 0

		//generate random number for iv/masterkey
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < 16; i++ {
			newClient.iv[i] = byte(r.Intn(256))
		}
		for i := 0; i < 32; i++ {
			newClient.masterKey[i] = byte(r.Intn(256))
		}

		//new client store (build DB)
		newClient.BuildDatabase(path)

		// SaveStoredData
		pwdhash := sha256.Sum256(passwordKey)
		err := newClient.SaveStoredData("PasswordHash", pwdhash[:])
		if err != nil {
			log.Error(err)
			return nil
		}
		err = newClient.SaveStoredData("IV", newClient.iv[:])
		if err != nil {
			log.Error(err)
			return nil
		}

		aesmk, err := crypto.AesEncrypt(newClient.masterKey[:], passwordKey, newClient.iv)
		if err == nil {
			err = newClient.SaveStoredData("MasterKey", aesmk)
			if err != nil {
				log.Error(err)
				return nil
			}
		} else {
			log.Error(err)
			return nil
		}
	} else {
		if b := newClient.verifyPasswordKey(passwordKey); b == false {
			return nil
		}
		if err := newClient.loadClient(passwordKey); err != nil {
			return nil
		}
	}
	ClearBytes(passwordKey, len(passwordKey))
	return newClient
}

func (cl *ClientImpl) loadClient(passwordKey []byte) error {
	var err error
	cl.iv, err = cl.LoadStoredData("IV")
	if err != nil {
		fmt.Println("error: failed to load iv")
		return err
	}
	encryptedMasterKey, err := cl.LoadStoredData("MasterKey")
	if err != nil {
		fmt.Println("error: failed to load master key")
		return err
	}
	cl.masterKey, err = crypto.AesDecrypt(encryptedMasterKey, passwordKey, cl.iv)
	if err != nil {
		fmt.Println("error: failed to decrypt master key")
		return err
	}
	return nil
}

func (cl *ClientImpl) GetDefaultAccount() (*Account, error) {
	for programHash, _ := range cl.accounts {
		return cl.GetAccountByProgramHash(programHash), nil
	}

	return nil, NewDetailErr(errors.New("Can't load default account."), ErrNoCode, "")
}

func (cl *ClientImpl) GetAccount(pubKey *crypto.PubKey) (*Account, error) {
	signatureRedeemScript, err := ccntmract.CreateSignatureRedeemScript(pubKey)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "CreateSignatureRedeemScript failed")
	}
	programHash := ToCodeHash(signatureRedeemScript)
	return cl.GetAccountByProgramHash(programHash), nil
}

func (cl *ClientImpl) GetAccountByProgramHash(programHash Uint160) *Account {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if account, ok := cl.accounts[programHash]; ok {
		return account
	}
	return nil
}

func (cl *ClientImpl) GetCcntmract(programHash Uint160) *ct.Ccntmract {
	log.Debug()
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if ccntmract, ok := cl.ccntmracts[programHash]; ok {
		return ccntmract
	}
	return nil
}

func (cl *ClientImpl) ChangePassword(oldPassword []byte, newPassword []byte) bool {
	// check password
	oldPasswordKey := crypto.ToAesKey(oldPassword)
	if !cl.verifyPasswordKey(oldPasswordKey) {
		fmt.Println("error: password verification failed")
		return false
	}
	if err := cl.loadClient(oldPasswordKey); err != nil {
		fmt.Println("error: load wallet info failed")
		return false
	}

	// encrypt master key with new password
	newPasswordKey := crypto.ToAesKey(newPassword)
	newMasterKey, err := crypto.AesEncrypt(cl.masterKey, newPasswordKey, cl.iv)
	if err != nil {
		fmt.Println("error: set new password failed")
		return false
	}

	// update wallet file
	newPasswordHash := sha256.Sum256(newPasswordKey)
	if err := cl.SaveStoredData("PasswordHash", newPasswordHash[:]); err != nil {
		fmt.Println("error: wallet update failed(password hash)")
		return false
	}
	if err := cl.SaveStoredData("MasterKey", newMasterKey); err != nil {
		fmt.Println("error: wallet update failed (encrypted master key)")
		return false
	}
	ClearBytes(newPasswordKey, len(newPasswordKey))
	ClearBytes(cl.masterKey, len(cl.masterKey))

	return true
}

func (cl *ClientImpl) CcntmainsAccount(pubKey *crypto.PubKey) bool {
	signatureRedeemScript, err := ccntmract.CreateSignatureRedeemScript(pubKey)
	if err != nil {
		return false
	}
	programHash := ToCodeHash(signatureRedeemScript)
	if cl.GetAccountByProgramHash(programHash) != nil {
		return true
	} else {
		return false
	}
}

func (cl *ClientImpl) CreateAccount() (*Account, error) {
	ac, err := NewAccount()
	if err != nil {
		return nil, err
	}

	cl.mu.Lock()
	cl.accounts[ac.ProgramHash] = ac
	cl.mu.Unlock()

	err = cl.SaveAccount(ac)
	if err != nil {
		return nil, err
	}

	ct, err := ccntmract.CreateSignatureCcntmract(ac.PublicKey)
	if err == nil {
		cl.AddCcntmract(ct)
		address, _ := ct.ProgramHash.ToAddress()
		log.Info("[CreateCcntmract] Address: ", address)
	}

	log.Info("Create account Success")
	return ac, nil
}

func (cl *ClientImpl) CreateAccountByPrivateKey(privateKey []byte) (*Account, error) {
	ac, err := NewAccountWithPrivatekey(privateKey)
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if err != nil {
		return nil, err
	}

	cl.accounts[ac.ProgramHash] = ac
	err = cl.SaveAccount(ac)
	if err != nil {
		return nil, err
	}
	return ac, nil
}

func (cl *ClientImpl) ProcessBlocks() {
	for {
		if !cl.isrunning {
			break
		}

		for {
			if cl.currentHeight > ledger.DefLedger.GetCurrentBlockHeight(){
				break
			}

			cl.mu.Lock()

			block, _ := ledger.DefLedger.GetBlockByHeight(cl.currentHeight)
			if block != nil {
				cl.ProcessNewBlock(block)
			}

			cl.mu.Unlock()
		}

		for i := 0; i < 20; i++ {
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (cl *ClientImpl) ProcessNewBlock(block *types.Block) {
	//TODO: ProcessNewBlock

}

func (cl *ClientImpl) Sign(ccntmext *ct.CcntmractCcntmext) bool {
	log.Debug()
	fSuccess := false
	for i, hash := range ccntmext.ProgramHashes {
		ccntmract := cl.GetCcntmract(hash)
		if ccntmract == nil {
			ccntminue
		}
		account := cl.GetAccountByProgramHash(hash)
		if account == nil {
			ccntminue
		}

		signature, err := sig.SignBySigner(ccntmext.Data, account)
		if err != nil {
			return fSuccess
		}
		err = ccntmext.AddCcntmract(ccntmract, account.PublicKey, signature)

		if err != nil {
			fSuccess = false
		} else {
			if i == 0 {
				fSuccess = true
			}
		}
	}
	return fSuccess
}

func (cl *ClientImpl) verifyPasswordKey(passwordKey []byte) bool {
	savedPasswordHash, err := cl.LoadStoredData("PasswordHash")
	if err != nil {
		fmt.Println("error: failed to load password hash")
		return false
	}
	if savedPasswordHash == nil {
		fmt.Println("error: saved password hash is nil")
		return false
	}
	passwordHash := sha256.Sum256(passwordKey)
	///ClearBytes(passwordKey, len(passwordKey))
	if !IsEqualBytes(savedPasswordHash, passwordHash[:]) {
		fmt.Println("error: password wrcntm")
		return false
	}
	return true
}

func (cl *ClientImpl) EncryptPrivateKey(prikey []byte) ([]byte, error) {
	enc, err := crypto.AesEncrypt(prikey, cl.masterKey, cl.iv)
	if err != nil {
		return nil, err
	}

	return enc, nil
}

func (cl *ClientImpl) DecryptPrivateKey(prikey []byte) ([]byte, error) {
	if prikey == nil {
		return nil, NewDetailErr(errors.New("The PriKey is nil"), ErrNoCode, "")
	}
	if len(prikey) != 96 {
		return nil, NewDetailErr(errors.New("The len of PriKeyEnc is not 96bytes"), ErrNoCode, "")
	}

	dec, err := crypto.AesDecrypt(prikey, cl.masterKey, cl.iv)
	if err != nil {
		return nil, err
	}

	return dec, nil
}

func (cl *ClientImpl) SaveAccount(ac *Account) error {
	decryptedPrivateKey := make([]byte, 96)
	temp, err := ac.PublicKey.EncodePoint(false)
	if err != nil {
		return err
	}
	for i := 1; i <= 64; i++ {
		decryptedPrivateKey[i-1] = temp[i]
	}

	for i := len(ac.PrivateKey) - 1; i >= 0; i-- {
		decryptedPrivateKey[96+i-len(ac.PrivateKey)] = ac.PrivateKey[i]
	}

	encryptedPrivateKey, err := cl.EncryptPrivateKey(decryptedPrivateKey)
	if err != nil {
		return err
	}

	ClearBytes(decryptedPrivateKey, 96)

	err = cl.SaveAccountData(ac.ProgramHash.ToArray(), encryptedPrivateKey)
	if err != nil {
		return err
	}

	return nil
}

func (cl *ClientImpl) LoadAccount() map[Uint160]*Account {
	i := 0
	accounts := map[Uint160]*Account{}
	for true {
		_, prikeyenc, err := cl.LoadAccountData(i)
		if err != nil {
			// TODO: report the error
		}

		decryptedPrivateKey, err := cl.DecryptPrivateKey(prikeyenc)
		if err != nil {
			log.Error(err)
		}

		prikey := decryptedPrivateKey[64:96]
		ac, err := NewAccountWithPrivatekey(prikey)
		accounts[ac.ProgramHash] = ac
		i++
		break
	}

	return accounts
}

func (cl *ClientImpl) LoadCcntmracts() map[Uint160]*ct.Ccntmract {
	i := 0
	ccntmracts := map[Uint160]*ct.Ccntmract{}

	for true {
		ph, _, rd, err := cl.LoadCcntmractData(i)
		if err != nil {
			fmt.Println(err)
			break
		}

		rdreader := bytes.NewReader(rd)
		ct := new(ct.Ccntmract)
		ct.Deserialize(rdreader)

		programhash, err := Uint160ParseFromBytes(ph)
		ct.ProgramHash = programhash
		ccntmracts[ct.ProgramHash] = ct
		i++
		break
	}

	return ccntmracts
}

func (cl *ClientImpl) AddCcntmract(ct *ccntmract.Ccntmract) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if cl.accounts[ct.ProgramHash] == nil {
		return NewDetailErr(errors.New("AddCcntmract(): ccntmract.OwnerPubkeyHash not in []accounts"), ErrNoCode, "")
	}

	cl.ccntmracts[ct.ProgramHash] = ct

	err := cl.SaveCcntmractData(ct)
	return err
}

func (cl *ClientImpl) GetBookKeepers() ([]*crypto.PubKey, error) {
	var pubKeys = []*crypto.PubKey{}
	consensusType := config.Parameters.ConsensusType
	if consensusType == "solo" {
		ac, err := cl.GetDefaultAccount()
		if err != nil {
			return nil, fmt.Errorf("GetDefaultAccount error:%s", err)
		}
		pubKeys = append(pubKeys, ac.PublicKey)
		return pubKeys, nil
	}

	sort.Strings(config.Parameters.BookKeepers)
	for _, key := range config.Parameters.BookKeepers {
		pubKey := []byte(key)
		pubKey, err := hex.DecodeString(key)
		// TODO Convert the key string to byte
		k, err := crypto.DecodePoint(pubKey)
		if err != nil {
			log.Error("Incorrectly book keepers key")
			return nil, fmt.Errorf("Incorrectly book keepers key:%s")
		}
		pubKeys = append(pubKeys, k)
	}

	return pubKeys, nil
}

func clientIsDefaultBookKeeper(publicKey string) bool {
	for _, bookKeeper := range config.Parameters.BookKeepers {
		if strings.Compare(bookKeeper, publicKey) == 0 {
			return true
		}
	}
	return false
}

func nodeType(typeName string) int {
	if "service" == config.Parameters.NodeType {
		return protocol.SERVICENODE
	} else {
		return protocol.VERIFYNODE
	}
}

func GetClient() Client {
	if !FileExisted(WalletFileName) {
		log.Fatal(fmt.Sprintf("No %s detected, please create a wallet by using command line.", WalletFileName))
		os.Exit(1)
	}
	passwd, err := password.GetAccountPassword()
	if err != nil {
		log.Fatal("Get password error.")
		os.Exit(1)
	}
	c := Open(WalletFileName, passwd)
	if c == nil {
		return nil
	}
	return c
}

func GetBookKeepers() []*crypto.PubKey {
	var pubKeys = []*crypto.PubKey{}
	sort.Strings(config.Parameters.BookKeepers)
	for _, key := range config.Parameters.BookKeepers {
		pubKey := []byte(key)
		pubKey, err := hex.DecodeString(key)
		// TODO Convert the key string to byte
		k, err := crypto.DecodePoint(pubKey)
		if err != nil {
			log.Error("Incorrectly book keepers key")
			return nil
		}
		pubKeys = append(pubKeys, k)
	}

	return pubKeys
}
