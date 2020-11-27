// This code is released under the MIT License
// Copyright (c) 2020 Pix4D and the terravalet contributors.

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/integrii/flaggy"
)

var (
	// Filled by the linker.
	fullVersion  = "unknown" // example: v0.0.9-8-g941583d027-dirty
	shortVersion = "unknown" // example: v0.0.9
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	planPath := ""
	upPath := ""
	downPath := ""
	localStatePath := "local.tfstate"

	flaggy.ResetParser() // flaggy keeps gobal state; workaround for testing :-(
	flaggy.SetDescription("A simple valet for terraform operations (WIP).")
	flaggy.String(&planPath, "", "plan", "Path to the output of terraform plan.")
	flaggy.String(&upPath, "", "up", "Path to the up migration script to generate (NNN_TITLE.up.sh).")
	flaggy.String(&downPath, "", "down", "Path to the down migration script to generate (NNN_TITLE.down.sh).")
	flaggy.String(&localStatePath, "", "local-state", "Path to the local state to modify (both src and dst).")

	flaggy.SetVersion(fullVersion)
	flaggy.ParseArgs(args) // This might call os.Exit() :-/

	if planPath == "" {
		return fmt.Errorf("missing value for -plan")
	}
	if upPath == "" {
		return fmt.Errorf("missing value for -up")
	}
	if downPath == "" {
		return fmt.Errorf("missing value for -down")
	}

	planFile, err := os.Open(planPath)
	if err != nil {
		return fmt.Errorf("opening the terraform plan file: %v", err)
	}
	defer planFile.Close()

	upFile, err := os.Create(upPath)
	if err != nil {
		return fmt.Errorf("creating the up file: %v", err)
	}
	defer upFile.Close()

	downFile, err := os.Create(downPath)
	if err != nil {
		return fmt.Errorf("creating the down file: %v", err)
	}
	defer downFile.Close()

	create, destroy, err := parse(planFile)
	if err != nil {
		return fmt.Errorf("parse: %v", err)
	}

	upMatches, downMatches, err := match(create, destroy)
	if err != nil {
		return fmt.Errorf("match: %v", err)
	}

	if err := script(upMatches, localStatePath, upFile); err != nil {
		return fmt.Errorf("writing the up script: %v", err)
	}

	if err := script(downMatches, localStatePath, downFile); err != nil {
		return fmt.Errorf("writing the down script: %v", err)
	}

	return nil
}

// Parse the output of "terraform plan" and return a list of elements to be created and a
// list of elements to be destroyed. The two lists are unordered.
//
// For example:
// " # module.ci.aws_instance.docker will be destroyed"
// " # aws_instance.docker will be created"
// " # module.ci.module.workers["windows-vs2019"].aws_autoscaling_schedule.night_mode will be destroyed"
// " # module.workers["windows-vs2019"].aws_autoscaling_schedule.night_mode will be created"
func parse(rd io.Reader) ([]string, []string, error) {
	var re = regexp.MustCompile(`# (.+) will be (.+)`)

	var create []string
	var destroy []string

	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		line := scanner.Text()
		if m := re.FindStringSubmatch(line); m != nil {
			if len(m) != 3 {
				return create, destroy,
					fmt.Errorf("could not parse line %q: %q", line, m)
			}
			switch m[2] {
			case "created":
				create = append(create, m[1])
			case "destroyed":
				destroy = append(destroy, m[1])
			case "read during apply":
				// do nothing
			default:
				return create, destroy,
					fmt.Errorf("line %q, unexpected action %q", line, m[2])
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return create, destroy, err
	}

	return create, destroy, nil
}

// Given two unordered lists create and destroy, return a map that matches each old
// element in destroy to the corresponding new element in create.
// The criterium used to perform a match is that one of the two elements must be a
// prefix of the other. Note that the longest element could be the old or the new one,
// it depends.
func match(create, destroy []string) (map[string]string, map[string]string, error) {
	// old -> new (or equvalenty: destroy -> create)
	upMatches := map[string]string{}
	downMatches := map[string]string{}

	// 1. Create and destroy give us the direction:
	//    terraform state mv destroy[i] create[j]
	// 2. But, for each resource, we need to know i,j so that we can match which old state
	//    we want to move to which new state, for example both are theoretically valid:
	// 	    terraform state mv module.ci.aws_instance.docker           aws_instance.docker
	//      terraform state mv           aws_instance.docker module.ci.aws_instance.docker

	for _, d := range destroy {
		for _, c := range create {
			if strings.HasSuffix(c, d) || strings.HasSuffix(d, c) {
				upMatches[d] = c
				downMatches[c] = d
				break
			}
		}
	}
	return upMatches, downMatches, nil
}

// Given a map old->new, create a script that for each element in the map issues the
// command: "terraform state mv old new".
func script(matches map[string]string, statePath string, out io.Writer) error {
	fmt.Fprintf(out, "#! /usr/bin/sh\n")
	fmt.Fprintf(out, "# DO NOT EDIT. Generated by terravalet.\n")
	fmt.Fprintf(out, "# terravalet_output_format=2\n")
	fmt.Fprintf(out, "#\n")
	fmt.Fprintf(out, "# This script will move %d items.\n\n", len(matches))
	fmt.Fprintf(out, "set -e\n\n")

	// -lock=false greatly speeds up operations when the state has many elements
	// and is safe as long as we use -state=FILE, since this keeps operations
	// strictly local, without considering the configured backend.
	cmd := fmt.Sprintf("terraform state mv -lock=false -state=%s", statePath)

	// Go maps are unordered. We want instead a stable iteration order, to make it
	// possible to compare scripts.
	destroys := make([]string, 0, len(matches))
	for d := range matches {
		destroys = append(destroys, d)
	}
	sort.Strings(destroys)

	i := 1
	for _, d := range destroys {
		fmt.Fprintf(out, "%s \\\n    '%s' \\\n    '%s'\n\n", cmd, d, matches[d])
		i++
	}
	return nil
}
