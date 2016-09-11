package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/TheZ3ro/dupefinder"
)

const (
	generateHelp = `Usage: dupefinder -generate filename folder...
    Generates a catalog file at filename based on one or more folders`

	detectHelp = `Usage: dupefinder -detect [-dryrun / -rm] filename folder...
    Detects duplicates using a catalog file in on one or more folders`
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	detect := true
	generate := false

	flag.BoolVar(&detect, "detect", false, "Detect duplicate files using a catalog")
	flag.BoolVar(&generate, "generate", false, "Generate a catalog file")

	// Detection flags
	dryrun := true
	rm := false

	flag.BoolVar(&dryrun, "dryrun", false, "Print what would be deleted")
	flag.BoolVar(&rm, "rm", false, "Delete detected duplicates (at your own risk!)")

	flag.Usage = func() {
		fmt.Println(generateHelp)
		fmt.Println()
		fmt.Println(detectHelp)
		fmt.Println()
		flag.PrintDefaults()
	}

	flag.Parse()

	args := flag.Args()

	if !detect && !generate {
		fmt.Println("Either -generate or -detect should be specified")
		os.Exit(1)
	}

	if detect && generate {
		fmt.Println("Only one of -generate or -detect should be specified")
		os.Exit(1)
	}

	if generate {

		if len(args) < 2 {
			fmt.Println("Too few argument, got",len(args)," expected 1")
			fmt.Println(generateHelp)
			os.Exit(1)
		}

		_, catalog, err := dupefinder.Generate(args[1:]...)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = dupefinder.WriteCatalog(args[0], catalog)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("Catalog file saved successfully to",args[0]," !")
	}

	if detect {
		if !dryrun && !rm {
			fmt.Println("Either -rm or -dryrun should be specified")
			os.Exit(1)
		}

		if dryrun && rm {
			fmt.Println("Only one of -rm or -dryrun should be specified")
			os.Exit(1)
		}

		if len(args) < 2 {
			fmt.Println("Too few argument, got",len(args)," expected 1")
			fmt.Println(detectHelp)
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

		catalogEntries, err := dupefinder.ParseCatalog(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		delSize, deleteEntries, err := dupefinder.Detect(catalogEntries, false, args[1:]...)
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

		fmt.Printf("Total size saved: %d bytes\n", delSize)
	}
}
