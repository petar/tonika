// Tonika: A distributed social networking platform
// Copyright (C) 2010 Petar Maymounkov <petar@5ttt.org>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.


// *** SRC ***

package sys

import (
	"crypto/rsa"
	"crypto/sha1"
	//"fmt"
	"os"
	"tonika/crypto"
	"tonika/util/tube"
)

type authFunc func(tube.EncodeDecoder) (Id, os.Error)

// This function returns the authentication structure for a given remote,
// based on the accept key they provided.
type AuthLookupFunc func(key *DialKey) AuthRemote

func AuthConnect(local AuthLocal, remote AuthRemote, tube tube.TubedConn) (Id, tube.TubedConn, os.Error) {
	
	// Establish symmetric encryption
	xtube, err := authHello(tube)
	if err != nil {
		return 0, nil, err
	}

	// Send challange and dial key
	ch := GenerateSigChallange()
	m1 := &U_AuthConn_M1{
		DialKey: *remote.GetDialKey().Proto(),
		Challange: make([]byte, len(ch)),
	}
	copy(m1.Challange, ch)
	if err := xtube.Encode(m1); err != nil {
		return 0, nil, err
	}

	// Receive their challange
	m1_remote := &U_AuthAcc_M1{}
	if err := xtube.Decode(m1_remote); err != nil {
		return 0, nil, err
	}

	// Respond to their challange
	sign,err := local.GetSignatureKey().Sign(m1_remote.Challange)
	if err != nil {
		return 0, nil, err
	}
	m2 := &U_AuthConn_M2{sign}
	if err := xtube.Encode(m2); err != nil {
		return 0, nil, err
	}

	// Verify their dialkey (against our accept key) and their challange response
	m2_remote := &U_AuthAcc_M2{}
	if err := xtube.Decode(m2_remote); err != nil {
		return 0, nil, err
	}
	dk,err := UnprotoDialKey(&m2_remote.DialKey)
	if err != nil {
		return 0, nil, err
	}
	if *dk != *remote.GetAcceptKey() {
		return 0, nil, os.NewError("DialKey does not match AcceptKey")
	}
	if err := remote.GetSignatureKey().Verify(ch, m2_remote.Resp); err != nil {
		return 0, nil, err
	}
	
	return remote.GetSignatureKey().Id(), xtube, nil
}

func AuthAccept(local AuthLocal, lookup AuthLookupFunc, tube tube.TubedConn) (Id, tube.TubedConn, os.Error) {
	
	// Establish symmetric encryption
	xtube, err := authHello(tube)
	if err != nil {
		return 0, nil, err
	}

	// Send challange
	ch := GenerateSigChallange()
	m1 := &U_AuthAcc_M1{
		Challange: make([]byte, len(ch)),
	}
	copy(m1.Challange, ch)
	if err := xtube.Encode(m1); err != nil {
		return 0, nil, err
	}

	// Receive dial key and challange
	m1_remote := &U_AuthConn_M1{}
	if err := xtube.Decode(m1_remote); err != nil {
		return 0, nil, err
	}

	// Find credentials with accept key = dial key
	ak,err := UnprotoDialKey(&m1_remote.DialKey)
	if err != nil {
		return 0, nil, err
	}
	rauth := lookup(ak)
	if rauth == nil {
		return 0, nil, os.NewError("No remote auth")
	}
	
	// Send dial key and challange response
	sign,err := local.GetSignatureKey().Sign(m1_remote.Challange)
	if err != nil {
		return 0, nil, err
	}
	m2 := &U_AuthAcc_M2{
		DialKey: *rauth.GetDialKey().Proto(),
		Resp:    sign,
	}
	if err := xtube.Encode(m2); err != nil {
		return 0, nil, err
	}

	// Receive their challange response and verify correct
	m2_remote := &U_AuthConn_M2{}
	if err := xtube.Decode(m2_remote); err != nil {
		return 0, nil, err
	}
	if err := rauth.GetSignatureKey().Verify(ch, m2_remote.Resp); err != nil {
		return 0, nil, err
	}

	return rauth.GetSignatureKey().Id(), xtube, nil
}

// authHello establishes a symmetrically encrypted channel over t
func authHello(t tube.TubedConn) (tube.TubedConn, os.Error) {

	// Make my hello private key
	HelloA := GenerateHelloKey()
	HelloA_proto := HelloA.Proto()

	// Send my hello public key
	err := t.Encode(HelloA_proto)
	if err != nil {
		return nil, err
	}

	// Receive their hello public key
	HelloB_proto := &U_HelloKey{}
	err = t.Decode(HelloB_proto)
	if err != nil {
		return nil, err
	}
	HelloB,err := UnprotoHelloPubKey(HelloB_proto)
	if err != nil {
		return nil, err
	}

	// Make my session key half
	HalvesA := GenerateKeyHalves()

	// Encrypt my session key half with their intro public key
	HalvesA_HelloB, err := crypto.EncryptShortMsg(HelloB.RSAPubKey(), 
		HalvesA.Bytes(), []byte("key-halves"))
	if err != nil {
		return nil, err
	}

	// Send my encrypted session half key
	err = t.Encode(&U_KeyHalves{HalvesA_HelloB})
	if err != nil {
		return nil, err
	}

	// Receive their session key half, encrypted with my intro private key
	HalvesB_HelloA_proto := &U_KeyHalves{}
	err = t.Decode(HalvesB_HelloA_proto)
	if err != nil {
		return nil, err
	}

	// Decrypt their session key half
	HalvesB_bytes, err := crypto.DecryptShortMsg(HelloA.RSAPrivKey(), 
		HalvesB_HelloA_proto.Halves, []byte("key-halves"))
	if err != nil {
		return nil, err
	}
	HalvesB, err := BytesToKeyHalves(HalvesB_bytes)
	if err != nil {
		return nil, err
	}

	// Compute session keys me->them and them->me
	keyAB := makeSessionKey("SK", HalvesA.Bytes(), HalvesB.Bytes(), 
		HelloA.RSAPubKey(), HelloB.RSAPubKey())
	keyBA := makeSessionKey("SK", HalvesB.Bytes(), HalvesA.Bytes(), 
		HelloB.RSAPubKey(), HelloA.RSAPubKey())

	// Create encrypted tube
	return tube.NewRC4Tube(t, keyBA, keyAB), nil
}

func makeSessionKey(label string, ha, hb []byte, ka, kb *rsa.PublicKey) []byte {
	khash := sha1.New()
	khash.Write([]byte(label))
	khash.Write(ka.N.Bytes())
	khash.Write(ha[0 : len(ha)/2])
	khash.Write(kb.N.Bytes())
	khash.Write(hb[len(hb)/2 : len(hb)])
	return khash.Sum()
}
