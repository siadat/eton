package main

import (
	"database/sql"
	"database/sql/driver"
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
)

var out = colorable.NewColorableStdout()

// Attr holds the data fetched from a row
// Only 1 ValueXxx field should have value, the others should be nil
type Attr struct {
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
	CreatedAt  NullTime
	UpdatedAt  NullTime
	AccessedAt NullTime
	DeletedAt  NullTime
}

const sqlSelect = "id, value_text, name, parent_id, alias, mark, value_blob, created_at, updated_at"

// GetID returns the int64 value of attr's ID.
func (attr Attr) GetID() int64 {
	//var err error
	if value, err := attr.ID.Value(); err == nil && value != nil {
		return value.(int64)
	}
	return -1
}

func (attr Attr) GetCreatedAt() (t time.Time) {
	//var err error
	if value, err := attr.CreatedAt.Value(); err == nil && value != nil {
		t = value.(time.Time)
		return
	}
	return t
}

func (attr Attr) GetUpdatedAt() (t time.Time) {
	//var err error
	if value, err := attr.UpdatedAt.Value(); err == nil && value != nil {
		t = value.(time.Time)
		return
	}
	return t
}

func (attr Attr) GetAccessedAt() (t time.Time) {
	//var err error
	if value, err := attr.AccessedAt.Value(); err == nil && value != nil {
		t = value.(time.Time)
		return
	}
	return t
}

func (attr Attr) GetDeletedAt() (t time.Time) {
	//var err error
	if value, err := attr.DeletedAt.Value(); err == nil && value != nil {
		t = value.(time.Time)
		return
	}
	return t
}

// GetIDString returns the string value of attr's ID.
func (attr Attr) GetIDString() string {
	var err error
	if value, err := attr.ID.Value(); err == nil && value != nil {
		return strconv.Itoa(int(value.(int64)))
	}
	log.Fatal("Attr is not loaded, has no id")
	check(err)
	return ""
}

// GetMark returns the int value of attr's mark
func (attr Attr) GetMark() int {
	var err error
	if value, err := attr.Mark.Value(); err == nil && value != nil {
		return int(value.(int64))
	}
	log.Fatal("Mark is not loaded, has no 'mark'")
	check(err)
	return 0
}

// GetIdentifier returns attr's ID, or its Alias if it is not nil.
func (attr Attr) GetIdentifier() string {
	alias := attr.GetAlias()
	if len(alias) > 0 {
		return alias
	} else {
		return attr.GetIDString()
	}
}

// GetName is a helper to get attr's Name as string
func (attr Attr) GetName() string {
	var err error

	if value, err := attr.Name.Value(); err == nil && value != nil {
		return value.(string)
	}
	log.Fatal("Attr is not loaded, has no name")
	check(err)
	return ""
}

// GetName is a helper to get attr's Name as string
func (attr Attr) GetAlias() string {
	var err error

	if value, err := attr.Alias.Value(); err == nil && value != nil {
		return value.(string)
	}
	check(err)
	return ""
}

// GetTextValue returns a string representation of attr's value, whatever type it is
func (attr Attr) GetTextValue() string {
	var err error
	if value, err := attr.ValueText.Value(); err == nil && value != nil {
		return value.(string)
	}
	check(err)
	return ""
}

// GetValue returns a string representation of attr's value, in order of
// preference: first ValueBlob, then ValueText, then ValueInt, then ValueReal
func (attr Attr) GetValue() string {
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

	log.Fatal("Attr is not loaded, has no value")

	return ""
}

// Print pretty-prints attr's field values.
func (attr Attr) Print(w *tabwriter.Writer, verbose bool, indent int, highlighteds []string, after int) {
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
		//fmt.Fprintf(w, "%s\t", prettyAttr("name", attr.GetName()))

		// Value:
		//fmt.Printf(strings.Repeat("      ", indent))

		if attr.GetMark() == 0 {
			fmt.Fprintf(out, "[%s] %s\n", Color(attr.GetIdentifier(), "yellow+b"), attr.Title())
		} else {
			fmt.Fprintf(out, "%s %s\n", Color("("+attr.GetIdentifier()+")", "black+b:white"), Color(attr.Title(), "default"))
		}
		if len(highlighteds) > 0 {
			fmt.Fprintln(out, attr.PrettyMatches(highlighteds, after))
		}
	}
}

func (attr Attr) PrettyMatches(highlighteds []string, after int) string {
	var valueText string
	if len(highlighteds) == 0 {
		valueText = attr.Title()
	} else {
		valueText = strings.TrimSpace(attr.GetValue())

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
					matchCounter += 1
				}
			}
			if matched || isCoveredByLastMatch {
				//prefix := fmt.Sprintf("%s L%s:", strings.Repeat(" ", len(attr.GetIdentifier())), strconv.Itoa(linenumber+1))
				prefix := fmt.Sprintf("%s", strings.Repeat(" ", 3+len(attr.GetIdentifier())))
				matchinglines = append(matchinglines, Color(prefix, "black")+line)
				if maximumShownMatches != -1 && matchCounter >= maximumShownMatches {
					break
				}
			}
		}

		valueText = strings.Join(matchinglines, "\n")
	}
	return valueText + "\n"
}

func (attr Attr) Title() string {
	valueText := strings.TrimSpace(attr.GetTextValue())
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

func (attr Attr) prettyAt() string {
	if attr.GetUpdatedAt().IsZero() {
		return attr.GetCreatedAt().Local().Format(datelayout) // + " "
	} else {
		return attr.GetUpdatedAt().Local().Format(datelayout) // + "*"
	}
}

func (attr Attr) prettyCreatedAt() string {
	return attr.GetCreatedAt().Local().Format(datelayout)
}

func (attr Attr) prettyUpdatedAt() string {
	if !attr.GetUpdatedAt().IsZero() {
		return attr.GetUpdatedAt().Local().Format(datelayout)
	} else {
		return ""
	}
}

func (attr Attr) Filepath() string {
	f, err := ioutil.TempFile("", "eton-edit")
	check(err)
	f.Close()
	writeToFile(f.Name(), attr.GetValue())
	return f.Name()
}

// SetAlias sets attr's Alias to the given alias.
// If give alias is empty string, it will unset the alias (set it to NULL in the database).
func (attr Attr) SetAlias(db *sql.DB, alias string) {

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
		_, err = stmt.Exec(alias, attr.GetID())
	} else {
		_, err = stmt.Exec(nil, attr.GetID())
	}
	//check(err)
	if err == nil {
		if unset {
			fmt.Fprintf(out, "ID:%d unaliased\n", attr.GetID())
		} else {
			fmt.Fprintf(out, "alias set: %s => %s\n", attr.GetIdentifier(), alias)
		}
	} else {
		log.Fatalf("error while setting alias \"%s\" for ID:%d -- alias must be unique\n", alias, attr.GetID()) // , err)
	}
	//rowsAffected, err := result.RowsAffected()
}

func (attr Attr) SetMark(db *sql.DB, mark int) (rowsAffected int64) {
	stmt, err := db.Prepare("UPDATE attributes SET mark = ? WHERE id = ? AND deleted_at IS NULL")
	check(err)

	result, err := stmt.Exec(mark, attr.GetID())
	check(err)
	rowsAffected, err = result.RowsAffected()
	check(err)

	return rowsAffected
}

func (attr Attr) Rm(db *sql.DB) (rowsAffected int64) {
	stmt, err := db.Prepare("UPDATE attributes SET deleted_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL")
	check(err)

	result, err := stmt.Exec(attr.GetID())
	check(err)
	rowsAffected, err = result.RowsAffected()
	check(err)

	return rowsAffected
}

func (attr Attr) Unrm(db *sql.DB) (rowsAffected int64) {
	stmt, err := db.Prepare("UPDATE attributes SET deleted_at = NULL WHERE id = ? AND deleted_at IS NOT NULL")
	check(err)

	result, err := stmt.Exec(attr.GetID())
	check(err)
	rowsAffected, err = result.RowsAffected()
	check(err)

	return rowsAffected
}

func (attr Attr) IncrementFrequency(db *sql.DB) (rowsAffected int64) {
	stmt, err := db.Prepare("UPDATE attributes SET frequency = frequency + 1 WHERE id = ? AND deleted_at IS NULL")
	check(err)

	result, err := stmt.Exec(attr.GetID())
	check(err)

	rowsAffected, err = result.RowsAffected()
	check(err)
	return
}

func (attr Attr) Edit(db *sql.DB) (rowsAffected int64) {
	filepath := attr.Filepath()

	if openEditor(filepath) == false {
		return
	}

	value_text := readFile(filepath)

	if value_text != attr.GetValue() {
		update_stmt, err := db.Prepare("UPDATE attributes SET value_text = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?")
		check(err)

		result, err := update_stmt.Exec(value_text, attr.GetID())
		check(err)
		rowsAffected, err = result.RowsAffected()
		check(err)
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
	} else {
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

			line = re.ReplaceAllString(line[indexBegin:indexEnd], Color("$0", "black+b:green"))
			return beforeStr + line + afterStr, true
		}
		return line, false
	}
}

func prettyAttr(name, value string) string {
	if len(name) > 0 {
		name = name + ":"
	}
	if termutil.Isatty(os.Stdout.Fd()) {
		return ansi.Color(name, "black") + ansi.Color(value, "default")
	} else {
		return name + value
	}
}

func prettyAttr2(name, value string) string {
	if termutil.Isatty(os.Stdout.Fd()) {
		return ansi.Color(name+":", "black") + ansi.Color(value, "blue")
	} else {
		return name + ":" + value
	}
}

// Color is the same as ansi.Color but only if STDOUT is a TTY
func Color(str, color string) string {
	if termutil.Isatty(os.Stdout.Fd()) {
		return ansi.Color(str, color)
	} else {
		return str
	}
}

func findAttributeByID(db *sql.DB, ID int64) (attr Attr) {
	var err error
	var stmt *sql.Stmt

	defer func() {
		if err == nil {
			attr.IncrementFrequency(db)
		}
	}()

	stmt, err = db.Prepare("SELECT " + sqlSelect + " FROM attributes WHERE id = ? AND deleted_at IS NULL LIMIT 1")
	check(err)

	err = stmt.QueryRow(ID).Scan(&attr.ID, &attr.ValueText, &attr.Name, &attr.ParentID, &attr.Alias, &attr.Mark, &attr.ValueBlob, &attr.CreatedAt, &attr.UpdatedAt)
	if err != nil {
		log.Fatalln("No record found with id", ID, err)
	}
	return
}

func findAttributeByAlias(db *sql.DB, alias string, exactMatchOnly bool) (attr Attr) {
	var err error
	var stmt *sql.Stmt

	defer func() {
		if err == nil {
			attr.IncrementFrequency(db)
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

func findAttributeByAliasOrID(db *sql.DB, alias_or_id string) (attr Attr) {
	attr = findAttributeByAlias(db, alias_or_id, false)
	if attr.GetID() <= 0 {

		intID, err := strconv.Atoi(alias_or_id)

		if err != nil {
			return attr
		}

		attr = findAttributeByID(db, int64(intID))
	}

	return attr
}

func listWithFilters(db *sql.DB, opts Options) (attrs []Attr) {
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

	attrs = make([]Attr, 0, 0)

	for rows.Next() {
		attr := Attr{}
		err = rows.Scan(&attr.ID, &attr.ValueText, &attr.Name, &attr.ParentID, &attr.Alias, &attr.Mark, &attr.ValueBlob, &attr.CreatedAt, &attr.UpdatedAt)
		check(err)
		attrs = append(attrs, attr)

		var optsNew Options
		optsNew = opts
		optsNew.RootID = attr.GetID()
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

func saveString(db *sql.DB, value_text string) (lastInsertId int64) {
	new_stmt, err := db.Prepare("INSERT INTO attributes (name, value_text) VALUES ('note', ?)")
	check(err)

	result, err := new_stmt.Exec(value_text)
	check(err)

	lastInsertId, err = result.LastInsertId()
	check(err)

	return
}

func InitializeDatabase(db *sql.DB) bool {
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

// NullTime allows timestamps to be NULL
type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

// Scan implements the Scanner interface.
func (nt *NullTime) Scan(value interface{}) error {
	nt.Time, nt.Valid = value.(time.Time)
	return nil
}

// Value implements the driver Valuer interface.
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}
