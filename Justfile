help:
  @just --list


ssh:
  @ssh jarppe-dev


build:
  #!/bin/bash
  cd src
  go build

test: build
  ./src/gobin -h jarppe-dev -s /Users/jarppe/swd/jarppe/gobin/example
