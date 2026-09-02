package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPartitionSkipsExistingOutput(t *testing.T) {
	dir := t.TempDir()
	done := filepath.Join(dir, "done.jpg")
	if err := os.WriteFile(done, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	jobs := []Job{
		{Src: filepath.Join(dir, "done.heic"), Dst: done},
		{Src: filepath.Join(dir, "fresh.heic"), Dst: filepath.Join(dir, "fresh.jpg")},
	}

	todo, skipped := partition(jobs, false)

	if skipped != 1 {
		t.Errorf("skipped = %d, want 1", skipped)
	}
	if len(todo) != 1 || todo[0].Dst != filepath.Join(dir, "fresh.jpg") {
		t.Errorf("todo = %v, want only fresh.jpg", todo)
	}
}

func TestPartitionForceIncludesExistingOutput(t *testing.T) {
	dir := t.TempDir()
	done := filepath.Join(dir, "done.jpg")
	if err := os.WriteFile(done, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	jobs := []Job{{Src: filepath.Join(dir, "done.heic"), Dst: done}}

	todo, skipped := partition(jobs, true)

	if skipped != 0 {
		t.Errorf("skipped = %d, want 0", skipped)
	}
	if len(todo) != 1 {
		t.Errorf("todo = %v, want the existing job included", todo)
	}
}

// One unreadable file must not cost the rest of the batch.
func TestRunContinuesAfterAFailure(t *testing.T) {
	dir := t.TempDir()
	broken := filepath.Join(dir, "broken.heic")
	if err := os.WriteFile(broken, []byte("not a HEIC"), 0o644); err != nil {
		t.Fatal(err)
	}
	jobs := []Job{
		{Src: broken, Dst: filepath.Join(dir, "broken.jpg")},
		{Src: fixture, Dst: filepath.Join(dir, "good.jpg")},
	}

	failed := run(jobs, 95, 2, false)

	if failed != 1 {
		t.Errorf("failed = %d, want 1", failed)
	}
	if _, err := os.Stat(filepath.Join(dir, "good.jpg")); err != nil {
		t.Errorf("good file was not converted: %v", err)
	}
}

// Overwriting must go through the same rename, leaving a complete file.
func TestConvertOverwritesExistingFile(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "out.jpg")
	if err := os.WriteFile(dst, []byte("stale contents"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Convert(fixture, dst, 95); err != nil {
		t.Fatalf("Convert: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 2 || data[0] != 0xFF || data[1] != 0xD8 {
		t.Error("destination was not replaced with a JPEG")
	}
}
