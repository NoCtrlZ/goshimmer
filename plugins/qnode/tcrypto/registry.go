package tcrypto

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/packages/database"
	"github.com/iotaledger/goshimmer/plugins/qnode/db"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/pkg/errors"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/share"
)

func dbKey(addr *HashValue) []byte {
	var buf bytes.Buffer
	buf.WriteString("dkshare")
	buf.Write(addr.Bytes())
	return buf.Bytes()
}

func (ks *DKShare) SaveToRegistry() error {
	if !ks.Committed {
		return fmt.Errorf("uncommited DK share: can't be saved to the registry")
	}
	dbase, err := db.Get()
	if err != nil {
		return err
	}
	dbkey := dbKey(ks.Address)
	exists, err := dbase.Contains(dbkey)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("attempt to overwrite existing DK key share")
	}

	var buf bytes.Buffer

	err = ks.Write(&buf)
	if err != nil {
		return err
	}
	return dbase.Set(database.Entry{
		Key:   dbkey,
		Value: buf.Bytes(),
	})
}

func LoadDKShare(address *HashValue, maskPrivate bool) (*DKShare, error) {
	dbase, err := db.Get()
	if err != nil {
		return nil, err
	}
	dbkey := dbKey(address)
	entry, err := dbase.Get(dbkey)
	if err != nil {
		return nil, err
	}
	ret, err := unmarshalDKShare(entry.Value, maskPrivate)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func ExistDKShareInRegistry(addr *HashValue) (bool, error) {
	dbase, err := db.Get()
	if err != nil {
		return false, err
	}
	dbkey := dbKey(addr)
	return dbase.Contains(dbkey)
}

func unmarshalDKShare(data []byte, maskPrivate bool) (*DKShare, error) {
	ret := &DKShare{}

	err := ret.Read(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	ret.Aggregated = true
	ret.Committed = true
	ret.PubPoly, err = recoverPubPoly(ret.Suite, ret.PubKeys, ret.T, ret.N)
	if err != nil {
		return nil, err
	}
	pubKeyOwn := ret.Suite.G2().Point().Mul(ret.PriKey, nil)
	if !pubKeyOwn.Equal(ret.PubKeys[ret.Index]) {
		return nil, errors.New("crosscheck I: inconsistency while calculating public key")
	}
	ret.PubKeyOwn = ret.PubKeys[ret.Index]
	ret.PubKeyMaster = ret.PubPoly.Commit()
	if maskPrivate {
		ret.PriKey = nil
	}
	binPK, err := ret.PubKeyMaster.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if !HashData(binPK).Equal(ret.Address) {
		return nil, errors.New("crosscheck II: !HashData(binPK).Equal(ret.Address)")
	}
	return ret, nil
}

func recoverPubPoly(suite *bn256.Suite, pubKeys []kyber.Point, t, n uint16) (*share.PubPoly, error) {
	pubShares := make([]*share.PubShare, len(pubKeys))
	for i, v := range pubKeys {
		pubShares[i] = &share.PubShare{
			I: i,
			V: v,
		}
	}
	return share.RecoverPubPoly(suite.G2(), pubShares, int(t), int(n))
}
