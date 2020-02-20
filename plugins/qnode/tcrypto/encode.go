package tcrypto

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/packages/database"
	"github.com/iotaledger/goshimmer/plugins/qnode/db"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/pkg/errors"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/share"
)

func dbKey(aid, addr *HashValue) []byte {
	var buf bytes.Buffer
	buf.WriteString("dkshare")
	buf.Write(aid.Bytes())
	buf.Write(addr.Bytes())
	return buf.Bytes()
}

func dbGroupPrefix(aid *HashValue) []byte {
	var buf bytes.Buffer
	buf.WriteString("dkshare")
	buf.Write(aid.Bytes())
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
	dbkey := dbKey(ks.AssemblyId, ks.Account)
	exists, err := dbase.Contains(dbkey)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("attempt to overwrite existing DK key share")
	}

	ks.PubKeysEncoded = make([]string, len(ks.PubKeys))
	for i, pk := range ks.PubKeys {
		data, err := pk.MarshalBinary()
		if err != nil {
			return err
		}
		ks.PubKeysEncoded[i] = hex.EncodeToString(data)
	}
	pkb, err := ks.PriKey.MarshalBinary()
	if err != nil {
		return err
	}
	ks.PriKeyEncoded = hex.EncodeToString(pkb)

	jsonData, err := json.Marshal(ks)
	if err != nil {
		return err
	}
	return dbase.Set(database.Entry{
		Key:   dbkey,
		Value: jsonData,
	})
}

func LoadDKShare(assemblyId *HashValue, address *HashValue, maskPrivate bool) (*DKShare, error) {
	dbase, err := db.Get()
	if err != nil {
		return nil, err
	}
	dbkey := dbKey(assemblyId, address)
	entry, err := dbase.Get(dbkey)
	if err != nil {
		return nil, err
	}
	ret, err := unmarshalDKShare(entry.Value, maskPrivate)
	if err != nil {
		return nil, err
	}
	if !ret.AssemblyId.Equal(assemblyId) || !ret.Account.Equal(address) {
		return nil, errors.New("inconsistency in key share registry data")
	}
	return ret, nil
}

func ExistDKShareInRegistry(assemblyId, addr *HashValue) (bool, error) {
	dbase, err := db.Get()
	if err != nil {
		return false, err
	}
	dbkey := dbKey(assemblyId, addr)
	return dbase.Contains(dbkey)
}

func LoadDKShares(aid *HashValue, maskPrivate bool) ([]*DKShare, error) {
	dbase, err := db.Get()
	if err != nil {
		return nil, err
	}
	ret := make([]*DKShare, 0)
	pref := dbGroupPrefix(aid)
	err = dbase.ForEachPrefix(pref, func(entry database.Entry) bool {
		if dks, err := unmarshalDKShare(entry.Value, maskPrivate); err == nil {
			ret = append(ret, dks)
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func unmarshalDKShare(jsonData []byte, maskPrivate bool) (*DKShare, error) {
	ret := &DKShare{}
	err := json.Unmarshal(jsonData, ret)
	if err != nil {
		return nil, err
	}
	ret.Suite = bn256.NewSuite()
	ret.Aggregated = true
	ret.Committed = true

	// decode some fields
	// private key
	pkb, err := hex.DecodeString(ret.PriKeyEncoded)
	if err != nil {
		return nil, err
	}
	ret.PriKey = ret.Suite.G2().Scalar()
	if err = ret.PriKey.UnmarshalBinary(pkb); err != nil {
		return nil, err
	}
	// public keys
	ret.PubKeys = make([]kyber.Point, len(ret.PubKeysEncoded))
	for i, pke := range ret.PubKeysEncoded {
		data, err := hex.DecodeString(pke)
		if err != nil {
			return nil, err
		}
		ret.PubKeys[i] = ret.Suite.G2().Point()
		err = ret.PubKeys[i].UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
	}
	ret.PubPoly, err = recoverPubPoly(ret.Suite, ret.PubKeys, ret.T, ret.N)
	if err != nil {
		return nil, err
	}
	ret.PubKeyOwn = ret.Suite.G2().Point().Mul(ret.PriKey, nil)
	if !ret.PubKeyOwn.Equal(ret.PubKeys[ret.Index]) {
		return nil, errors.New("crosscheck I: inconsistency while calculating public key")
	}
	ret.PubKeyMaster = ret.PubPoly.Commit()
	if maskPrivate {
		ret.PriKey = nil
		ret.PriKeyEncoded = ""
	}
	binPK, err := ret.PubKeyMaster.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if !HashData(binPK).Equal(ret.Account) {
		return nil, errors.New("crosscheck II: !HashData(binPK).Equal(ret.Account)")
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
