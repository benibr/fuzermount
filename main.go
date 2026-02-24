package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/phuslu/log"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Target         string   `yaml:"target"`
	AllowedParents []string `yaml:"allowedParents"`
	MandatoryOpts  []string `yaml:"mandatoryOpts"`
	ForbiddenOpts  []string `yaml:"forbiddenOpts"`
}

var config Config

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
	if slices.Contains(config.AllowedParents, parentPath) {
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
		if slices.Contains(config.MandatoryOpts, opt) {
			config.MandatoryOpts = slices.DeleteFunc(config.MandatoryOpts, func(s string) bool {
				return s == opt
			})
		}
		// check for forbidden options
		if slices.Contains(config.ForbiddenOpts, opt) {
			return "", fmt.Errorf("'%s' is a forbidden mount option. Denying fuse mount", opt)
		}
		// remove empty strings
		if opt == "" {
			continue
		}
		return_opts = append(return_opts, opt)
	}
	// check if all mandatory options are set
	if len(config.MandatoryOpts) == 0 {
		ret := strings.Join(return_opts, ",")
		return ret, nil
	} else {
		missing_opts := strings.Join(config.MandatoryOpts, ",")
		return "", fmt.Errorf("not all mandatory mount options set. Missing '%s'\nDenying fuse mount", missing_opts)
	}
}

func main() {
	logger := log.Logger{
		Level: log.ParseLevel("info"),
		Writer: &log.FileWriter{
			Filename: "fuzermount.log",
		},
	}

	configYaml, err := os.ReadFile("/etc/fuzermount/fuzermount.yaml")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	err = yaml.Unmarshal(configYaml, &config)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	// Forward all arguments except argv[0]
	args := os.Args[1:]

	logger.Info().Msg(strings.Join(args, " "))

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

	// check if parent is allowed
	if args[0] != "-u" {
		parentAllowed, err := check_parent()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		if !parentAllowed {
			fmt.Println("Calling fusermount3 is only allowed from")
			fmt.Println(config.AllowedParents)
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
			// checkDirectory cannot be called here because
			// with DAOS local root might not be allowed to read
			// the mounted directory
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

	cmd := exec.Command(config.Target, safe_args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logger.Fatal().Msg(fmt.Sprintf("error running %s: %v", config.Target, err))
	}
}
