package main

import (
	"log"
	"os"
	"os/exec"
)

func parse_mountopts(opts string) (string, error) {
	// TODO: add checks here if necessary
	return opts, nil
}

func main() {
	target := "/usr/local/sbin/fusermount3"

	// Forward all arguments except argv[0]
	args := os.Args[1:]

	var parsed_opts, mountpoint, action string

	for argno := range args {
		if args[argno] == "-o" {
			parsed_opts, _ = parse_mountopts(args[argno+1])

		}
		if args[argno] == "--" {
			mountpoint = args[argno+1]
			action = "mount"
		}
		if args[argno] == "-u" {
			mountpoint = args[argno+1]
			action = "umount"
		}

	}

	var safe_args []string

	// build fusermount arguments
	if action == "mount" {
		safe_args = []string{"-o", parsed_opts, "--", mountpoint}
	}
	if action == "umount" {
		safe_args = []string{"-u", mountpoint}
	}

	cmd := exec.Command(target, safe_args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("error running %s: %v", target, err)
	}
}
