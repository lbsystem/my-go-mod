package encrypt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
)

type ECDHEManager struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	curve      elliptic.Curve
}

func NewECDHEManager(curveName string) (*ECDHEManager, error) {
	var curve elliptic.Curve
	switch curveName {
	case "P224":
		curve = elliptic.P224()
	case "P256":
		curve = elliptic.P256()
	case "P384":
		curve = elliptic.P384()
	case "P521":
		curve = elliptic.P521()
	default:
		return nil, errors.New("unsupported curve")
	}

	privKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, err
	}
	return &ECDHEManager{
		privateKey: privKey,
		publicKey:  &privKey.PublicKey,
		curve:      curve,
	}, nil
}
//重新设置 加密长度
func (e *ECDHEManager) SetCurve(curve elliptic.Curve) {
	e.curve = curve
}
//返回公钥以便发送给对端
func (e *ECDHEManager) PublicKeyBytes() []byte {
	return elliptic.Marshal(e.curve, e.publicKey.X, e.publicKey.Y)
}
//用对端的公钥 算出相同的密码
func (e *ECDHEManager) ComputeSharedSecret(peerPubKeyBytes []byte) ([]byte, error) {
	x, y := elliptic.Unmarshal(e.curve, peerPubKeyBytes)
	if x == nil || y == nil {
		return nil, fmt.Errorf("failed to unmarshal public key")
	}
	peerPubKey := &ecdsa.PublicKey{Curve: e.curve, X: x, Y: y}

	secretX, _ := e.curve.ScalarMult(peerPubKey.X, peerPubKey.Y, e.privateKey.D.Bytes())
	if secretX == nil {
		return nil, fmt.Errorf("failed to generate shared secret")
	}
	return secretX.Bytes(), nil
}
