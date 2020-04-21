package registry

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/goshimmer/packages/database"
	"github.com/iotaledger/goshimmer/plugins/qnode/db"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto"
	"github.com/pkg/errors"
	"go.dedis.ch/kyber/v3"
)

func dbKey(addr *HashValue) []byte {
	var buf bytes.Buffer
	buf.WriteString("dkshare")
	buf.Write(addr.Bytes())
	return buf.Bytes()
}

func CommitDKShare(ks *tcrypto.DKShare, pubKeys []kyber.Point) error {
	if err := ks.FinalizeDKS(pubKeys); err != nil {
		return err
	}
	return SaveDKShareToRegistry(ks)
}

func SaveDKShareToRegistry(ks *tcrypto.DKShare) error {
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

func LoadDKShare(address *HashValue, maskPrivate bool) (*tcrypto.DKShare, error) {
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

func unmarshalDKShare(data []byte, maskPrivate bool) (*tcrypto.DKShare, error) {
	ret := &tcrypto.DKShare{}

	err := ret.Read(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	ret.Aggregated = true
	ret.Committed = true
	ret.PubPoly, err = tcrypto.RecoverPubPoly(ret.Suite, ret.PubKeys, ret.T, ret.N)
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
