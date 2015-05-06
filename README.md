# eton

eton is a note-taking cli tool.

## Install / Upgrade

    go get -u github.com/siadat/eton

## Examples

    # print the complete help
    eton -h

    # quick add
    eton new '[ ] do something'

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

    # alias matching is fuzzy
    eton cat prcs

    # pass items to xargs as filenames:
    eton ls something -l |xargs -i less {}

Notes are stored in ~/.etondb

    echo 'SELECT * from attributes LIMIT 10;' |sqlite3 ~/.etondb

Set `$EDITOR` environment variable to your prefered editor used by the `edit` command. E.g.,

I would love to hear how you use eton. Make pull requests, report bugs, suggest ideas.
