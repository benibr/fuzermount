package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/phuslu/log"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Target         string   `yaml:"target"`
	Mode           string   `yaml:"mode"`
	AllowedParents []string `yaml:"allowedParents"`
	MandatoryOpts  []string `yaml:"mandatoryOpts"`
	ForbiddenOpts  []string `yaml:"forbiddenOpts"`
}

var config Config
var logger log.Logger

func dropPrivileges(euid int, egid int) {
	// Unset supplementary group IDs.
	err := syscall.Setgroups([]int{})
	if err != nil {
		log.Fatal().Msg("Fa.Error()iled to unset supplementary group IDs: " + err.Error())
	}
	// Set group ID (real and effective).
	err = syscall.Setgid(egid)
	if err != nil {
		log.Fatal().Msg("Failed to set group ID: " + err.Error())
	}
	// Set user ID (real and effective).
	err = syscall.Setuid(euid)
	if err != nil {
		log.Fatal().Msg("Failed to set user ID: " + err.Error())
	}
}

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
	logger.Info().Msg(fmt.Sprintf("called by '%s'", parentPath))
	return false, nil
}

func checkOptIsMandatory(opt string) {
	if slices.Contains(config.MandatoryOpts, opt) {
		// remove all found mandatory opts from slice
		config.MandatoryOpts = slices.DeleteFunc(config.MandatoryOpts, func(s string) bool {
			return s == opt
		})
	}
}
func checkOptIsForbidden(opt string) error {
	// check for forbidden options
	if slices.Contains(config.ForbiddenOpts, opt) {
		return fmt.Errorf("'%s' is a forbidden mount option. Denying fuse mount", opt)
	}
	return nil
}

func checkAllMandatoryOptsPresent() error {
	if len(config.MandatoryOpts) == 0 {
		return nil
	}
	missing_opts := strings.Join(config.MandatoryOpts, ",")
	return fmt.Errorf("not all mandatory mount options set. Missing '%s'\nDenying fuse mount", missing_opts)
}

func parseMountOpts(opts string) []string {
	// this function checks for mount options that must or must not exist
	var return_opts []string

	for opt := range strings.SplitSeq(opts, ",") {
		// remove empty strings
		if opt == "" {
			continue
		}
		return_opts = append(return_opts, opt)
	}
	return return_opts
}

func main() {
	// init logging
	logger = log.Logger{
		Level: log.ParseLevel("info"),
		Writer: &log.FileWriter{
			Filename: "fuzermount.log",
		},
	}

	// parse config file
	configYaml, err := os.ReadFile("/etc/fuzermount/fuzermount.yaml")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	err = yaml.Unmarshal(configYaml, &config)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	// parse args
	args := os.Args[1:]
	argsString := strings.Join(args, " ")
	logger.Info().Msg(fmt.Sprintf("called with args: '%s'", argsString))
	// help output
	if len(args) < 2 {
		fmt.Println("This is NHR@ZIB mounting wrapper for FUSE filesystems.")
		fmt.Println(errors.New("unexpected arguments"))
		fmt.Println("Available commands are:")
		fmt.Println("  '-o some,mount,options -- /path/to/mountpoint/'")
		fmt.Println("  '-u /path/to/mountpoint/'")
		os.Exit(1)
	}

	// here the filtering starts
	var parsed_opts []string
	var mountpoint string
	action := "unknown"

	if config.Mode == "strict" || config.Mode == "relaxed" {

		// parse and check arguments
		for argno := range args {
			if args[argno] == "-o" {
				parsed_opts = parseMountOpts(args[argno+1])
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}
				for _, opt := range parsed_opts {
					err = checkOptIsForbidden(opt)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					checkOptIsMandatory(opt)
				}
				if config.Mode == "strict" {
					err = checkAllMandatoryOptsPresent()
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
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
	} else {
		action = "fallthrough"
	}

	if config.Mode == "relaxed" && action == "unknown" {
		action = "fallthrough"
	}

	var new_args []string

	if action == "fallthrough" {
		// Real UID and GID from syscall
		ruid := syscall.Getuid()
		rgid := syscall.Getgid()
		dropPrivileges(ruid, rgid)
		// exec fusermount3 with all args
		new_args = args
	}
	// FIXME: this could be simpler by calling checkParent before setting the action
	if action == "mount" || action == "unmount" {
		if config.Mode == "strict" {
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
	}
	if action == "mount" {
		newOptsString := strings.Join(parsed_opts, ",")
		new_args = []string{"-o", newOptsString, "--", mountpoint}
	}
	if action == "umount" {
		new_args = []string{"-u", mountpoint}
	}
	if action == "unknown" {
		fmt.Println("fusermount3 parameter not allowed")
		os.Exit(1)
	}

	cmd := exec.Command(config.Target, new_args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logger.Fatal().Msg(fmt.Sprintf("error running %s: %v", config.Target, err))
	}
}
