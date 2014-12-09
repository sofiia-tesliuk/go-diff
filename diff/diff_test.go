package diff

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseHunkNoChunksize(t *testing.T) {
	filename := "sample_no_chunksize.diff"
	diffData, err := ioutil.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatal(err)
	}
	diff, err := ParseHunks(diffData)
	if err != nil {
		t.Errorf("%s: got ParseHunks err %v,  want %v", filename, err, nil)
	}
	if len(diff) != 1 {
		t.Errorf("%s: Got %d hunks, want only one", filename, len(diff))
	}

	correct := &Hunk{
		NewLines:      1,
		NewStartLine:  1,
		OrigLines:     0,
		OrigStartLine: 0,
	}
	h := diff[0]
	h.Body = nil // We're not testing the body.
	if !reflect.DeepEqual(h, correct) {
		t.Errorf("%s: Got %#v, want %#v", filename, h, correct)
	}
}

func TestParseHunksAndPrintHunks(t *testing.T) {
	tests := []struct {
		filename     string
		wantParseErr error
	}{
		{
			filename: "sample_hunk.diff",
		},
		{
			filename: "sample_hunks.diff",
		},
		{
			filename:     "sample_bad_hunks.diff",
			wantParseErr: nil,
		},
		{
			filename: "sample_hunks_no_newline.diff",
		},
		{
			filename: "sample_hunks_no_newline2.diff",
		},
		{filename: "empty.diff"},
		{filename: "oneline_hunk.diff"},
	}
	for _, test := range tests {
		diffData, err := ioutil.ReadFile(filepath.Join("testdata", test.filename))
		if err != nil {
			t.Fatal(err)
		}
		diff, err := ParseHunks(diffData)
		if err != test.wantParseErr {
			t.Errorf("%s: got ParseHunks err %v, want %v", test.filename, err, test.wantParseErr)
			continue
		}
		if test.wantParseErr != nil {
			continue
		}

		printed, err := PrintHunks(diff)
		if err != nil {
			t.Errorf("%s: PrintHunks: %s", test.filename, err)
		}
		if !bytes.Equal(printed, diffData) {
			t.Errorf("%s: printed diff hunks != original diff hunks\n\n# PrintHunks output:\n%s\n\n# Original:\n%s", test.filename, printed, diffData)
		}
	}
}

func TestStripModifierLines(t *testing.T) {
	orig := `- a
\ No newline at end of file
+ b
\ No newline at end of file
`
	want := `- a
+ b
`

	hunks := []*Hunk{{Body: []byte(orig)}}
	StripModifierLines(hunks)
	stripped := string(hunks[0].Body)

	if stripped != want {
		t.Errorf("Want:\n%s\nbut got:\n%s", want, stripped)
	}
}

func TestParseFileDiffAndPrintFileDiff(t *testing.T) {
	tests := []struct {
		filename     string
		wantParseErr error
	}{
		{
			filename: "sample_file.diff",
		},
		{
			filename: "sample_file_no_timestamp.diff",
		},
		{
			filename: "sample_file_extended.diff",
		},
		{
			filename:     "empty.diff",
			wantParseErr: &ParseError{0, 0, ErrExtendedHeadersEOF},
		},
	}
	for _, test := range tests {
		diffData, err := ioutil.ReadFile(filepath.Join("testdata", test.filename))
		if err != nil {
			t.Fatal(err)
		}
		diff, err := ParseFileDiff(diffData)
		if !reflect.DeepEqual(err, test.wantParseErr) {
			t.Errorf("%s: got ParseFileDiff err %v, want %v", test.filename, err, test.wantParseErr)
			continue
		}
		if test.wantParseErr != nil {
			continue
		}

		printed, err := PrintFileDiff(diff)
		if err != nil {
			t.Errorf("%s: PrintFileDiff: %s", test.filename, err)
		}
		if !bytes.Equal(printed, diffData) {
			t.Errorf("%s: printed file diff != original file diff\n\n# PrintFileDiff output:\n%s\n\n# Original:\n%s", test.filename, printed, diffData)
		}
	}
}

func TestParseMultiFileDiffAndPrintMultiFileDiff(t *testing.T) {
	tests := []struct {
		filename     string
		wantParseErr error
	}{
		{
			filename: "sample_multi_file.diff",
		},
		{
			filename: "sample_multi_file_single.diff",
		},
		{filename: "empty.diff"},
	}
	for _, test := range tests {
		diffData, err := ioutil.ReadFile(filepath.Join("testdata", test.filename))
		if err != nil {
			t.Fatal(err)
		}
		diff, err := ParseMultiFileDiff(diffData)
		if err != test.wantParseErr {
			t.Errorf("%s: got ParseMultiFileDiff err %v, want %v", test.filename, err, test.wantParseErr)
			continue
		}
		if test.wantParseErr != nil {
			continue
		}

		printed, err := PrintMultiFileDiff(diff)
		if err != nil {
			t.Errorf("%s: PrintMultiFileDiff: %s", test.filename, err)
		}
		if !bytes.Equal(printed, diffData) {
			t.Errorf("%s: printed multi-file diff != original multi-file diff\n\n# PrintMultiFileDiff output:\n%s\n\n# Original:\n%s", test.filename, printed, diffData)
		}
	}
}

func TestFileDiff_Stat(t *testing.T) {
	tests := map[string]struct {
		hunks []*Hunk
		want  Stat
	}{
		"no change": {
			hunks: []*Hunk{
				{Body: []byte(`@@ -0,0 +0,0
 a
 b
`)},
			},
			want: Stat{},
		},
		"added/deleted": {
			hunks: []*Hunk{
				{Body: []byte(`@@ -0,0 +0,0
+a
 b
-c
 d
`)},
			},
			want: Stat{Added: 1, Deleted: 1},
		},
		"changed": {
			hunks: []*Hunk{
				{Body: []byte(`@@ -0,0 +0,0
+a
+b
-c
-d
 e
`)},
			},
			want: Stat{Added: 1, Changed: 1, Deleted: 1},
		},
		"many changes": {
			hunks: []*Hunk{
				{Body: []byte(`@@ -0,0 +0,0
+a
-b
+c
-d
 e
`)},
			},
			want: Stat{Added: 0, Changed: 2, Deleted: 0},
		},
	}
	for label, test := range tests {
		fdiff := &FileDiff{Hunks: test.hunks}
		stat := fdiff.Stat()
		if !reflect.DeepEqual(stat, test.want) {
			t.Errorf("%s: got diff stat %+v, want %+v", label, stat, test.want)
			continue
		}
	}
}
