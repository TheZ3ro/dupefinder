package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/TheZ3ro/dupefinder"
)

const (
	txtHelp = `Usage: gofuniq [-dryrun|-rm] folder
    Delete every duplicate leaving only 1 of each unique file in folder`
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Detection flags
	dryrun := true
	rm := false

	flag.BoolVar(&dryrun, "dryrun", false, "Print what would be deleted")
	flag.BoolVar(&rm, "rm", false, "Delete detected duplicates (at your own risk!)")

	flag.Usage = func() {
		fmt.Println(txtHelp)
		fmt.Println()
		flag.PrintDefaults()
	}

	flag.Parse()

	args := flag.Args()

	if !dryrun && !rm {
		fmt.Println("Either -rm or -dryrun should be specified")
		os.Exit(1)
	}

	if dryrun && rm {
		fmt.Println("Only one of -rm or -dryrun should be specified")
		os.Exit(1)
	}

	if len(args) < 1 {
    fmt.Println("Too few argument, got",len(args)," expected 1")
		fmt.Println(txtHelp)
		os.Exit(1)
	}

  if rm {
    var choice string = "n"
    fmt.Println("Are you sure to delete all the duplicates? [y/n] \nRun with -dryrun to see what file will be deleted")
    if _, err := fmt.Scan(&choice); err != nil {
      fmt.Println("  Scan for choice failed, due to", err)
      os.Exit(1)
    }
    if choice == "y" {
      fmt.Println("  Ok")
    } else {
      rm = false
      dryrun = true
    }
  }

  count, catalog, err := dupefinder.Generate(args[0:]...)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  delSize, deleteEntries, err := dupefinder.Detect(catalog, true, args[0:]...)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  for index := range deleteEntries {
    entry := deleteEntries[index]
    if dryrun {
      fmt.Printf("Would delete %s (matches %s)\n", entry.Filename, entry.Origin)
    }

    if rm {
      fmt.Printf("Deleting %s (matches %s)\n", entry.Filename, entry.Origin)
      err := os.Remove(entry.Filename)
      if err != nil {
        fmt.Println(err)
        os.Exit(1)
      }
    }
  }

  fmt.Println("Total unique file:",len(catalog)," /",count)
  fmt.Printf("Total size of duplicates: %d bytes\n", delSize)
}
