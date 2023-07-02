// Copyright (C) 2021-2023 YumeMichi
//
// SPDX-License-Identifier: Apache-2.0
package utils

import (
	"os"
)

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func ReadAllText(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}

func WriteAllText(path, text string) {
	_ = os.WriteFile(path, []byte(text), 0644)
}

func Xor(s1, s2 []byte) (res []byte) {
	for k, b := range s1 {
		newBt := b ^ s2[k]
		res = append(res, newBt)
	}

	return
}