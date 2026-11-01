package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lukasjoc/prec/snap"
)

func ttyPrompt(rd *bufio.Reader, prompttext string) (bool, error) {
	for {
		fmt.Fprintf(os.Stdout, "%s [Yn]? ", prompttext)
		anstext, err := rd.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("ttyPrompt: could not read from tty: error: %w\n", err)
		}
		ans := strings.Trim(anstext, " \n\r")
		switch ans {
		case "", "y", "Y":
			return true, nil
		case "n", "N":
			return false, nil
		}
		fmt.Fprintf(os.Stdout, "error: invalid reply: \"%s\"\n", ans)
	}
}

func bailWithUsage(err error) {
	program := os.Args[0]
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nhelp: %s --help\n", program)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "usage: %s [--snapshots-dir,--help]\n", program)
	fmt.Fprintf(os.Stderr, "    --snapshots-dir   output directory for the snapshot files (default: .snapdata)\n")
	fmt.Fprintf(os.Stderr, "    -i, --interactive enter interactive fix-mode\n")
	fmt.Fprintf(os.Stderr, "    -h, --help        show this help\n")
	fmt.Fprintf(os.Stderr, "example: %s --snapshots-dir .snapdata2 -i\n", program)
	os.Exit(0)
}

func main() {
	flags := flag.NewFlagSet("options", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	snapshotsDir := flags.String("snapshots-dir", ".snapdata", "")

	var modeInteractive bool
	flags.BoolVar(&modeInteractive, "i", false, "")
	flags.BoolVar(&modeInteractive, "interactive", false, "")

	if err := flags.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			bailWithUsage(nil)
		}
		bailWithUsage(err)
	}

	if modeInteractive && snapshotsDir != nil {
		if err := os.Mkdir(*snapshotsDir, 0755); err != nil && !os.IsExist(err) {
			fmt.Fprintf(os.Stderr, "snap: could not create snapshots dir: error: %v", err)
			os.Exit(1)
		}
	}

	var ttyin *bufio.Reader
	if modeInteractive {
		tty, err := os.Open("/dev/tty")
		if err != nil {
			fmt.Fprintf(os.Stderr, "snap: could not get tty: error: %v", err)
			os.Exit(1)
		}
		defer tty.Close()

		ttyin = bufio.NewReader(tty)
	}

	failed := false
	_ = ttyin

	modified, err := filepath.Glob(fmt.Sprintf("%s/*.snap.modified", *snapshotsDir))
	if err != nil {
		panic(err)
	}

	news, err := filepath.Glob(fmt.Sprintf("%s/*.snap.added", *snapshotsDir))
	if err != nil {
		panic(err)
	}

	for _, path := range news {
		name := filepath.Base(path)

		// NOTE: here there should not exist a .snap for this
		// but we should verify that later
		incoming, err, found := snap.LoadSnapshot(name, *snapshotsDir)
		if !found {
			panic(err)
		}
		if err != nil {
			panic(err)
		}

		testname := strings.TrimSuffix(name, ".snap.added")

		prompttext := "Accept new snapshot?"
		// TODO(NOMERGE): pretty printed, indented
		fmt.Fprintf(os.Stdout, "[A]: Name: %s\n", testname)
		fmt.Fprintf(os.Stdout, "[A]: Description: %s\n", incoming.Metadata.Description)
		fmt.Fprintf(os.Stdout, "[A]: %s\n\n", incoming.Output)

		if modeInteractive {
			ans, err := ttyPrompt(ttyin, prompttext)
			if err != nil {
				fmt.Fprintf(os.Stderr, "snap: could not read from tty: error: %v\n", err)
				break
			}
			if ans {
				if err := snap.WriteSnapshot(fmt.Sprintf("%s.snap", testname), *snapshotsDir, incoming); err != nil {
					fmt.Fprintf(os.Stderr, "snap: could not write snapshot: error: %v\n", err)
					break
				}
				if err := os.Remove(path); err != nil {
					panic(err)
				}
			} else {
				failed = true
			}
		} else {
			failed = true
		}
	}

	for _, path := range modified {
		name := filepath.Base(path)
		incoming, err, found := snap.LoadSnapshot(name, *snapshotsDir)
		if !found {
			panic(err)
		}
		if err != nil {
			panic(err)
		}
		// NOTE: Its fine to do this in the loop over all of them
		// as we dont really use the '.snap' dirents at all here.
		// But maybe we want to improve this with Glob?
		// NOTE: here there should exist a .snap for this
		testname := strings.TrimSuffix(name, ".snap.modified")

		existing, err, found := snap.LoadSnapshot(fmt.Sprintf("%s.snap", testname), *snapshotsDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "snap: could not load snapshot: error: %v\n", err)
			break
		}

		prompttext := "Accept updates?"
		fmt.Fprintf(os.Stdout, "[M] Name: %s\n", testname)
		fmt.Fprintf(os.Stdout, "Description: %s\n", existing.Metadata.Description)
		fmt.Fprintf(os.Stdout, "EXISTING ==>\n")
		fmt.Fprintf(os.Stdout, "%s\n", existing.Output)
		fmt.Fprintf(os.Stdout, "<== INCOMING\n")
		fmt.Fprintf(os.Stdout, "%s\n\n", incoming.Output)

		if modeInteractive {
			ans, err := ttyPrompt(ttyin, prompttext)
			if err != nil {
				fmt.Fprintf(os.Stderr, "snap: could not read from tty: error: %v\n", err)
				break
			}
			if ans {
				if err := snap.WriteSnapshot(fmt.Sprintf("%s.snap", testname), *snapshotsDir, incoming); err != nil {
					fmt.Fprintf(os.Stderr, "snap: could not write snapshot: error: %v\n", err)
					break
				}
				if err := os.Remove(path); err != nil {
					panic(err)
				}
			} else {
				failed = true
			}
		} else {
			failed = true
		}
	}

	if failed {
		os.Exit(1)
	}
}
