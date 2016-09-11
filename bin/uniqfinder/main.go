package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/thez3ro/dupefinder"
)

const (
	txtHelp = `Usage: gofuniq [-dryrun|-rm] folder
    Detects duplicate of file in the same folder that have different names`
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

  err := dupefinder.Generate(args[0], args[1:]...)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

	err := dupefinder.Detect(args[0], dryrun, rm, args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
