package main

import (
	"os/user"
	"path/filepath"
	"strconv"
)

const (
	novalue         = "nil"
	datelayout      = "06/01/02 03:04pm"
	ellipsis        = "…"
	maxShownMatches = -1
)

type options struct {
	ID              int64
	Alias           string
	IDs             []int64
	Aliases         []string
	Limit           int
	Offset          int
	RootID          int64
	Indent          int
	Filters         []string
	FromStdin       bool
	Recursive       bool
	IncludeRemoved  bool
	ShortMode       bool
	Verbose         bool
	ListFilepaths   bool
	MountPoint      string
	Note            string
	AfterLinesCount int
	Alias1          string
	Alias2          string
}

func optionsFromArgs(args map[string]interface{}) (opts options) {
	// log.Printf("%v\n", args)
	var err error

	opts.RootID = -1
	opts.Indent = 0
	opts.ListFilepaths = args["--list-files"].(bool)

	if args["<note>"] != nil {
		opts.Note = args["<note>"].(string)
	}

	opts.Offset, err = strconv.Atoi(args["--offset"].(string))
	check(err)

	opts.AfterLinesCount, err = strconv.Atoi(args["--after"].(string))
	check(err)

	if args["--all"].(bool) || args["--limit"].(string) == "all" {
		opts.Limit = -1
	} else {
		opts.Limit, err = strconv.Atoi(args["--limit"].(string))
		check(err)
	}

	if args["<id1>"] != nil {
		intID, err := strconv.Atoi(args["<id1>"].(string))
		if err == nil {
			opts.ID = int64(intID)
		} else {
			opts.Alias1 = args["<id1>"].(string)
		}
	}

	if args["<mountpoint>"] != nil {
		opts.MountPoint = args["<mountpoint>"].(string)
	} else {
		opts.MountPoint = filepath.Join(homeDir(), "eton-default-mount-point")
	}

	if args["<id2>"] != nil {
		intID, err := strconv.Atoi(args["<id2>"].(string))
		if err == nil {
			opts.ID = int64(intID)
		} else {
			opts.Alias2 = args["<id2>"].(string)
		}
	}

	if args["<id>"] != nil {
		intID, err := strconv.Atoi(args["<id>"].(string))
		if err == nil {
			opts.ID = int64(intID)
		} else {
			opts.Alias = args["<id>"].(string)
		}
	}

	for _, id := range args["<ids>"].([]string) {
		intID, err := strconv.Atoi(id)
		if err == nil {
			opts.IDs = append(opts.IDs, int64(intID))
		} else {
			opts.Aliases = append(opts.Aliases, id)
		}
	}

	if args["<alias>"] != nil {
		opts.Alias = args["<alias>"].(string)
	}

	opts.Filters = args["<filters>"].([]string)
	opts.FromStdin = args["-"].(bool)
	opts.Recursive = false // args["--recursive"].(bool)
	opts.IncludeRemoved = args["--removed"].(bool)
	opts.ShortMode = args["--short"].(bool)
	opts.Verbose = args["--verbose"].(bool)
	return opts
}

func (opts options) getIDsArrayOfInterface() []interface{} {
	var interfaceIds = make([]interface{}, len(opts.IDs), len(opts.IDs))
	for i, id := range opts.IDs {
		interfaceIds[i] = id
	}
	return interfaceIds
}

func homeDir() string {
	usr, err := user.Current()
	check(err)
	return usr.HomeDir
}
