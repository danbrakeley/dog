# Dog

A logger that spins up a web browser and sends logs in real-time via websocket to one or more clients.

## How to Build

Running `build-dev.sh` will build/install the `bpak` generator, run `go generate`, run tests, then finally build the example executable. From there just run `/local/example` and connect your web browser to `127.0.0.1:8080`.

## TODO

- ~~cleanup http server and any open connections on Close()~~
- add visually distinct formatting for fatal log lines
- Instead of mirroring frog's interface, maybe make it a module that requires frog to work?
- Add options to periodically send memory/gc stats to the web viewer
  - maybe a generic system for adding named fields and displaying them on the sidebar?
