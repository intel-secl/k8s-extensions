/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package algorithm

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"intel/isecl/k8s-extended-scheduler/v3/constants"
	v1 "k8s.io/api/core/v1"
	"strings"
)

type JwtHeader struct {
	KeyId     string `json:"kid,omitempty"`
	Type      string `json:"typ,omitempty"`
	Algorithm string `json:"alg,omitempty"`
}

//ParseRSAPublicKeyFromPEM is used for parsing and verify public key
func ParseRSAPublicKeyFromPEM(pubKey []byte) (*rsa.PublicKey, error) {
	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKey)
	if err != nil {
		defaultLog.Errorf("error in ParseRSAPublicKeyFromPEM")
		return nil, err
	}
	return verifyKey, err
}

//ValidateAnnotationByPublicKey is used for validate the annotation(cipher) by public key
func ValidateAnnotationByPublicKey(cipherText string, iHubPubKey []byte) error {
	parts := strings.Split(cipherText, ".")
	if len(parts) != 3 {
		return errors.New("Invalid token received, token must have 3 parts")
	}

	jwtHeaderStr := parts[0]
	if l := len(parts[0]) % 4; l > 0 {
		jwtHeaderStr += strings.Repeat("=", 4-l)
	}

	jwtHeaderRcvd, err := base64.URLEncoding.DecodeString(jwtHeaderStr)
	if err != nil {
		return errors.Wrap(err, "Failed to decode jwt header")
	}
	var jwtHeader JwtHeader
	err = json.Unmarshal(jwtHeaderRcvd, &jwtHeader)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal jwt header")
	}

	//Validate the keyid in jwt header
	block, _ := pem.Decode(iHubPubKey)
	if block == nil || block.Type != "PUBLIC KEY" {
		return errors.New("Failed to decode PEM block containing public key")
	}

	keyIdBytes := sha1.Sum(block.Bytes)
	keyIdStr := base64.StdEncoding.EncodeToString(keyIdBytes[:])

	var key *rsa.PublicKey
	if jwtHeader.KeyId == keyIdStr {
		key, err = ParseRSAPublicKeyFromPEM(iHubPubKey)
		if err != nil {
			return errors.New("Invalid IHub public key")
		}
	}

	signatureString, err := base64.URLEncoding.DecodeString(parts[2])
	if err != nil {
		return errors.Wrap(err, "Error while base64 decoding of signature")
	}

	h := sha512.New384()
	h.Write([]byte(parts[0] + "." + parts[1]))
	return rsa.VerifyPKCS1v15(key, crypto.SHA384, h.Sum(nil), signatureString)
}

//JWTParseWithClaims uses ParseUnverified from dgrijalva/jwt-go for parsing and adding the annotation values in claims map
//ParseUnverified doesnt do signature validation. But however the signature validation is being done at ValidateAnnotationByPublicKey
func JWTParseWithClaims(cipherText string, claim jwt.MapClaims) bool {
	_, _, err := new(jwt.Parser).ParseUnverified(cipherText, claim)
	if err != nil {
		defaultLog.Errorf("Error while parsing the annotation %v", err)
		return false
	}
	return true
}

//CheckAnnotationAttrib is used to validate node with respect to time,trusted and location tags
func CheckAnnotationAttrib(cipherText string, node []v1.NodeSelectorRequirement, iHubPubKeys map[string][]byte, tagPrefix, attestationType string) bool {

	var validationStatus error
	validationStatus = ValidateAnnotationByPublicKey(cipherText, iHubPubKeys[attestationType])
	if validationStatus == nil {
		defaultLog.Info("Signature is valid, trust report is from valid Integration Hub")
	} else {
		defaultLog.Errorf("Signature validation failed with error %v", validationStatus)
		return false
	}

	var claims = jwt.MapClaims{}
	//cipherText is the annotation applied to the node, claims is the parsed AH report assigned as the annotation
	jwtParseStatus := JWTParseWithClaims(cipherText, claims)
	if !jwtParseStatus {
		return false
	}

	nodeValidated := false
	if attestationType == constants.HVSAttestation {
		nodeValidated = ValidatePodWithHvsAnnotation(node, claims, tagPrefix)
	} else if attestationType == constants.SGXAttestation {
		nodeValidated = ValidatePodWithSgxAnnotation(node, claims, tagPrefix)
	}

	return nodeValidated
}
