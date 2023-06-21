package helper

import (
	"fmt"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"log"
	"strconv"
	"strings"
)

const hardenedOffset = 0x80000000

func GenerateMnemonic() string {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		log.Panic(err)
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		log.Panic(err)
	}
	return mnemonic
}

func DeriveKeys(mnemonic, derivationPath string) (string, string, []byte) {
	seed := bip39.NewSeed(mnemonic, "")
	masterKey, _ := bip32.NewMasterKey(seed)
	derivationIndexes := parseDerivationPath(derivationPath)
	derivedKey := masterKey
	for _, index := range derivationIndexes {
		var err error
		derivedKey, err = derivedKey.NewChildKey(index)
		if err != nil {
			log.Panic(err)
		}
	}
	privateKey := derivedKey.B58Serialize()
	pubKey := derivedKey.PublicKey().B58Serialize()

	return privateKey, pubKey, seed
}

func parseDerivationPath(path string) []uint32 {
	parts := strings.Split(path, "/")[1:]
	result := make([]uint32, len(parts))
	for i, part := range parts {
		if strings.HasSuffix(part, "'") {
			n, err := strconv.Atoi(part[:len(part)-1])
			if err != nil {
				log.Panic(err)
			}
			result[i] = uint32(n) + hardenedOffset
		} else {
			n, err := strconv.Atoi(part)
			if err != nil {
				log.Panic(err)
			}
			result[i] = uint32(n)
		}
	}
	return result
}

func DerivePath(extKey *hdkeychain.ExtendedKey, path string) (*hdkeychain.ExtendedKey, error) {
	segments := strings.Split(path, "/")
	childKey := extKey

	for _, segment := range segments {
		index, err := strconv.ParseUint(segment, 10, 32)
		if err != nil {
			return nil, err
		}

		childKey, err = childKey.Child(uint32(index))
		if err != nil {
			return nil, err
		}
	}

	return childKey, nil
}

func GeneratePriv(extendedKey string, derivePath string) string {
	// ExtendedKey'yi hdkeychain.ExtendedKey nesnesine çevirin
	extKey, err := hdkeychain.NewKeyFromString(extendedKey)
	if err != nil {
		panic(err)
	}

	// DerivePath fonksiyonunu kullanarak çocuk anahtarı elde edin
	childKey, err := DerivePath(extKey, derivePath)
	if err != nil {
		panic(err)
	}

	// Çocuk anahtarını özel anahtar ve genel anahtar çiftine çevirin
	privateKey, err := childKey.ECPrivKey()
	if err != nil {
		panic(err)
	}

	// Özel anahtarı bir ECDSA özel anahtara çevirin
	privateKeyECDSA := privateKey.ToECDSA()

	// ECDSA genel anahtardan Ethereum adresini elde edin
	address := crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)

	// Ethereum adresini hex biçiminde döndürün
	return address.Hex()
}

func GeneratePub(publicKey string, index int32) common.Address {

	extPubKeyStr := publicKey

	extKey, err := hdkeychain.NewKeyFromString(extPubKeyStr)
	if err != nil {
		panic(err)
	}

	path := fmt.Sprintf("0/%d", index)

	childKey, err := DerivePath(extKey, path)
	if err != nil {
		panic(err)
	}

	rawPubKey, err := childKey.ECPubKey()
	if err != nil {
		panic(err)
	}

	ethAddress := crypto.PubkeyToAddress(*rawPubKey.ToECDSA())

	return ethAddress

}
