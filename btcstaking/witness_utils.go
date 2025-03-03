package btcstaking

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/wire"
)

func (si *SpendInfo) CreateTimeLockPathWitness(delegatorSig *schnorr.Signature) (wire.TxWitness, error) {
	if si == nil {
		panic("cannot build witness without spend info")
	}
	if delegatorSig == nil {
		return nil, fmt.Errorf("delegator signature should not be nil")
	}
	return CreateWitness(si, [][]byte{delegatorSig.Serialize()})
}

// CreateUnbondingPathWitness helper function to create a witness to spend
// transaction through the unbonding path.
// It is up to the caller to ensure that the amount of covenantSigs matches the
// expected quorum of covenenant members and the transaction has unbonding path.
func (si *SpendInfo) CreateUnbondingPathWitness(
	covenantSigs []*schnorr.Signature,
	delegatorSig *schnorr.Signature,
) (wire.TxWitness, error) {
	if si == nil {
		panic("cannot build witness without spend info")
	}

	var witnessStack [][]byte

	// add covenant signatures to witness stack
	// NOTE: only a quorum number of covenant signatures needs to be non-nil
	if len(covenantSigs) == 0 {
		return nil, fmt.Errorf("covenant signatures should not be empty")
	}
	for _, covSig := range covenantSigs {
		if covSig == nil {
			witnessStack = append(witnessStack, []byte{})
		} else {
			witnessStack = append(witnessStack, covSig.Serialize())
		}
	}

	// add delegator signature to witness stack
	if delegatorSig == nil {
		return nil, fmt.Errorf("delegator signature should not be nil")
	}
	witnessStack = append(witnessStack, delegatorSig.Serialize())

	return CreateWitness(si, witnessStack)
}

// CreateSlashingPathWitness helper function to create a witness to spend
// transaction through the slashing path.
// It is up to the caller to ensure that the amount of covenantSigs matches the
// expected quorum of covenenant members, the finality provider sigs respect the finality providers
// that the delegation belongs to, and the transaction has slashing path.
func (si *SpendInfo) CreateSlashingPathWitness(
	covenantSigs []*schnorr.Signature,
	fpSigs []*schnorr.Signature,
	delegatorSig *schnorr.Signature,
) (wire.TxWitness, error) {
	if si == nil {
		panic("cannot build witness without spend info")
	}

	var witnessStack [][]byte

	// add covenant signatures to witness stack
	// NOTE: only a quorum number of covenant signatures needs to be non-nil
	if len(covenantSigs) == 0 {
		return nil, fmt.Errorf("covenant signatures should not be empty")
	}
	for _, covSig := range covenantSigs {
		if covSig == nil {
			witnessStack = append(witnessStack, []byte{})
		} else {
			witnessStack = append(witnessStack, covSig.Serialize())
		}
	}

	// add finality provider signatures to witness stack
	// NOTE: only 1 of the finality provider signatures needs to be non-nil
	if len(fpSigs) == 0 {
		return nil, fmt.Errorf("finality provider signatures should not be empty")
	}
	for _, fpSig := range fpSigs {
		if fpSig == nil {
			witnessStack = append(witnessStack, []byte{})
		} else {
			witnessStack = append(witnessStack, fpSig.Serialize())
		}
	}

	// add delegator signature to witness stack
	if delegatorSig == nil {
		return nil, fmt.Errorf("delegator signature should not be nil")
	}
	witnessStack = append(witnessStack, delegatorSig.Serialize())

	return CreateWitness(si, witnessStack)
}

// createWitness creates witness for spending the tx corresponding to
// the given spend info with the given stack of signatures
// The returned witness stack follows the structure below:
// - first come signatures
// - then whole revealed script
// - then control block
func CreateWitness(si *SpendInfo, signatures [][]byte) (wire.TxWitness, error) {
	numSignatures := len(signatures)

	controlBlockBytes, err := si.ControlBlock.ToBytes()
	if err != nil {
		return nil, err
	}

	// witness stack has:
	// all signatures
	// whole revealed script
	// control block
	witnessStack := wire.TxWitness(make([][]byte, numSignatures+2))

	for i, sig := range signatures {
		sc := sig
		witnessStack[i] = sc
	}

	witnessStack[numSignatures] = si.GetPkScriptPath()
	witnessStack[numSignatures+1] = controlBlockBytes

	return witnessStack, nil
}
