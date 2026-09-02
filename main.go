// Command heic2jpg converts every HEIC image in a directory to a
// high-quality JPEG, preserving EXIF metadata.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

func main() {
	quality := flag.Int("q", 95, "JPEG quality (1-100)")
	outDir := flag.String("out", "", "write JPEGs to this directory instead of beside the originals")
	recursive := flag.Bool("r", false, "descend into subdirectories")
	force := flag.Bool("f", false, "overwrite existing JPEGs")
	workers := flag.Int("j", runtime.NumCPU(), "number of files to convert in parallel")
	dryRun := flag.Bool("n", false, "list what would be converted without writing anything")
	verbose := flag.Bool("v", false, "report each file as it is converted")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "heic2jpg converts HEIC images to high-quality JPEGs.\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] [directory]\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "The directory defaults to the current one. Originals are never modified.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() > 1 {
		fatalf("expected at most one directory, got %d", flag.NArg())
	}
	dir := "."
	if flag.NArg() == 1 {
		dir = flag.Arg(0)
	}
	if *quality < 1 || *quality > 100 {
		fatalf("quality %d out of range (1-100)", *quality)
	}
	if *workers < 1 {
		fatalf("worker count must be at least 1")
	}

	jobs, err := Scan(dir, *outDir, *recursive)
	if err != nil {
		fatalf("%v", err)
	}
	if len(jobs) == 0 {
		fmt.Printf("no HEIC files found in %s\n", dir)
		return
	}

	todo, skipped := partition(jobs, *force)

	if *dryRun {
		for _, j := range todo {
			fmt.Printf("%s -> %s\n", j.Src, j.Dst)
		}
		fmt.Printf("%d to convert, %d skipped\n", len(todo), skipped)
		return
	}

	failed := run(todo, *quality, *workers, *verbose)

	fmt.Printf("%d converted, %d skipped, %d failed\n", len(todo)-failed, skipped, failed)
	if failed > 0 {
		os.Exit(1)
	}
}

// partition splits jobs into those to convert and a count of those whose
// output already exists. Without -f, finished work is left alone so a
// re-run costs nothing.
func partition(jobs []Job, force bool) (todo []Job, skipped int) {
	for _, j := range jobs {
		if !force {
			if _, err := os.Stat(j.Dst); err == nil {
				skipped++
				continue
			}
		}
		todo = append(todo, j)
	}
	return todo, skipped
}

// run converts jobs across workers goroutines and returns the number that
// failed. One bad file does not stop the rest of the batch.
func run(jobs []Job, quality, workers int, verbose bool) int {
	if workers > len(jobs) {
		workers = len(jobs)
	}

	queue := make(chan Job)
	var mu sync.Mutex
	failed := 0

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range queue {
				err := Convert(j.Src, j.Dst, quality)
				mu.Lock()
				switch {
				case err != nil:
					failed++
					fmt.Fprintf(os.Stderr, "heic2jpg: %s: %v\n", j.Src, err)
				case verbose:
					fmt.Printf("%s -> %s\n", j.Src, j.Dst)
				}
				mu.Unlock()
			}
		}()
	}

	for _, j := range jobs {
		queue <- j
	}
	close(queue)
	wg.Wait()

	return failed
}

func warnf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "heic2jpg: warning: "+format+"\n", args...)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "heic2jpg: "+format+"\n", args...)
	os.Exit(2)
}
