package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"
)

func check_mountopts(opts string) (string, error) {
	// this function allows to set mountoptions that must or must not exist
	mandatory_opts := []string{"nosuid", "nodev", "noatime", "default_permissions", "fsname=dfuse", "subtype=daos"}
	forbidden_opts := []string{"suid"}
	var return_opts []string

	for opt := range strings.SplitSeq(opts, ",") {
		// check for mandatory options
		if slices.Contains(mandatory_opts, opt) {
			mandatory_opts = slices.DeleteFunc(mandatory_opts, func(s string) bool {
				return s == opt
			})
		}
		// check for forbidden options
		if slices.Contains(forbidden_opts, opt) {
			return "", fmt.Errorf("'%s' is a forbidden mount option. Denying fuse mount", opt)
		}

		// remove empty strings
		if opt == "" {
			continue
		}
		return_opts = append(return_opts, opt)
	}
	if len(mandatory_opts) == 0 {
		ret := strings.Join(return_opts, ",")
		return ret, nil
	} else {
		missing_opts := strings.Join(mandatory_opts, ",")
		return "", fmt.Errorf("not all mandatory mount options set. Missing '%s'\nDenying fuse mount", missing_opts)
	}
}

func main() {
	// FIXME: we need a better place for the original
	//        which is not in path
	target := "/opt/fuzermount/fusermount3"

	// Forward all arguments except argv[0]
	args := os.Args[1:]

	var parsed_opts, mountpoint, action string
	var err error

	for argno := range args {
		// TODO: add help output
		if args[argno] == "-o" {
			parsed_opts, err = check_mountopts(args[argno+1])
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

		}
		if args[argno] == "--" {
			mountpoint = args[argno+1]
			//TODO: check if mountpoint is a path
			action = "mount"
		}
		if args[argno] == "-u" {
			mountpoint = args[argno+1]
			//TODO: check if mountpoint is a path
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
