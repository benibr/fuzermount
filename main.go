package main

import (
	"log"
	"os"
	"os/exec"
)

func parse_mountopts(opts string) (string, error) {
	// stub
	return opts, nil
}

func main() {
	target := "/usr/local/sbin/fusermount3"

	// Forward all arguments except argv[0]
	args := os.Args[1:]

	var parsed_opts string
	var mountpoint string

	for argno := range args {
		if args[argno] == "-o" {
			parsed_opts, _ = parse_mountopts(args[argno+1])

		}
		if args[argno] == "--" {
			mountpoint = args[argno+1]
		}
	}

	safe_args := []string{"-o", parsed_opts, "--", mountpoint}
	// Prepare the command
	cmd := exec.Command(target, safe_args...)

	// Connect stdio so it behaves like the wrapped program
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run it
	if err := cmd.Run(); err != nil {
		log.Fatalf("error running %s: %v", target, err)
	}
}
