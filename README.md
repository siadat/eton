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

# quick add
eton new '[ ] try eton'

# create a new note and open $EDITOR to edit it
eton new

# edit the last item
eton edit

# add text from STDIN
ps aux | eton new -

# add a file
eton addfile file.txt

# use the alias instead of its id
eton show 2

# set a unique alias
eton alias 2 procs

# use the alias, so you don't need to remember its id
eton cat procs

# you can rename an alias
eton alias procs processes

# list all items
eton ls -a

# filter items containing words "word1" AND "word2"
eton ls word1 word2

# you can mark specific items
eton mark processes 1

# only list marked items (short mode)
eton ls -s

# edit items
eton edit processes 1

# alias matching is fuzzy for these commands: cat, show, edit, mark, unmark
eton cat prcs

# pass items to xargs as filenames:
eton ls '[ ]' -l |xargs -i less {}
```

Notes are stored in `~/.etondb`

```shell
echo 'SELECT * FROM attributes LIMIT 10;' |sqlite3 ~/.etondb
```

Set `$EDITOR` environment variable to edit notes in your prefered editor, e.g., `export EDITOR=vim`.

I would love to hear how you use eton. Make pull requests, report bugs, suggest ideas.
