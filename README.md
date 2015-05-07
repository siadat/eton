# eton

eton is a note-taking cli tool.

## Install / Upgrade

    go get -u github.com/siadat/eton

## Examples

```shell
# display the help message
eton -h

# quick add
eton new '[ ] do something'

# create a new note and open $EDITOR to edit it
eton new

# edit the last item
eton edit

# add text from STDIN
ps aux | eton new -

# add a file
eton addfile file.txt

# a unique alias can be set and used instead of the numeric id
eton alias 2 all-processes

# the order of id and alias does not matter
eton alias all-processes 2

# you can rename an alias
eton alias all-processes processes

# list all items
eton ls -Lall

# filter items containing words "thing" AND "some"
eton ls thing some

# you can mark specific items
eton mark processes 1

# only list marked items (short mode)
eton ls -s

# open an item in less
eton show processes

# edit items
eton edit {1..3} 4 prcs

# alias matching is fuzzy for cat, show, edit, mark, unmark commands
eton cat prcs

# pass items to xargs as filenames:
eton ls something -l |xargs -i less {}
```

Notes are stored in ~/.etondb

```shell
echo 'SELECT * FROM attributes LIMIT 10;' |sqlite3 ~/.etondb
```

Set `$EDITOR` environment variable to edit notes in your prefered editor, e.g., `export EDITOR=vim`.

I would love to hear how you use eton. Make pull requests, report bugs, suggest ideas.
