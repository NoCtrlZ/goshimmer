package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/dkgapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto/tbdn"
	"github.com/pkg/errors"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/bdn"
	"net/http"
	"os"
	"time"
)

var apiPorts = []int{8080, 8081, 8082, 8083}

const (
	aidString  = "6fa90bddeb44a93128726531f4f775154ac865ddc0995d14703f56efba2b8962"
	dksName    = "distributed key set 3"
	N          = 4
	T          = 3
	dataToSign = "Hello, world"
)

func main() {
	timeStart := time.Now()
	if N != len(apiPorts) {
		panic("wrong params")
	}
	aid, err := hashing.HashValueFromString(aidString)
	if err != nil {
		panic(err)
	}
	dksId := hashing.HashStrings(dksName)
	digest := hashing.HashStrings(dataToSign)

	fmt.Printf("checking cosistency of distributed key set at nodes (ports) %+v\n", apiPorts)
	fmt.Printf("assembly id = %s\ndistributed set name = '%s'\nid = %s\n", aid, dksName, dksId.String())
	fmt.Println("---------------------------------------")

	// get public keys
	reqPubs := &dkgapi.GetPubsRequest{
		AssemblyId: aid,
		Id:         dksId,
	}
	numerr := 0
	pkResps := make([]*dkgapi.GetPubsResponse, len(apiPorts))
	for i, port := range apiPorts {
		pkResps[i], err = getPubs(reqPubs, port)
		fmt.Printf("GetPubs (%d, %d). err = %v, resp = %+v\n", i, port, err, pkResps[i])
		if err != nil {
			numerr++
		}
	}
	if numerr != 0 {
		fmt.Printf("exitting due to errors...")
		os.Exit(1)
	}
	// pub keys must all be equal
	rp0 := pkResps[0]
	for _, rp := range pkResps {
		if rp0.PubKeyMaster != rp.PubKeyMaster {
			panic("not equal master public keys from two nodes")
		}
		for i := range rp0.PubKeys {
			if rp0.PubKeys[i] != rp.PubKeys[i] {
				panic("not equal public key sets from two nodes")
			}
		}
	}
	fmt.Printf("all public keys received form nodes are equal!")

	// sign and check signatures
	suite := bn256.NewSuite()
	pubKeys, err := decodePubKeys(suite, rp0.PubKeys)
	pubShares := make([]*share.PubShare, len(pubKeys))
	for i, v := range pubKeys {
		pubShares[i] = &share.PubShare{
			I: i,
			V: v,
		}
	}
	pubPoly, err := share.RecoverPubPoly(suite.G2(), pubShares, T, N)
	if err != nil {
		panic(err)
	}
	reqSign := &dkgapi.SignDigestRequest{
		AssemblyId: aid,
		Id:         dksId,
		DataDigest: digest,
	}
	// sign and check each
	signB := make([][]byte, len(apiPorts))

	for i, port := range apiPorts {
		resp, err := signDigest(reqSign, port)
		fmt.Printf("Sign (%d, %d). err = %v, resp = %+v\n", i, port, err, resp)
		if err != nil {
			continue
		}
		signB[i], err = hex.DecodeString(resp.SigShare)
		if err != nil {
			panic(err)
		}
		err = tbdn.Verify(suite, pubPoly, digest.Bytes(), signB[i])
		if err != nil {
			fmt.Printf("sigShare verification #%d: %v\n", i, err)
		} else {
			fmt.Printf("sigShare verification #%d: ok\n", i)
		}
	}
	res, err := checkCombi(suite, []int{0, 1, 2, 3}, signB, pubPoly, digest.Bytes())
	expectResult(err, res, true, []int{0, 1, 2, 3})

	res, err = checkCombi(suite, []int{0, 1, 2}, signB, pubPoly, digest.Bytes())
	expectResult(err, res, true, []int{0, 1, 2})

	res, err = checkCombi(suite, []int{1, 2, 3}, signB, pubPoly, digest.Bytes())
	expectResult(err, res, true, []int{1, 2, 3})

	res, err = checkCombi(suite, []int{1, 2}, signB, pubPoly, digest.Bytes())
	expectResult(err, res, false, []int{1, 2})

	fmt.Printf("Total duration: %v\n", time.Since(timeStart))
}

func expectResult(err error, res bool, expect bool, indices []int) {
	if err != nil {
		fmt.Printf("signature aggregation and verification for %+v: error %v\n", indices, err)
		return
	}
	if res == expect {
		fmt.Printf("signature aggregation and verification for %+v: PASSED\n", indices)
	} else {
		fmt.Printf("signature aggregation and verification for %+v: FAILED\n", indices)
	}
}

func checkCombi(suite *bn256.Suite, indices []int, allSigs [][]byte, pubPoly *share.PubPoly, msg []byte) (bool, error) {
	fmt.Printf("checkCombi %+v", indices)
	ss, err := makeSubset(indices, allSigs)
	if err != nil {
		fmt.Printf("checkCombi: %v\n", err)
		return false, err
	}
	sig, err := tbdn.Recover(suite, pubPoly, msg, ss, T, N)
	if err != nil {
		fmt.Printf("sigShare Recover %v\n", err)
	} else {
		fmt.Printf("Recovered signature %s\n", hex.EncodeToString(sig))
	}

	pubKey := pubPoly.Commit()
	pkb, _ := pubKey.MarshalBinary()
	fmt.Printf("++++++++++++++++++ Public key length: %d bytes, signature length: %d bytes\n", len(pkb), len(sig))

	err = bdn.Verify(suite, pubKey, msg, sig)
	if err != nil {
		fmt.Printf("Recovered signature verification: %v\n", err)
	} else {
		fmt.Printf("Recovered signature verification: ok\n")
	}
	return err == nil, nil
}

func makeSubset(indices []int, allSigs [][]byte) ([][]byte, error) {
	ret := make([][]byte, 0)
	for _, idx := range indices {
		if idx < 0 || idx >= len(allSigs) {
			return nil, errors.New("wrong index")
		}
		ret = append(ret, allSigs[idx])
	}
	return ret, nil
}

func decodePubKeys(suite *bn256.Suite, pubKeysS []string) ([]kyber.Point, error) {
	ret := make([]kyber.Point, len(pubKeysS))
	for i, s := range pubKeysS {
		b, err := hex.DecodeString(s)
		if err != nil {
			return nil, err
		}
		p := suite.G2().Point()
		if err := p.UnmarshalBinary(b); err != nil {
			return nil, err
		}
		ret[i] = p
	}
	return ret, nil
}

func getPubs(req *dkgapi.GetPubsRequest, port int) (*dkgapi.GetPubsResponse, error) {
	url := fmt.Sprintf("http://localhost:%d/adm/getpubs", port)
	dat, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(dat))
	if err != nil {
		return nil, err
	}

	var ret dkgapi.GetPubsResponse

	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, err
	}
	if ret.Err != "" {
		return nil, errors.New(ret.Err)
	}
	return &ret, nil
}

func signDigest(req *dkgapi.SignDigestRequest, port int) (*dkgapi.SignDigestResponse, error) {
	url := fmt.Sprintf("http://localhost:%d/adm/signdigest", port)
	dat, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(dat))
	if err != nil {
		return nil, err
	}

	var ret dkgapi.SignDigestResponse

	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, err
	}
	if ret.Err != "" {
		return nil, errors.New(ret.Err)
	}
	return &ret, nil
}
