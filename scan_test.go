package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFiles(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, n := range names {
		p := filepath.Join(dir, n)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func srcNames(jobs []Job, dir string) []string {
	out := make([]string, len(jobs))
	for i, j := range jobs {
		rel, _ := filepath.Rel(dir, j.Src)
		out[i] = rel
	}
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestScanMatchesHeicExtensionsCaseInsensitively(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "a.heic", "b.HEIC", "c.HeIf", "d.heif")

	jobs, err := Scan(dir, "", false)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a.heic", "b.HEIC", "c.HeIf", "d.heif"}
	if got := srcNames(jobs, dir); !equal(got, want) {
		t.Errorf("sources = %v, want %v", got, want)
	}
}

func TestScanIgnoresOtherFiles(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "photo.heic", "photo.jpg", "notes.txt", "movie.mov", "noext")

	jobs, err := Scan(dir, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if got := srcNames(jobs, dir); !equal(got, []string{"photo.heic"}) {
		t.Errorf("sources = %v, want [photo.heic]", got)
	}
}

func TestScanIsNotRecursiveByDefault(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "top.heic", "sub/nested.heic")

	jobs, err := Scan(dir, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if got := srcNames(jobs, dir); !equal(got, []string{"top.heic"}) {
		t.Errorf("sources = %v, want [top.heic]", got)
	}
}

func TestScanRecursiveFindsNestedFiles(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "top.heic", "sub/nested.heic", "sub/deep/deeper.heic")

	jobs, err := Scan(dir, "", true)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"sub/deep/deeper.heic", "sub/nested.heic", "top.heic"}
	if got := srcNames(jobs, dir); !equal(got, want) {
		t.Errorf("sources = %v, want %v", got, want)
	}
}

func TestScanWritesBesideSourceByDefault(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "sub/nested.HEIC")

	jobs, err := Scan(dir, "", true)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "sub", "nested.jpg")
	if jobs[0].Dst != want {
		t.Errorf("dst = %q, want %q", jobs[0].Dst, want)
	}
}

func TestScanMirrorsTreeIntoOutDir(t *testing.T) {
	dir := t.TempDir()
	out := t.TempDir()
	writeFiles(t, dir, "top.heic", "sub/nested.heic")

	jobs, err := Scan(dir, out, true)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		filepath.Join(dir, "top.heic"):           filepath.Join(out, "top.jpg"),
		filepath.Join(dir, "sub", "nested.heic"): filepath.Join(out, "sub", "nested.jpg"),
	}
	for _, j := range jobs {
		if want[j.Src] != j.Dst {
			t.Errorf("dst for %q = %q, want %q", j.Src, j.Dst, want[j.Src])
		}
	}
}

func TestScanRejectsMissingDirectory(t *testing.T) {
	if _, err := Scan(filepath.Join(t.TempDir(), "nope"), "", false); err == nil {
		t.Fatal("Scan succeeded on a missing directory, want error")
	}
}

// Re-running the tool must not pick up the JPGs it produced, even when
// the output directory sits inside the scanned tree.
func TestScanSkipsOutputDirectory(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "converted")
	writeFiles(t, dir, "top.heic", "converted/stray.heic")

	jobs, err := Scan(dir, out, true)
	if err != nil {
		t.Fatal(err)
	}
	if got := srcNames(jobs, dir); !equal(got, []string{"top.heic"}) {
		t.Errorf("sources = %v, want [top.heic]", got)
	}
}
