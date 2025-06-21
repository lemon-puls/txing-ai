package utils

import (
	"crypto/sha512"
	"fmt"
	"github.com/anaskhan96/go-password-encoder"
	"strings"
	"txing-ai/internal/global/logging/log"
)

var options = &password.Options{SaltLen: 16, Iterations: 100, KeyLen: 32, HashFunction: sha512.New}

//options := &password.Options{SaltLen: 10, Iterations: 10000, KeyLen: 50, HashFunction: passwd.New}

func EncryptPasswd(rawPassword string) string {
	salt, encodedPwd := password.Encode(rawPassword, options)
	dbPassword := fmt.Sprintf("$sha512$%s$%s", salt, encodedPwd)
	return dbPassword
}

func VerifyPasswd(rawPassword string, dbPassword string) bool {
	splits := strings.Split(dbPassword, "$")
	if len(splits) != 4 {
		log.Error("password format error")
		return false
	}
	// [ sha512 WI56n0Wmte5Ul0ui f89fdc8f8c7dc87c220d2331007b48d53e5eb2a9f64d2a31f0cedb4dd3f7c874]
	// splits[0]是空字符 splits[2]才是盐
	valid := password.Verify(rawPassword, splits[2], splits[3], options)
	return valid
}
