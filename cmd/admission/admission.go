/*
 * Copyright 2017 caicloud authors. All rights reserved.
 */

package main

import (
	"fmt"

	"github.com/zxq-bit/kube-admission-test/pkg/version"
)

func main() {
	fmt.Println("Hello world admin")
	fmt.Printf("version '%v', commit '%v', repo root '%v'\n",
		version.VERSION, version.COMMIT, version.REPOROOT)
}
