package main

import (
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"text/tabwriter"
)

var globalDB *sql.DB
var globalOpts Options

// const orderby = "-frequency, -mark, CASE WHEN updated_at IS NULL THEN created_at ELSE updated_at END DESC"
const orderby = "CASE WHEN updated_at IS NULL THEN created_at ELSE updated_at END DESC"
const defaultEditor = "vi"

func cmdShow(db *sql.DB, opts Options) bool {
	if len(opts.IDs) == 0 && len(opts.Aliases) == 0 {
		opts.IDs = append(opts.IDs, int64(getLastAttrID(db)))
	}

	for _, id := range opts.IDs {
		attr := findAttributeByID(db, id)
		//fmt.Printf(attr.GetValue())
		printToLess(attr.GetValue())
	}

	for _, alias := range opts.Aliases {
		attr := findAttributeByAlias(db, alias)
		//fmt.Printf(attr.GetValue())
		printToLess(attr.GetValue())
	}
	return true
}

func cmdCat(db *sql.DB, opts Options) bool {
	if len(opts.IDs) == 0 && len(opts.Aliases) == 0 {
		opts.IDs = append(opts.IDs, int64(getLastAttrID(db)))
	}

	for _, id := range opts.IDs {
		attr := findAttributeByID(db, id)
		fmt.Printf(attr.GetValue())
	}

	for _, alias := range opts.Aliases {
		attr := findAttributeByAlias(db, alias)
		fmt.Printf(attr.GetValue())
	}
	return true
}

func cmdMount(db *sql.DB, opts Options) bool {
	log.Println("Not implemented yet")
	/*
		globalDB = db
		globalOpts = opts
		globalOpts.Offset = 0
		globalOpts.Limit = 40
		Mount(opts.MountPoint)
	*/
	return true
}

func cmdAddFiles(db *sql.DB, files []string) bool {
	tx, err := db.Begin()

	// stmt, err := tx.Prepare("INSERT INTO attributes (name, pwd, value_text, value_blob) VALUES (?, ?, ?, ?)")
	stmt, err := tx.Prepare("INSERT INTO attributes (name, value_text, value_blob) VALUES (?, ?, ?)")

	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}

		fileAbsPath, err := filepath.Abs(file)
		// fileRelPath, err := filepath.Rel(pwd, fileAbsPath)

		if err != nil {
			log.Fatal(err)
		}

		//_, err = stmt.Exec("file", pwd, fileRelPath, content)
		_, err = stmt.Exec("file", fileAbsPath, content)
		if err != nil {
			log.Fatal(err)
		}
	}

	tx.Commit()
	return true
}

func cmdLs(db *sql.DB, w *tabwriter.Writer, opts Options) bool {
	attrs := listWithFilters(db, opts)
	for _, attr := range attrs {
		if opts.ListFilepaths {
			fmt.Println(attr.Filepath())
		} else {
			attr.Print(w, opts.Recursive, opts.Indent, opts.Filters, opts.AfterLinesCount)
		}
	}
	return true
}

func cmdNew(db *sql.DB, opts Options) bool {

	var value_text string

	if opts.FromStdin {
		lines := make([]string, 0, 0)
		reader := bufio.NewReader(os.Stdin)

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		go func() {
			for _ = range c {
				// CTRL-c
				// Do nothing, this will continue executing the rest of the code
			}
		}()

		for {
			line, _, err := reader.ReadLine()
			if err != nil {
				// EOF
				break
			}
			lines = append(lines, string(line))
			log.Printf("%s\n", prettyAttr("eton", string(line)))
		}
		value_text = strings.Join(lines, "\n")
	} else if len(opts.Note) > 0 {
		value_text = opts.Note
	} else {
		f, err := ioutil.TempFile("", "eton-edit")
		check(err)

		openEditor(f.Name())
		value_text = readFile(f.Name())
	}

	saveString(db, value_text)

	return true
}

func cmdAdd(db *sql.DB, id int, attrs []string) bool {
	// TODO
	return false
}

func cmdAddAttr(db *sql.DB, id int, attrs []string) bool {
	var stmt *sql.Stmt

	tx, err := db.Begin()

	if id == -1 {
		stmt, err = tx.Prepare("INSERT INTO attributes (name, value_text) VALUES (?, ?)")
	} else {
		stmt, err = tx.Prepare("INSERT INTO attributes (name, value_text, parent_id) VALUES (?, ?, ?)")
	}

	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()

	if err != nil {
		log.Fatal(err)
	}

	for _, attr := range attrs {
		name := ""
		value := ""

		nameValuePair := strings.SplitN(attr, ":", 2)

		switch len(nameValuePair) {
		case 1:
			value = nameValuePair[0]
		case 2:
			name = nameValuePair[0]
			value = nameValuePair[1]
		}

		if id == -1 {
			_, err = stmt.Exec(name, value)
		} else {
			_, err = stmt.Exec(name, value, id)
		}
		if err != nil {
			log.Fatal(err)
		}
	}

	tx.Commit()

	return true
}

func cmdUnalias(db *sql.DB, opts Options) bool {
	attr := findAttributeByAlias(db, opts.Alias)
	if attr.GetID() == -1 {
		log.Fatalf("alias \"%s\" not found", opts.Alias)
	} else {
		attr.SetAlias(db, "")
	}
	return true
}

func cmdAlias(db *sql.DB, opts Options) bool {
	if !(opts.ID > 0 && len(opts.Alias1) > 0 || len(opts.Alias2) > 0) && !(len(opts.Alias1) > 0 && len(opts.Alias2) > 0) {
		return false
	}

	var attr Attr

	if opts.ID > 0 {
		attr = findAttributeByID(db, opts.ID)
		if len(opts.Alias1) > 0 {
			attr.SetAlias(db, opts.Alias1)
		} else if len(opts.Alias2) > 0 {
			attr.SetAlias(db, opts.Alias2)
		}
	} else if len(opts.Alias1) > 0 && len(opts.Alias2) > 0 {
		attr1 := findAttributeByAlias(db, opts.Alias1)
		attr2 := findAttributeByAlias(db, opts.Alias2)

		if attr1.GetID() > 0 && attr2.GetID() <= 0 {
			attr1.SetAlias(db, opts.Alias2)
		} else if attr1.GetID() <= 0 && attr2.GetID() > 0 {
			attr2.SetAlias(db, opts.Alias1)
		} else {
			log.Println("not changing anything", attr1.GetID(), attr2.GetID())
		}
	}
	return true
}

func cmdEdit(db *sql.DB, opts Options) bool {
	var totalUpdated int64

	if len(opts.IDs) == 0 && len(opts.Aliases) == 0 {
		opts.IDs = append(opts.IDs, int64(getLastAttrID(db)))
	}

	for _, id := range opts.IDs {
		attr := findAttributeByID(db, id)
		totalUpdated += attr.Edit(db)
	}

	for _, alias := range opts.Aliases {
		attr := findAttributeByAlias(db, alias)
		totalUpdated += attr.Edit(db)
	}

	if totalUpdated > 0 {
		fmt.Println(totalUpdated, "records updated")
	}

	return true
}

func cmdRm(db *sql.DB, opts Options) bool {

	var totalUpdated int64

	for _, id := range opts.IDs {
		attr := findAttributeByID(db, id)
		totalUpdated += attr.Rm(db)
	}

	for _, alias := range opts.Aliases {
		attr := findAttributeByAlias(db, alias)
		totalUpdated += attr.Rm(db)
	}

	if totalUpdated > 0 {
		fmt.Println(totalUpdated, "deleted")
	}

	return true
}

func cmdUnrm(db *sql.DB, opts Options) bool {
	var totalUpdated int64

	for _, id := range opts.IDs {
		attr := findAttributeByID(db, id)
		totalUpdated += attr.Unrm(db)
	}

	for _, alias := range opts.Aliases {
		attr := findAttributeByAlias(db, alias)
		totalUpdated += attr.Unrm(db)
	}

	if totalUpdated > 0 {
		fmt.Println(totalUpdated, "recovered")
	}

	return true
}

func cmdInit(db *sql.DB) bool {
	InitializeDatabase(db)
	return true
}

func cmdMark(db *sql.DB, opts Options) bool {
	var totalUpdated int64
	for _, id := range opts.IDs {
		attr := findAttributeByID(db, id)
		totalUpdated += attr.SetMark(db, 1)
	}

	for _, alias := range opts.Aliases {
		attr := findAttributeByAlias(db, alias)
		totalUpdated += attr.SetMark(db, 1)
	}

	fmt.Println(totalUpdated, "marked")
	return true
}

func cmdUnmark(db *sql.DB, opts Options) bool {
	var totalUpdated int64
	for _, id := range opts.IDs {
		attr := findAttributeByID(db, id)
		totalUpdated += attr.SetMark(db, 0)
	}

	for _, alias := range opts.Aliases {
		attr := findAttributeByAlias(db, alias)
		totalUpdated += attr.SetMark(db, 0)
	}

	fmt.Println(totalUpdated, "marked")
	return true
}

/******************************************************************************/

func openEditor(filepath string) {
	var cmd *exec.Cmd

	editor := os.Getenv("EDITOR")

	if len(editor) > 0 {
		cmd = exec.Command("/usr/bin/env", editor, filepath)
	} else {
		if _, err := os.Stat("/usr/bin/sensible-editor"); err == nil {
			cmd = exec.Command("/usr/bin/sensible-editor", filepath)
		} else {
			cmd = exec.Command("/usr/bin/env", defaultEditor, filepath)
		}
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func readFile(filepath string) string {
	data, err := ioutil.ReadFile(filepath)
	check(err)
	return string(data)
}

func check(e error) {
	if e != nil {
		// log.Fatal(e)
		panic(e)
	}
}

func printToLess(text string) {
	// declare your pager
	cmd := exec.Command("/usr/bin/env", "less")
	// create a pipe (blocking)
	r, stdin := io.Pipe()
	// Set your i/o's
	cmd.Stdin = r
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Create a blocking chan, Run the pager and unblock once it is finished
	c := make(chan struct{})
	go func() {
		defer close(c)
		cmd.Run()
	}()

	// Pass anything to your pipe
	fmt.Fprintf(stdin, text)

	// Close stdin (result in pager to exit)
	stdin.Close()

	// Wait for the pager to be finished
	<-c
}
