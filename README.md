# Dog

A logger that spins up a web browser and sends logs in real-time via websocket.

## How to Build

1. If you don't already have it, [install a recent version of Go](https://golang.org/doc/install).
2. If you are on Windows, you'll want a bash shell.
   - I use git-bash, which comes with [Git for Windows](https://gitforwindows.org).
3. In bash, `cd` into the path where you cloned this repo, then run `./build-dev.sh`, which does the following:
   - rebuilds/installs bpak.exe, if needed
   - regenerates the bpak'd html/jss/cs files
   - runs tests
   - builds cmd/example, writing the output to `local/example[.exe]`.
4. Run `local/example[.exe]`, then point a web browser at `localhost:8080`

I've also been playing with [Air](https://github.com/cosmtrek/air) for hot-reloading. I create [my own fork/branch](https://github.com/danbrakeley/air/tree/brakeley) to pull in a fix and add a [quick install script](https://github.com/danbrakeley/air/blob/brakeley/brakeley-install.sh).

## TODO

- display structured data
  - toggle on/off?
  - add filter for specific fields that must exist
  - add way to filter on field values (exact, and ranges for numbers)
- "currently visible" -> "active lines", then add "visible" as readonly display of actually visible lines
- resizable column widths
- Add options to periodically send memory/gc stats to the web viewer
  - maybe a generic system for adding named fields and displaying them on the sidebar?
- Instead of mirroring frog's interface, maybe make it a module that requires frog to work?
- add multiple filters that can be quickly switched between
- ~~cleanup http server and any open connections on Close()~~
- ~~add visually distinct formatting for fatal log lines~~
- ~~save captured logs to text ndjson file~~
- ~~add toggles for log levels~~
- ~~add simple text filter (or regex?)~~
