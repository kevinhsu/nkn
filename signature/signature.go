package signature

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/nknorg/nkn/common"
	"github.com/nknorg/nkn/crypto"
	"github.com/nknorg/nkn/pb"
	"github.com/nknorg/nkn/vm/interfaces"
)

//SignableData describe the data need be signed.
type SignableData interface {
	interfaces.ICodeContainer
	GetProgramHashes() ([]common.Uint160, error)
	SetPrograms([]*pb.Program)
	GetPrograms() []*pb.Program
	SerializeUnsigned(io.Writer) error
}

func SignBySigner(data SignableData, signer Signer) ([]byte, error) {
	rtx, err := Sign(data, signer.PrivKey())
	if err != nil {
		return nil, fmt.Errorf("[Signature],SignBySigner failed:%v", err)
	}
	return rtx, nil
}

func GetHashData(data SignableData) []byte {
	b_buf := new(bytes.Buffer)
	data.SerializeUnsigned(b_buf)
	return b_buf.Bytes()
}

func GetHashForSigning(data SignableData) []byte {
	temp := sha256.Sum256(GetHashData(data))
	return temp[:]
}

func Sign(data SignableData, prikey []byte) ([]byte, error) {
	signature, err := crypto.Sign(prikey, GetHashData(data))
	if err != nil {
		return nil, fmt.Errorf("[Signature],Sign failed:%v", err)
	}
	return signature, nil
}
