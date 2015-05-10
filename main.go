package main

import (
	"bufio"
	"database/sql"
	"github.com/docopt/docopt-go"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var pwd string

const dbfilename string = ".etondb"

const usage string = `Usage:
    eton new [-|<note>] [-v]
    eton (ls|grep) [<filters>...] [-asl] [-o OFFSET] [-L LIMIT] [--after AFTER] [--removed]
    eton edit [<ids>...] [-v]
    eton alias <id1> <id2>
    eton unalias <alias>
    eton mark <ids>...
    eton unmark <ids>...
    eton cat [<ids>...]
    eton show [<ids>...]
    eton (rm|remove) <ids>...
    eton (unrm|unremove|recover) <ids>...
    eton addfile (-|<file>...)
    eton mount [<mountpoint>]

Options:
    -A, --after AFTER    lines to print after a match [default: 0]
    -o, --offset OFFSET  offset for the items listed [default: 0]
    -L, --limit LIMIT    maximum number of rows returned, pass -Lall to list everything [default: 10]
    -r, --recursive      recursive mode
    -l, --list-files     list items as filenames
    -s, --short          short mode lists rows with aliases only
    -v, --verbose        talk a lot
    -a, --all            list all items, alias for --limit -1
    --removed            only removed items
`

func main() {
	args, err := docopt.Parse(usage, nil, true, "version 0.0.0", false, true)
	check(err)

	opts := optionsFromArgs(args)

	//pwd, _ = os.Getwd()

	dbfile := filepath.Join(homeDirectory(), dbfilename)
	var db *sql.DB

	dbfileExists := false

	if _, err = os.Stat(dbfile); err == nil {
		dbfileExists = true
	}

	//if dbfileExists || args["init"].(bool) {
	if true {
		var err error
		db, err = sql.Open("sqlite3", dbfile)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
	} else {
		log.Fatal(`database file not found, use "init" command`)
	}

	if !dbfileExists {
		cmdInit(db)
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 0, 2, ' ', 0)

	switch true {
	// case args["init"].(bool):
	// 	if dbfileExists {
	// 		log.Fatal("database already exists, command ignored.")
	// 	}
	// 	cmdInit(db)
	case args["mount"].(bool):
		cmdMount(db, opts)
	case args["new"].(bool):
		cmdNew(db, opts)
	case args["addfile"].(bool):
		if len(args["<file>"].([]string)) > 0 {
			cmdAddFiles(db, args["<file>"].([]string))
		} else {
			reader := bufio.NewReader(os.Stdin)
			for {
				line, _, err := reader.ReadLine()
				if err != nil {
					break
				}
				sline := string(line)
				cmdAddFiles(db, []string{sline})
			}
		}
	case args["ls"].(bool) || args["grep"].(bool):
		cmdLs(db, w, opts)
	case args["cat"].(bool):
		cmdCat(db, opts)
	case args["show"].(bool):
		cmdShow(db, opts)
	case args["rm"].(bool) || args["remove"].(bool):
		cmdRm(db, opts)
	case args["unrm"].(bool) || args["unremove"].(bool) || args["recover"].(bool):
		cmdUnrm(db, opts)
	case args["edit"].(bool):
		cmdEdit(db, opts)
	case args["mark"].(bool):
		cmdMark(db, opts)
	case args["unmark"].(bool):
		cmdUnmark(db, opts)
	case args["alias"].(bool):
		cmdAlias(db, opts)
	case args["unalias"].(bool):
		cmdUnalias(db, opts)
	case args["addattr"].(bool):
		id, _ := strconv.Atoi(args["<id>"].(string))
		cmdAddAttr(db, id, args["<filters>"].([]string))
	default:
		log.Println("Never reached")
	}

	//w.Flush()
}

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
