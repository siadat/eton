# eton

eton is a note-taking cli tool.

## Install

    go get github.com/siadat/eton

## Examples

    # print the complete help
    eton -h

    # quick add
    eton new '[ ] do something'

    # add text from STDIN
    ps aux | eton new -

    # add a file
    eton addfile file.txt

    # unique aliases can be set and used instead of numeric ids
    etone alias 2 processes

    # list all items
    eton ls -Lall

    # filter items containing words "thing" AND "some"
    eton ls thing some

    # you can mark specific items
    etone mark processes 1

    # only list marked or aliased items (short mode)
    etone ls -s

    # open an item in less
    etone show processes

    # edit items
    etone edit {1..3}

    # alias matching is fuzzy
    etone cat prcs

    # pass items to xargs as filenames:
    etone ls something -l |xargs -i less {}

Notes are stored in ~/.etondb

    echo 'SELECT * from attributes LIMIT 10;' |sqlite3 ~/.etondb

I would love to hear how you use etone. Make pull requests, report bugs, suggest ideas.
