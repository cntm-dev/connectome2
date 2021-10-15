package types

import (
	"github.com/Ontology/common/serialization"
	"io"
)

type VmType byte

const (
	NativeVM = VmType(0xFF)
	NEOVM    = VmType(0x80)
	WASM     = VmType(0x90)
	// EVM = VmType(0x90)
)

type VmCode struct {
	CodeType VmType
	Code     []byte
}

func (self *VmCode) Serialize(w io.Writer) error {
	w.Write([]byte{byte(self.CodeType)})
	return serialization.WriteVarBytes(w, self.Code)

}

func (self *VmCode) Deserialize(r io.Reader) error {
	var b [1]byte
	r.Read(b[:])
	buf, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	self.CodeType = VmType(b[0])
	self.Code = buf
	return nil
}
