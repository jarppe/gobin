help:
  @just --list


ssh:
  @ssh jarppe-dev


pack:
  rm -fr ./dist
  yarn install
  yarn build
  yarn install --prod
  tar czf robin.tgz node_modules -C ./dist .
  rm -fr ./dist
  yarn install


install: pack
  rm -fr /opt/robin/*
  tar -C /opt/robin -xvzf robin.tgz
