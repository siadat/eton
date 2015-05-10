# eton

eton is a note-taking cli tool.

  * sqlite database storage
  * fast search
  * quick and direct access to a note using a unique alias
  * mark important notes

## Install or upgrade

    go get -u github.com/siadat/eton

## Usage examples

```shell
# display the help message
eton -h
```

### new

```shell
# quick add
eton new 'eton is simple'
eton new 'https://...'
eton new '[ ] try eton'

# create a new note using $EDITOR
eton new

# add from STDIN
ps aux |eton new -

# add files
eton addfile file1.txt file2.txt
find -type f |eton addfile -
```

### edit

```shell
# edit last item
eton edit

# edit items
eton edit processes 1
```

### alias

```shell
# set a unique alias
eton alias 3 procs

# use the alias, so you don't need to remember its id
eton cat procs
eton edit procs
eton show procs

# rename an alias
eton alias procs processes

# remove an alias
eton unalias processes
```

### mark

```shell
# mark an item
eton mark processes

# unmark an item
eton unmark processes
```

### ls

```shell
# list recent items
eton ls

# filter items containing words "eton" AND "simple"
eton ls eton simple

# list all items
eton ls -a

# only list marked items (short mode)
eton ls -s

# pass items to xargs as filenames:
eton ls '[ ]' -l |xargs -i less {}
```

### more

```shell
# alias matching is fuzzy for these commands: cat, show, edit, mark, unmark
eton cat prs

# view items 1, 2, and 3 using less
eton show 1 2 3

# Notes are stored in `~/.etondb`
echo 'SELECT * FROM attributes LIMIT 10;' |sqlite3 ~/.etondb
```

Set `$EDITOR` environment variable to edit notes in your prefered editor, e.g., `export EDITOR=vim`.

I would love to hear how you use eton. Make pull requests, report bugs, suggest ideas.
