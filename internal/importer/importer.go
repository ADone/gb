// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package importer

import (
	"fmt"
	"go/build"
	"os"
	pathpkg "path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

// A Package describes the Go package found in a directory.
type Package struct {
	*build.Package
	Standard bool // package found in GOROOT
}

type Importer struct {
	*build.Context
	Root string // root directory
}

func (i *Importer) Import(path string) (*Package, error) {
	if path == "" {
		return nil, fmt.Errorf("import %q: invalid import path", path)
	}

	if path == "." || path == ".." || strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") {
		return nil, fmt.Errorf("import %q: relative import not supported", path)
	}

	if strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("import %q: cannot import absolute path", path)
	}

	p := &Package{
		Standard: i.Root == runtime.GOROOT(),
	}

	loadPackage := func(importpath, dir string) error {
		pkg, err := i.ImportDir(dir, 0)
		if err != nil {
			return err
		}
		p.Package = pkg
		p.ImportPath = importpath
		return nil
	}

	// if this is the stdlib, then search vendor first.
	// this isn't real vendor support, just enough to make net/http compile.
	if p.Standard {
		path := pathpkg.Join("vendor", path)
		dir := filepath.Join(i.Root, "src", filepath.FromSlash(path))
		fi, err := os.Stat(dir)
		if err == nil && fi.IsDir() {
			err := loadPackage(path, dir)
			return p, err
		}
	}

	dir := filepath.Join(i.Root, "src", filepath.FromSlash(path))
	fi, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, errors.Errorf("import %q: not a directory", path)
	}
	err = loadPackage(path, dir)
	return p, err
}
