package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// lists with permission
// TODO: these global vars could go to a config file
var target = "/opt/fuzermount/fusermount3"
var allowedParents = []string{"/usr/bin/dfuse"}
var mandatory_opts = []string{"nosuid", "nodev", "noatime", "default_permissions", "fsname=dfuse"}
var forbidden_opts = []string{"suid"}

func checkDirectory(path string) error {
	// checks if the given string is actually a available path in the system
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		return nil
	}
	return errors.New("given mountpoint is not a directory")
}

func check_parent() (bool, error) {
	// this function checks if the parent PID is a executable
	// whose full path is in the allowed list
	ppid := os.Getppid()
	procPath := fmt.Sprintf("/proc/%d/exe", ppid)

	parentPath, err := os.Readlink(procPath)
	if err != nil {
		return false, err
	}
	parentPath, err = filepath.EvalSymlinks(parentPath)
	if err != nil {
		return false, err
	}
	if slices.Contains(allowedParents, parentPath) {
		return true, nil
	}
	// FIXME: this should be returned and printed by main()
	fmt.Printf("fusermount3 was called by '%s'", parentPath)
	return false, nil
}

func check_mountopts(opts string) (string, error) {
	// this function checks for mount options that must or must not exist
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
	// check if all mandatory options are set
	if len(mandatory_opts) == 0 {
		ret := strings.Join(return_opts, ",")
		return ret, nil
	} else {
		missing_opts := strings.Join(mandatory_opts, ",")
		return "", fmt.Errorf("not all mandatory mount options set. Missing '%s'\nDenying fuse mount", missing_opts)
	}
}

func main() {

	// Forward all arguments except argv[0]
	args := os.Args[1:]

	// help output
	if len(args) < 2 {
		fmt.Println("This is NHR@ZIB mounting wrapper for FUSE filesystems.")
		fmt.Println(errors.New("unexpected arguments"))
		fmt.Println("Available commands are:")
		fmt.Println("  '-o some,mount,options -- /path/to/mountpoint/'")
		fmt.Println("  '-u /path/to/mountpoint/'")
		os.Exit(1)
	}

	var parsed_opts, mountpoint, action string
	var err error

	// check if parent is allowed
	if args[0] != "-u" {
		parentAllowed, err := check_parent()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		if !parentAllowed {
			fmt.Println("Calling fusermount3 is only allowed from")
			fmt.Println(allowedParents)
			os.Exit(1)
		}
	}

	// parse and check arguments
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
			err := checkDirectory(mountpoint)
			if err == nil {
				action = "mount"
			} else {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		if args[argno] == "-u" {
			mountpoint = args[argno+1]
			err := checkDirectory(mountpoint)
			if err == nil {
				action = "umount"
			} else {
				fmt.Println(err)
				os.Exit(1)
			}
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
