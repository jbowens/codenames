#!/bin/bash
set -ex
rm -rf ./dist
parcel build app.tsx game.css lobby.css
