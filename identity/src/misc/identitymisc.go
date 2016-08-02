/*
// ----------------------------------------------------------------------------
// util.go
// Countertop Identity Microservice Misc Utility Functions

// Created by Paul Pietkiewicz on 8/27/2015
// Copyright (c) 2015 The Orange Chef Company. All rights reserved.
// ----------------------------------------------------------------------------
*/

package identitymisc

import (
	"crypto/rand"
	"encoding/base64"
)

const SESSION_KEY_LENGTH int = 64

// Bootlegged from https://elithrar.github.io/article/generating-secure-random-numbers-crypto-rand/

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateSessionKey() (string, error) {
	b, err := GenerateRandomBytes(SESSION_KEY_LENGTH)
	return base64.URLEncoding.EncodeToString(b), err
}
