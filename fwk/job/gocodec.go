// Copyright 2017 The go-hep Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package job

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"os"
	"reflect"
	"strings"
)

// NewGoEncoder returns a new encoder that writes to w
func NewGoEncoder(w io.Writer) *GoEncoder {
	if w == nil {
		w = os.Stdout
	}
	return &GoEncoder{
		w:   w,
		buf: new(bytes.Buffer),
	}
}

// A GoEncoder writes a go representation to an output stream
type GoEncoder struct {
	w    io.Writer
	buf  *bytes.Buffer
	init bool
}

// Encode encodes data into the underlying io.Writer
func (enc *GoEncoder) Encode(data interface{}) error {
	var err error
	stmts, ok := data.([]Stmt)
	if !ok {
		return fmt.Errorf("fwk/job: expected a []job.Stmt as input. got %T", data)
	}

	err = enc.encode(stmts)
	if err != nil {
		return err
	}

	buf, err := format.Source(enc.buf.Bytes())
	if err != nil {
		return err
	}

	_, err = enc.w.Write(buf)
	if err != nil {
		return err
	}

	return err
}

func (enc *GoEncoder) encode(stmts []Stmt) error {
	var err error
	if !enc.init {
		fmt.Fprintf(enc.buf, `// automatically generated by go-hep.org/x/hep/fwk/job.
// do NOT edit!

package main

`,
		)

		dpkgs := make(map[string]struct{}) // dash-import packages
		ipkgs := make(map[string]struct{}) // import packages
		for _, stmt := range stmts {
			switch stmt.Type {
			case StmtNewApp, StmtCreate:
				typename := stmt.Data.Type
				i := strings.LastIndex(typename, ".")
				if i == -1 {
					return fmt.Errorf("fwk/job: invalid package name %q (no dot)", typename)
				}
				//typ := typename[i+1:]
				pkg := typename[:i]
				dpkgs[pkg] = struct{}{}
			}

			for _, v := range stmt.Data.Props {
				typ := reflect.TypeOf(v)
				pkg := typ.PkgPath()
				if pkg == "" {
					continue
				}
				ipkgs[pkg] = struct{}{}
			}
		}

		fmt.Fprintf(enc.buf, "import (\n")
		fmt.Fprintf(enc.buf, "\t%q\n\n", "go-hep.org/x/hep/fwk/job")

		for pkg := range ipkgs {
			fmt.Fprintf(enc.buf, "\t%q\n", pkg)
		}

		for pkg := range dpkgs {
			_, dup := ipkgs[pkg]
			if dup {
				// already imported.
				continue
			}
			fmt.Fprintf(enc.buf, "\t_ %q\n", pkg)
		}
		fmt.Fprintf(enc.buf, ")\n")

		// first stmt should be the NewApp one.
		if stmts[0].Type != StmtNewApp {
			return fmt.Errorf("fwk/job: invalid stmts! expected stmts[0].Type==%v. got=%v",
				StmtNewApp,
				stmts[0].Type,
			)
		}

		if stmts[0].Data.Type != "go-hep.org/x/hep/fwk.appmgr" {
			// only support fwk.appmgr for now...
			return fmt.Errorf("fwk/job: invalid fwk.App concrete type (%v)", stmts[0].Data.Type)
		}

		fmt.Fprintf(enc.buf, `
func newApp() *job.Job {
	app := job.New(%s)
	return app
}

`,
			enc.repr(stmts[0].Data.Props),
		)

		enc.init = true
	}

	fmt.Fprintf(enc.buf, "\nfunc config(app *job.Job) {\n\n")

	for _, stmt := range stmts {

		switch stmt.Type {
		case StmtNewApp:
			continue

		case StmtCreate:
			const tmpl = `
	app.Create(job.C{
		Type:  %q,
		Name:  %q,
		Props: `
			fmt.Fprintf(
				enc.buf,
				tmpl,
				stmt.Data.Type,
				stmt.Data.Name,
			)
			repr := enc.repr(stmt.Data.Props)
			fmt.Fprintf(enc.buf, "%s,\n", string(repr))

			fmt.Fprintf(enc.buf, "\t})\n")

		case StmtSetProp:
			var key string
			var val interface{}
			for k, v := range stmt.Data.Props {
				key = k
				val = v
			}
			fmt.Fprintf(
				enc.buf,
				"app.SetProp(app.App().Component(%q), %q, %s)\n",
				stmt.Data.Name,
				key,
				enc.value(val),
			)

		default:
			return fmt.Errorf("fwk/job: invalid statement type (%#v)", stmt.Type)
		}
	}

	fmt.Fprintf(enc.buf, "\n} // config\n")

	return err
}

func (enc *GoEncoder) repr(props P) []byte {
	var buf bytes.Buffer
	if len(props) <= 0 {
		fmt.Fprintf(&buf, "nil")
		return buf.Bytes()
	}

	fmt.Fprintf(&buf, "job.P{\n")
	for k, v := range props {
		prop := enc.value(v)
		fmt.Fprintf(&buf, "\t%q: %s,\n", k, prop)
	}
	fmt.Fprintf(&buf, "}")

	return buf.Bytes()
}

func (enc *GoEncoder) value(v interface{}) string {
	typ := reflect.TypeOf(v)
	prop := ""
	switch typ.Kind() {
	case reflect.Struct:
		prop = fmt.Sprintf("%#v", v)
	case reflect.String:
		prop = fmt.Sprintf("%q", v)
	default:
		pkgname := typ.PkgPath()
		if pkgname != "" {
			idx := strings.LastIndex(pkgname, "/")
			pkgname = pkgname[idx+1:] + "."
		} else {
			pkgname = ""
		}
		prop = fmt.Sprintf("%s%s(%v)", pkgname, typ.Name(), v)
	}
	return prop
}
