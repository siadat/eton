package main

import (
	"database/sql"
	"fmt"
	"github.com/andrew-d/go-termutil"
	"github.com/mattn/go-colorable"
	"github.com/mgutz/ansi"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
	"gopkg.in/fsnotify.v1"
)

var out = colorable.NewColorableStdout()

// attrStruct holds the data fetched from a row
// Only 1 ValueXxx field should have value, the others should be nil
type attrStruct struct {
	// Meta
	ID        sql.NullInt64
	ParentID  sql.NullInt64
	Name      sql.NullString
	Alias     sql.NullString
	Path      sql.NullString
	Frequency sql.NullInt64
	Mark      sql.NullInt64

	// Values
	ValueText sql.NullString
	ValueBlob []byte
	ValueInt  sql.NullInt64
	ValueReal sql.NullFloat64
	ValueTime time.Time

	// Timestamps
	CreatedAt  nullTime
	UpdatedAt  nullTime
	AccessedAt nullTime
	DeletedAt  nullTime
}

const sqlSelect = "id, value_text, name, parent_id, alias, mark, value_blob, created_at, updated_at"

// getID returns the int64 value of attr's ID.
func (attr attrStruct) getID() int64 {
	//var err error
	if value, err := attr.ID.Value(); err == nil && value != nil {
		return value.(int64)
	}
	return -1
}

// getCreatedAt returns created_at time
func (attr attrStruct) getCreatedAt() (t time.Time) {
	//var err error
	if value, err := attr.CreatedAt.Value(); err == nil && value != nil {
		t = value.(time.Time)
		return
	}
	return t
}

// getUpdatedAt returns updated_at time
func (attr attrStruct) getUpdatedAt() (t time.Time) {
	//var err error
	if value, err := attr.UpdatedAt.Value(); err == nil && value != nil {
		t = value.(time.Time)
		return
	}
	return t
}

// getAccessedAt returns accessed_at time
func (attr attrStruct) getAccessedAt() (t time.Time) {
	//var err error
	if value, err := attr.AccessedAt.Value(); err == nil && value != nil {
		t = value.(time.Time)
		return
	}
	return t
}

// getDeletedAt returns deleted_at time
func (attr attrStruct) getDeletedAt() (t time.Time) {
	//var err error
	if value, err := attr.DeletedAt.Value(); err == nil && value != nil {
		t = value.(time.Time)
		return
	}
	return t
}

// getIDString returns the string value of attr's ID.
func (attr attrStruct) getIDString() string {
	var err error
	if value, err := attr.ID.Value(); err == nil && value != nil {
		return strconv.Itoa(int(value.(int64)))
	}
	log.Fatal("attrStruct is not loaded, has no id")
	check(err)
	return ""
}

// getMark returns the int value of attr's mark
func (attr attrStruct) getMark() int {
	var err error
	if value, err := attr.Mark.Value(); err == nil && value != nil {
		return int(value.(int64))
	}
	log.Fatal("Mark is not loaded, has no 'mark'")
	check(err)
	return 0
}

// getIdentifier returns attr's ID, or its Alias if it is not nil.
func (attr attrStruct) getIdentifier() string {
	alias := attr.getAlias()
	if len(alias) > 0 {
		return alias
	}
	return attr.getIDString()
}

// getName is a helper to get attr's Name as string
func (attr attrStruct) getName() string {
	var err error

	if value, err := attr.Name.Value(); err == nil && value != nil {
		return value.(string)
	}
	log.Fatal("attrStruct is not loaded, has no name")
	check(err)
	return ""
}

// getAlias returns attr's alias
func (attr attrStruct) getAlias() string {
	var err error

	if value, err := attr.Alias.Value(); err == nil && value != nil {
		return value.(string)
	}
	check(err)
	return ""
}

// getTextValue returns a string representation of attr's value, whatever type it is
func (attr attrStruct) getTextValue() string {
	var err error
	if value, err := attr.ValueText.Value(); err == nil && value != nil {
		return value.(string)
	}
	check(err)
	return ""
}

// getValue returns a string representation of attr's value, in order of
// preference: first ValueBlob, then ValueText, then ValueInt, then ValueReal
func (attr attrStruct) getValue() string {
	var err error

	// if ValueBlov exists
	if len(attr.ValueBlob) > 0 {
		return string(attr.ValueBlob)
	}

	if value, err := attr.ValueText.Value(); err == nil && value != nil {
		return value.(string)
	}
	check(err)

	if value, err := attr.ValueInt.Value(); err == nil && value != nil {
		return strconv.Itoa(value.(int))
	}
	check(err)

	if value, err := attr.ValueReal.Value(); err == nil && value != nil {
		return strconv.FormatFloat(value.(float64), 'f', 2, 32)
	}
	check(err)

	log.Fatal("attrStruct is not loaded, has no value")

	return ""
}

// print pretty-prints attr's field values.
func (attr attrStruct) print(w *tabwriter.Writer, verbose bool, indent int, highlighteds []string, after int) {
	debug := false

	if debug {
		if value, err := attr.ParentID.Value(); err == nil && value != nil {
			fmt.Fprintf(w, "%s:%d\t", "ParentID", value)
		} else {
			fmt.Fprintf(w, "%s:%s\t", "ParentID", novalue)
		}

		if value, err := attr.Name.Value(); err == nil && value != nil {
			fmt.Fprintf(w, "%s:%s\t", "Name", value)
		} else {
			fmt.Fprintf(w, "%s:%s\t", "Name", novalue)
		}

		if value, err := attr.ValueText.Value(); err == nil && value != nil {
			fmt.Fprintf(w, "%s:%s\t", "ValueText", value)
		} else {
			fmt.Fprintf(w, "%s:%s\t", "ValueText", novalue)
		}

		if attr.ValueBlob != nil {
			fmt.Fprintf(w, "%s:%d\t", "ValueBlob-len", len(attr.ValueBlob))
		} else {
			fmt.Fprintf(w, "%s:%s\t", "ValueBlob-len", novalue)
		}
	} else {
		// Last modifier:
		//fmt.Fprintf(w, "%s\t", prettyAttr("at", attr.prettyAt()))

		// Name:
		//fmt.Fprintf(w, "%s\t", prettyAttr("name", attr.getName()))

		// Value:
		//fmt.Printf(strings.Repeat("      ", indent))

		if attr.getMark() == 0 {
			fmt.Fprintf(out, "[%s] %s\n", color(attr.getIdentifier(), "yellow+b"), attr.title())
		} else {
			fmt.Fprintf(out, "%s %s\n", color("("+attr.getIdentifier()+")", "black+b:white"), color(attr.title(), "default"))
		}
		if len(highlighteds) > 0 {
			fmt.Fprintln(out, attr.prettyMatches(highlighteds, after))
		}
	}
}

func (attr attrStruct) prettyMatches(highlighteds []string, after int) string {
	var valueText string
	if len(highlighteds) == 0 {
		valueText = attr.title()
	} else {
		valueText = strings.TrimSpace(attr.getValue())

		matchinglines := make([]string, 0, 0)

		lastMatchingLine := -1
		var matchCounter int
		for linenumber, line := range strings.Split(valueText, "\n") {
			line = strings.TrimSpace(line)
			isCoveredByLastMatch := lastMatchingLine != -1 && linenumber <= lastMatchingLine+after

			line, matched := highlightLine(line, highlighteds)
			if matched {
				lastMatchingLine = linenumber
				if true || !isCoveredByLastMatch {
					matchCounter++
				}
			}
			if matched || isCoveredByLastMatch {
				//prefix := fmt.Sprintf("%s L%s:", strings.Repeat(" ", len(attr.getIdentifier())), strconv.Itoa(linenumber+1))
				prefix := fmt.Sprintf("%s", strings.Repeat(" ", 3+len(attr.getIdentifier())))
				matchinglines = append(matchinglines, color(prefix, "black")+line)
				if maximumShownMatches != -1 && matchCounter >= maximumShownMatches {
					break
				}
			}
		}

		valueText = strings.Join(matchinglines, "\n")
	}
	return valueText + "\n"
}

func (attr attrStruct) title() string {
	valueText := strings.TrimSpace(attr.getTextValue())
	firstLineEndIndex := strings.Index(valueText, "\n")

	if firstLineEndIndex >= 0 {
		valueText = valueText[0:firstLineEndIndex]
	} else {
		if len(valueText) > 80 {
			valueText = valueText[0:80] + ellipsis
		}
	}
	return valueText
}

func (attr attrStruct) prettyAt() string {
	if attr.getUpdatedAt().IsZero() {
		return attr.getCreatedAt().Local().Format(datelayout) // + " "
	}
	return attr.getUpdatedAt().Local().Format(datelayout) // + "*"
}

func (attr attrStruct) prettyCreatedAt() string {
	return attr.getCreatedAt().Local().Format(datelayout)
}

func (attr attrStruct) prettyUpdatedAt() string {
	if !attr.getUpdatedAt().IsZero() {
		return attr.getUpdatedAt().Local().Format(datelayout)
	}
	return ""
}

func (attr attrStruct) filepath() string {
	f, err := ioutil.TempFile("", "eton-edit")
	check(err)
	f.Close()
	writeToFile(f.Name(), attr.getValue())
	return f.Name()
}

// setAlias sets attr's Alias to the given alias.
// If give alias is empty string, it will unset the alias (set it to NULL in the database).
func (attr attrStruct) setAlias(db *sql.DB, alias string) {

	unset := len(alias) == 0
	if !unset {
		var validAlias = regexp.MustCompile(`[^\s\d]+`)
		if !validAlias.MatchString(alias) {
			fmt.Fprintln(out, "Alias must contain a non-numeric character")
			return
		}
	}

	stmt, err := db.Prepare("UPDATE attributes SET alias = ? WHERE id = ?")
	check(err)

	//var result sql.Result
	if !unset {
		_, err = stmt.Exec(alias, attr.getID())
	} else {
		_, err = stmt.Exec(nil, attr.getID())
	}
	//check(err)
	if err == nil {
		if unset {
			fmt.Fprintf(out, "ID:%d unaliased\n", attr.getID())
		} else {
			fmt.Fprintf(out, "alias set: %s => %s\n", attr.getIdentifier(), alias)
		}
	} else {
		log.Fatalf("error while setting alias \"%s\" for ID:%d -- alias must be unique\n", alias, attr.getID()) // , err)
	}
	//rowsAffected, err := result.RowsAffected()
}

func (attr attrStruct) setMark(db *sql.DB, mark int) (rowsAffected int64) {
	stmt, err := db.Prepare("UPDATE attributes SET mark = ? WHERE id = ? AND deleted_at IS NULL")
	check(err)

	result, err := stmt.Exec(mark, attr.getID())
	check(err)
	rowsAffected, err = result.RowsAffected()
	check(err)

	return rowsAffected
}

func (attr attrStruct) rm(db *sql.DB) (rowsAffected int64) {
	stmt, err := db.Prepare("UPDATE attributes SET deleted_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL")
	check(err)

	result, err := stmt.Exec(attr.getID())
	check(err)
	rowsAffected, err = result.RowsAffected()
	check(err)

	return rowsAffected
}

func (attr attrStruct) unrm(db *sql.DB) (rowsAffected int64) {
	stmt, err := db.Prepare("UPDATE attributes SET deleted_at = NULL WHERE id = ? AND deleted_at IS NOT NULL")
	check(err)

	result, err := stmt.Exec(attr.getID())
	check(err)
	rowsAffected, err = result.RowsAffected()
	check(err)

	return rowsAffected
}

func (attr attrStruct) incrementFrequency(db *sql.DB) (rowsAffected int64) {
	stmt, err := db.Prepare("UPDATE attributes SET frequency = frequency + 1 WHERE id = ? AND deleted_at IS NULL")
	check(err)

	result, err := stmt.Exec(attr.getID())
	check(err)

	rowsAffected, err = result.RowsAffected()
	check(err)
	return
}

func (attr attrStruct) updateDb(db *sql.DB, valueText string) (rowsAffected int64) {
	updateStmt, err := db.Prepare("UPDATE attributes SET value_text = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?")
	check(err)

	result, err := updateStmt.Exec(valueText, attr.getID())
	check(err)
	rowsAffected, err = result.RowsAffected()
	check(err)
	return rowsAffected
}

func (attr attrStruct) edit(db *sql.DB) (rowsAffected int64) {
	filepath := attr.filepath()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					if (event.Name == filepath) {
						valueText := readFile(filepath)
						rowsAffected = attr.updateDb(db, valueText)
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			case <-done:
				// Edit the gofunction
				return
			}
		}
	}()

	err = watcher.Add("/tmp")
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		done <- true
	}()

	if openEditor(filepath) == false {
		return
	}

	valueText := readFile(filepath)

	if valueText != attr.getValue() {
		rowsAffected = attr.updateDb(db, valueText)
	}
	return
}

/******************************************************************************/

func writeToFile(filepath string, content string) {
	err := ioutil.WriteFile(filepath, []byte(content), 0644)
	check(err)
}

func highlightLine(line string, highlighteds []string) (string, bool) {
	if len(highlighteds) == 0 {
		return line, false
	}
	reFlags := "(?i)"

	quotedHighlighteds := make([]string, len(highlighteds), len(highlighteds))

	for i, str := range highlighteds {
		quotedHighlighteds[i] = regexp.QuoteMeta(str)
	}

	re := regexp.MustCompile(reFlags + "(" + strings.Join(quotedHighlighteds, "|") + ")")
	if indexes := re.FindStringIndex(line); indexes != nil {
		var indexBegin int
		var indexEnd int
		var beforeStr string
		var afterStr string

		if len(indexes) > 0 {
			firstIndex := indexes[0]
			indexBegin = firstIndex - 40
			if indexBegin < 0 {
				indexBegin = 0
			}
			if indexBegin != 0 {
				beforeStr = ellipsis
			}
			indexEnd = firstIndex + 40
		} else {
			indexEnd = 80
		}

		if indexEnd > indexBegin+80 {
			indexEnd = indexBegin + 80
		}
		if indexEnd > len(line) {
			indexEnd = len(line)
		}
		if indexEnd != len(line) {
			afterStr = ellipsis
		}

		line = re.ReplaceAllString(line[indexBegin:indexEnd], color("$0", "black+b:green"))
		return beforeStr + line + afterStr, true
	}
	return line, false
}

func prettyAttr(name, value string) string {
	if len(name) > 0 {
		name = name + ":"
	}
	if termutil.Isatty(os.Stdout.Fd()) {
		return ansi.Color(name, "black") + ansi.Color(value, "default")
	}
	return name + value
}

func prettyAttr2(name, value string) string {
	if termutil.Isatty(os.Stdout.Fd()) {
		return ansi.Color(name+":", "black") + ansi.Color(value, "blue")
	}
	return name + ":" + value
}

// color is the same as ansi.Color but only if STDOUT is a TTY
func color(str, color string) string {
	if termutil.Isatty(os.Stdout.Fd()) {
		return ansi.Color(str, color)
	}
	return str
}

func findAttributeByID(db *sql.DB, ID int64) (attr attrStruct) {
	var err error
	var stmt *sql.Stmt

	defer func() {
		if err == nil {
			attr.incrementFrequency(db)
		}
	}()

	stmt, err = db.Prepare("SELECT " + sqlSelect + " FROM attributes WHERE id = ? AND deleted_at IS NULL LIMIT 1")
	check(err)

	err = stmt.QueryRow(ID).Scan(&attr.ID, &attr.ValueText, &attr.Name, &attr.ParentID, &attr.Alias, &attr.Mark, &attr.ValueBlob, &attr.CreatedAt, &attr.UpdatedAt)
	if err != nil {
		// log.Fatalln("No record found with id", ID, err)
	}
	return
}

func findAttributeByAlias(db *sql.DB, alias string, exactMatchOnly bool) (attr attrStruct) {
	var err error
	var stmt *sql.Stmt

	defer func() {
		if err == nil {
			attr.incrementFrequency(db)
		}
	}()

	// Exact match
	stmt, err = db.Prepare("SELECT " + sqlSelect + "  FROM attributes WHERE alias = ? ORDER BY " + orderby + " LIMIT 1")
	check(err)
	err = stmt.QueryRow(alias).Scan(&attr.ID, &attr.ValueText, &attr.Name, &attr.ParentID, &attr.Alias, &attr.Mark, &attr.ValueBlob, &attr.CreatedAt, &attr.UpdatedAt)
	if err == nil {
		return
	}

	if exactMatchOnly {
		return
	}

	stmt, err = db.Prepare("SELECT " + sqlSelect + " FROM attributes WHERE alias LIKE ? ORDER BY " + orderby + " LIMIT 1")
	check(err)

	// Prefix match
	err = stmt.QueryRow(alias+"%").Scan(&attr.ID, &attr.ValueText, &attr.Name, &attr.ParentID, &attr.Alias, &attr.Mark, &attr.ValueBlob, &attr.CreatedAt, &attr.UpdatedAt)
	if err == nil {
		return
	}

	// Postfix match
	err = stmt.QueryRow("%"+alias).Scan(&attr.ID, &attr.ValueText, &attr.Name, &attr.ParentID, &attr.Alias, &attr.Mark, &attr.ValueBlob, &attr.CreatedAt, &attr.UpdatedAt)
	if err == nil {
		return
	}

	prunes := strings.Split(alias, "")

	// Fuzzy match
	err = stmt.QueryRow("%"+strings.Join(prunes, "%")+"%").Scan(&attr.ID, &attr.ValueText, &attr.Name, &attr.ParentID, &attr.Alias, &attr.Mark, &attr.ValueBlob, &attr.CreatedAt, &attr.UpdatedAt)
	if err == nil {
		return
	}

	return
}

func findAttributeByAliasOrID(db *sql.DB, indentifier string) (attr attrStruct) {
	attr = findAttributeByAlias(db, indentifier, false)
	if attr.getID() <= 0 {

		intID, err := strconv.Atoi(indentifier)

		if err != nil {
			return attr
		}

		attr = findAttributeByID(db, int64(intID))
	}

	return attr
}

func listWithFilters(db *sql.DB, opts options) (attrs []attrStruct) {
	var stmt *sql.Stmt
	var rows *sql.Rows
	var nolimit = opts.Limit == -1

	var sqlConditions string
	var sqlLimit string

	queryValues := make([]interface{}, 0, 5)

	if opts.IncludeRemoved {
		sqlConditions = "deleted_at IS NOT NULL"
	} else {
		sqlConditions = "deleted_at IS NULL"
	}

	if opts.RootID == -1 {
		sqlConditions += " AND parent_id IS NULL"
	} else {
		nolimit = true
		sqlConditions += fmt.Sprintf(" AND parent_id = %d ", opts.RootID)
	}

	if !opts.Recursive {
		sqlConditions += " AND parent_id IS NULL"
	}

	if opts.ShortMode {
		// sqlConditions += " AND ((alias IS NOT NULL AND alias != '') OR mark > 0)"
		sqlConditions += " AND mark > 0"
	}

	if opts.RootID == -1 && len(opts.Filters) > 0 {
		nolimit = true
		nameOrVal := make([]string, 0, 0)

		for _, filter := range opts.Filters {
			likeValue := "%" + filter + "%"
			// queryValues = append(queryValues, likeValue, likeValue, likeValue)
			queryValues = append(queryValues, likeValue, likeValue)

			// nameOrVal = append(nameOrVal, "(value_text LIKE ? OR name LIKE ? OR alias LIKE ?)")
			nameOrVal = append(nameOrVal, "(value_text LIKE ? OR alias LIKE ?)")
		}

		sqlConditions += " AND ( " + strings.Join(nameOrVal, " AND ") + " )"
	}

	if nolimit {
		sqlLimit = ""
	} else {
		queryValues = append(queryValues, opts.Offset)
		queryValues = append(queryValues, opts.Limit)
		sqlLimit = "LIMIT ?, ?"
	}

	// ===========================================================================

	tx, err := db.Begin()
	check(err)
	stmt, err = tx.Prepare("SELECT " + sqlSelect + " FROM attributes WHERE " + sqlConditions + " ORDER BY " + orderby + " " + sqlLimit)
	check(err)
	defer stmt.Close()
	rows, err = stmt.Query(queryValues...)
	check(err)
	defer rows.Close()

	attrs = make([]attrStruct, 0, 0)

	for rows.Next() {
		attr := attrStruct{}
		err = rows.Scan(&attr.ID, &attr.ValueText, &attr.Name, &attr.ParentID, &attr.Alias, &attr.Mark, &attr.ValueBlob, &attr.CreatedAt, &attr.UpdatedAt)
		check(err)
		attrs = append(attrs, attr)

		var optsNew options
		optsNew = opts
		optsNew.RootID = attr.getID()
		optsNew.Indent += 2
		//cmdLs(db, w, optsNew)
	}

	tx.Commit()
	return attrs
}

func getLastAttrID(db *sql.DB) int64 {
	// Experimental
	var ID int64

	stmt, err := db.Prepare("SELECT id FROM attributes WHERE deleted_at IS NULL ORDER BY " + orderby + " LIMIT 1")
	check(err)

	err = stmt.QueryRow().Scan(&ID)
	check(err)
	return ID
}

func saveString(db *sql.DB, valueText string) (lastInsertID int64) {
	stmt, err := db.Prepare("INSERT INTO attributes (name, value_text) VALUES ('note', ?)")
	check(err)

	result, err := stmt.Exec(valueText)
	check(err)

	lastInsertID, err = result.LastInsertId()
	check(err)

	return
}

func initializeDatabase(db *sql.DB) bool {
	// TODO use fts3 for faster full-text search: CREATE VIRTUAL TABLE attributes USING fts3 (...)
	sqlStmt := `
  DROP TABLE IF EXISTS attributes;
	CREATE TABLE attributes (
		id          INTEGER NOT NULL PRIMARY KEY,
		name        TEXT,
		alias       TEXT,
		parent_id   INTEGER,
		frequency   INTEGER DEFAULT 0,
		mark        INTEGER DEFAULT 0,
		-- pwd      TEXT,

		value_text  TEXT,
		value_blob  BLOB,
		value_int   INTEGER,
		value_real  REAL,
		value_time  DATETIME,

		accessed_at DATETIME,
		updated_at  DATETIME,
		deleted_at  DATETIME,
		created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

  CREATE UNIQUE INDEX IF NOT EXISTS index_on_alias        ON attributes (alias);
  CREATE        INDEX IF NOT EXISTS index_on_name         ON attributes (name);
  CREATE        INDEX IF NOT EXISTS index_on_value_text   ON attributes (value_text);
  CREATE        INDEX IF NOT EXISTS index_on_value_blob   ON attributes (value_blob);
  CREATE        INDEX IF NOT EXISTS index_on_value_int    ON attributes (value_int);
  CREATE        INDEX IF NOT EXISTS index_on_value_real   ON attributes (value_real);
  CREATE        INDEX IF NOT EXISTS index_on_accessed_at  ON attributes (accessed_at);
  CREATE        INDEX IF NOT EXISTS index_on_deleted_at   ON attributes (deleted_at);
  CREATE        INDEX IF NOT EXISTS index_on_frequency    ON attributes (frequency);
  CREATE        INDEX IF NOT EXISTS index_on_mark         ON attributes (mark);
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
		return false
	}
	fmt.Fprintln(out, "repository initiated")
	return true
}
