// Copyright ©2020 The go-hep Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rcmd_test

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go-hep.org/x/hep/groot/rcmd"
)

func TestDump(t *testing.T) {
	const deep = true
	loadRef := func(fname string) string {
		t.Helper()
		raw, err := ioutil.ReadFile(fname)
		if err != nil {
			t.Fatalf("could not load reference file %q: %+v", fname, err)
		}
		return string(raw)
	}

	for _, tc := range []struct {
		name string
		want string
	}{
		{
			name: "../testdata/simple.root",
			want: `key[000]: tree;1 "fake data" (TTree)
[000][one]: 1
[000][two]: 1.1
[000][three]: uno
[001][one]: 2
[001][two]: 2.2
[001][three]: dos
[002][one]: 3
[002][two]: 3.3
[002][three]: tres
[003][one]: 4
[003][two]: 4.4
[003][three]: quatro
`,
		},
		{
			name: "../testdata/root_numpy_struct.root",
			want: `key[000]: test;1 "identical leaf names in different branches" (TTree)
[000][branch1.intleaf]: 10
[000][branch1.floatleaf]: 15.5
[000][branch2.intleaf]: 20
[000][branch2.floatleaf]: 781.2
`,
		},
		{
			name: "../testdata/padding.root",
			want: `key[000]: tree;1 "tree w/ & w/o padding" (TTree)
[000][pad.x1]: 0
[000][pad.x2]: 548655054794
[000][pad.x3]: 0
[000][nop.x1]: 0
[000][nop.x2]: 0
[000][nop.x3]: 0
[001][pad.x1]: 1
[001][pad.x2]: 72058142692982730
[001][pad.x3]: 0
[001][nop.x1]: 1
[001][nop.x2]: 1
[001][nop.x3]: 1
[002][pad.x1]: 2
[002][pad.x2]: 144115736730910666
[002][pad.x3]: 0
[002][nop.x1]: 2
[002][nop.x2]: 2
[002][nop.x3]: 2
[003][pad.x1]: 3
[003][pad.x2]: 216173330768838602
[003][pad.x3]: 0
[003][nop.x1]: 3
[003][nop.x2]: 3
[003][nop.x3]: 3
[004][pad.x1]: 4
[004][pad.x2]: 288230924806766538
[004][pad.x3]: 0
[004][nop.x1]: 4
[004][nop.x2]: 4
[004][nop.x3]: 4
`,
		},
		{
			name: "../testdata/small-flat-tree.root",
			want: loadRef("testdata/small-flat-tree.root.txt"),
		},
		{
			name: "../testdata/small-evnt-tree-fullsplit.root",
			want: loadRef("testdata/small-evnt-tree-fullsplit.root.txt"),
		},
		{
			name: "../testdata/small-evnt-tree-nosplit.root",
			want: loadRef("testdata/small-evnt-tree-nosplit.root.txt"),
		},
		{
			name: "../testdata/leaves.root",
			want: loadRef("testdata/leaves.root.txt"),
		},
		{
			name: "../testdata/embedded-std-vector.root",
			want: `key[000]: modules;1 "Module Tree Analysis" (TTree)
[000][hits_n]: 10
[000][hits_time_mc]: [12.206399 11.711122 11.73492 12.45704 11.558057 11.56502 11.687759 11.528914 12.893241 11.429288]
[001][hits_n]: 11
[001][hits_time_mc]: [11.718019 12.985347 12.23121 11.825082 12.405976 15.339471 11.939051 12.935032 13.661691 11.969542 11.893113]
[002][hits_n]: 15
[002][hits_time_mc]: [12.231329 12.214683 12.194867 12.246092 11.859249 19.35934 12.155213 12.226966 -4.712372 11.851829 11.8806925 11.8204975 11.866335 13.285733 -4.6470475]
[003][hits_n]: 9
[003][hits_time_mc]: [11.33844 11.725604 12.774131 12.108594 12.192085 12.120591 12.129445 12.18349 11.591005]
[004][hits_n]: 13
[004][hits_time_mc]: [12.156414 12.641215 11.678816 12.329707 11.578169 12.512748 11.840462 14.120602 11.875188 14.133265 14.105912 14.905052 11.813884]
`,
		},
		{
			// recovered baskets
			name: "../testdata/uproot/issue21.root",
			want: loadRef("../testdata/uproot/issue21.root.txt"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := new(strings.Builder)
			err := rcmd.Dump(got, tc.name, deep, nil)
			if err != nil {
				t.Fatalf("could not run root-dump: %+v", err)
			}

			if got, want := got.String(), tc.want; got != want {
				diff := cmp.Diff(want, got)
				t.Fatalf("invalid root-dump output: -- (-ref +got)\n%s", diff)
			}
		})
	}
}
