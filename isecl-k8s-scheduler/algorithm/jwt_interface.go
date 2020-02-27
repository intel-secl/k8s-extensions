/*
Copyright © 2019 Intel Corporation
SPDX-License-Identifier: BSD-3-Clause
*/

package algorithm

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/sha1"
	//"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/json"
	"errors"
	"strings"
	"k8s_scheduler_cit_extension-k8s_extended_scheduler/util"
	jwt "github.com/dgrijalva/jwt-go"
	"k8s.io/api/core/v1"
)

type JwtHeader struct {
	KeyId                  string `json:"kid,omitempty"`
	Type       	       string `json:"typ,omitempty"`
	Algorithm              string `json:"alg,omitempty"`
}

//ParseRSAPublicKeyFromPEM is used for parsing and verify public key
func ParseRSAPublicKeyFromPEM(pubKey []byte) (*rsa.PublicKey, error) {
	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKey)
	if err != nil {
		Log.Errorf("error in ParseRSAPublicKeyFromPEM")
		return nil,err
	}
	return verifyKey, err
}

//ValidateAnnotationByPublicKey is used for validate the annotation(cipher) by public key
func ValidateAnnotationByPublicKey(cipherText string, key *rsa.PublicKey) error {
	parts := strings.Split(cipherText, ".")
	if len(parts) != 3 {
		return errors.New("Invalid token received, token must have 3 parts")
	}
	
	jwtHeaderRcvd, _ := base64.StdEncoding.DecodeString(parts[0])
	var jwtHeader JwtHeader
	err := json.Unmarshal(jwtHeaderRcvd, &jwtHeader)
	if err != nil {
		Log.Errorf("%+v", err)
		return errors.New("Failed to unmarshal jwt header")
	}
	pubKey := util.GetAHPublicKey()
	block, _ := pem.Decode(pubKey)
	if block == nil || block.Type != "PUBLIC KEY" {
		Log.Fatal("failed to decode PEM block containing public key")
	}
	keyIdBytes := sha1.Sum(block.Bytes)
	keyIdStr := base64.StdEncoding.EncodeToString(keyIdBytes[:])

	if jwtHeader.KeyId != keyIdStr{
		return errors.New("Invalid Kid")
	}

	signedContent, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		Log.Errorf("Error while base64 decoding of trust report content %+v", err)
                return err
        }

	signatureString, _ := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		Log.Errorf("Error while base64 decoding of signature %+v", err)
		return err
	}

	h := sha512.New384()
	h.Write(signedContent)
	return rsa.VerifyPKCS1v15(key, crypto.SHA384, h.Sum(nil), signatureString)
}

//JWTParseWithClaims is used for parsing and adding the annotation values in claims map
func JWTParseWithClaims(cipherText string, verifyKey *rsa.PublicKey, claim jwt.MapClaims) {
	token, err := jwt.ParseWithClaims(cipherText, claim, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})
	Log.Infof("Parsed token is :", token)
	if err != nil {
		Log.Errorf("error in JWTParseWithClaims")
	}
}

//CheckAnnotationAttrib is used to validate node with respect to time,trusted and location tags
func CheckAnnotationAttrib(cipherText string, node []v1.NodeSelectorRequirement, trustPrefix string) bool {
	var claims = jwt.MapClaims{}
	pubKey := util.GetAHPublicKey()
	verifyKey, err := ParseRSAPublicKeyFromPEM(pubKey)
	if err != nil {
		Log.Errorf("Invalid AH public key")
		return false
	}
	validationStatus := ValidateAnnotationByPublicKey(cipherText, verifyKey)
	if validationStatus == nil {
		Log.Infof("Signature is valid, trust report is from valid Attestation Hub")
	} else {
		Log.Errorf("%v", validationStatus)
		Log.Errorf("Signature validation failed")
		return false
	}

	//cipherText is the annotation applied to the node, claims is the parsed AH report assigned as the annotation
	JWTParseWithClaims(cipherText, verifyKey, claims)

	Log.Infof("CheckAnnotationAttrib - Parsed claims for %v",  claims)

	verify := ValidatePodWithAnnotation(node, claims, trustPrefix)
	if verify {
		Log.Infoln("Node label validated against node annotations succesful")
	} else {
		Log.Infoln("Node Label did not match node annotation ")
		return false
	}

	trustTimeFlag := ValidateNodeByTime(claims)

	if trustTimeFlag == 1 {
		Log.Infoln("Attested node validity time check passed")
		return true
	} else {
		Log.Infoln("Attested node validity time has expired")
		return false
	}
}