package main

import (
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

const fixture = "testdata/sample.heic"

func TestConvertProducesDecodableJPEG(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "out.jpg")

	if err := Convert(fixture, dst, 95); err != nil {
		t.Fatalf("Convert: %v", err)
	}

	f, err := os.Open(dst)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	defer f.Close()

	img, err := jpeg.Decode(f)
	if err != nil {
		t.Fatalf("output is not a decodable JPEG: %v", err)
	}
	if got, want := img.Bounds().Dx(), 200; got != want {
		t.Errorf("width = %d, want %d", got, want)
	}
	if got, want := img.Bounds().Dy(), 120; got != want {
		t.Errorf("height = %d, want %d", got, want)
	}
}

// The fixture's top-left 40x40 block is solid red. A decoder that flips,
// rotates, or misreads the image puts something else there.
func TestConvertPreservesImageOrientation(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "out.jpg")
	if err := Convert(fixture, dst, 95); err != nil {
		t.Fatalf("Convert: %v", err)
	}

	f, _ := os.Open(dst)
	defer f.Close()
	img, err := jpeg.Decode(f)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	r, g, b, _ := img.At(10, 10).RGBA()
	r, g, b = r>>8, g>>8, b>>8
	if r < 200 || g > 60 || b > 60 {
		t.Errorf("top-left pixel = (%d,%d,%d), want approximately red", r, g, b)
	}
}

// A HEIC with no EXIF at all must still convert. goheif reports a
// missing EXIF block as an error, which is not a conversion failure.
func TestConvertSucceedsWithoutExif(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "out.jpg")
	if err := Convert(fixture, dst, 95); err != nil {
		t.Fatalf("Convert on EXIF-less input: %v", err)
	}
}

func TestConvertCorruptInputLeavesNothingBehind(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "broken.heic")
	if err := os.WriteFile(src, []byte("this is not a HEIC file"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "broken.jpg")

	if err := Convert(src, dst, 95); err == nil {
		t.Fatal("Convert succeeded on corrupt input, want error")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != "broken.heic" {
			t.Errorf("left behind %q, want only the source file", e.Name())
		}
	}
}

func TestConvertRejectsBadQuality(t *testing.T) {
	for _, q := range []int{0, -1, 101} {
		dst := filepath.Join(t.TempDir(), "out.jpg")
		if err := Convert(fixture, dst, q); err == nil {
			t.Errorf("Convert with quality %d succeeded, want error", q)
		}
	}
}

// Higher quality must yield a larger file; this is what makes the -q
// flag meaningful.
func TestConvertQualityAffectsSize(t *testing.T) {
	dir := t.TempDir()
	low := filepath.Join(dir, "low.jpg")
	high := filepath.Join(dir, "high.jpg")

	if err := Convert(fixture, low, 40); err != nil {
		t.Fatal(err)
	}
	if err := Convert(fixture, high, 95); err != nil {
		t.Fatal(err)
	}

	lo, _ := os.Stat(low)
	hi, _ := os.Stat(high)
	if hi.Size() <= lo.Size() {
		t.Errorf("quality 95 produced %d bytes, quality 40 produced %d; want larger", hi.Size(), lo.Size())
	}
}
