package node

import (
	"bytes"
	"fmt"

	"github.com/gogo/protobuf/proto"
	. "github.com/nknorg/nkn/block"
	nknErrors "github.com/nknorg/nkn/errors"
	"github.com/nknorg/nkn/pb"
	. "github.com/nknorg/nkn/transaction"
	"github.com/nknorg/nkn/util/log"
	nnetpb "github.com/nknorg/nnet/protobuf"
)

// NewTransactionsMessage creates a TRANSACTIONS message
func NewTransactionsMessage(transactions []*Transaction) (*pb.UnsignedMessage, error) {
	transactionsBytes := make([][]byte, len(transactions), len(transactions))
	for i, transaction := range transactions {
		b := new(bytes.Buffer)
		err := transaction.Serialize(b)
		if err != nil {
			return nil, err
		}
		transactionsBytes[i] = b.Bytes()
	}

	msgBody := &pb.Transactions{
		Transactions: transactionsBytes,
	}

	buf, err := proto.Marshal(msgBody)
	if err != nil {
		return nil, err
	}

	msg := &pb.UnsignedMessage{
		MessageType: pb.TRANSACTIONS,
		Message:     buf,
	}

	return msg, nil
}

// transactionsMessageHandler handles a TRANSACTIONS message
func (localNode *LocalNode) transactionsMessageHandler(remoteMessage *RemoteMessage) ([]byte, bool, error) {
	msgBody := &pb.Transactions{}
	err := proto.Unmarshal(remoteMessage.Message, msgBody)
	if err != nil {
		return nil, false, err
	}

	if len(msgBody.Transactions) == 0 {
		return nil, false, fmt.Errorf("no transactions in message body")
	}

	for _, txnBytes := range msgBody.Transactions {
		txn := &Transaction{}
		err = txn.Deserialize(bytes.NewReader(txnBytes))
		if err != nil {
			log.Warningf("Deserialize transaction error: %v", err)
			continue
		}

		if localNode.ExistHash(txn.Hash()) {
			continue
		}

		if localNode.GetSyncState() != pb.PersistFinished {
			continue
		}

		errCode := localNode.AppendTxnPool(txn)
		if errCode != nknErrors.ErrNoError {
			log.Warningf("Verify transaction failed with %v when append to txn pool", errCode)
			continue
		}
	}

	return nil, false, nil
}

// BroadcastTransaction broadcast a transaction to the network using
// TRANSACTIONS message
func (localNode *LocalNode) BroadcastTransaction(txn *Transaction) error {
	msg, err := NewTransactionsMessage([]*Transaction{txn})
	if err != nil {
		return err
	}

	buf, err := localNode.SerializeMessage(msg, false)
	if err != nil {
		return err
	}

	if txn.UnsignedTx.Payload.Type == pb.CommitType {
		_, err = localNode.nnet.SendBytesBroadcastAsync(buf, nnetpb.BROADCAST_PUSH)
	} else {
		_, err = localNode.nnet.SendBytesBroadcastAsync(buf, nnetpb.BROADCAST_TREE)
	}
	if err != nil {
		return err
	}

	localNode.ExistHash(txn.Hash())

	return nil
}

func (localNode *LocalNode) cleanupTransactions(v interface{}) {
	if block, ok := v.(*Block); ok {
		localNode.TxnPool.CleanSubmittedTransactions(block.Transactions)
	}
}
