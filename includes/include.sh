#!/usr/bin/env bash

set -e

NC=$(tput sgr0)
RED=$(tput setaf 1)
GRN=$(tput setaf 2)
CYN=$(tput setaf 6)
PRP=$(tput setaf 5)
ERR=$(tput setaf 9)
BOLD=$(tput bold)

function echo_green() {
  echo "${GRN}$*${NC}"
}
function echo_red() {
  echo "${RED}$*${NC}"
}
function echo_cyan() {
  echo "${CYN}$*${NC}"
}
function echo_purple() {
  echo "${PRP}$*${NC}"
}
function echo_error() {
  echo "${ERR}${BOLD}❗❗ ${*^^}${NC}"
}

function slugify() {
  echo "$1" | iconv -t ascii//TRANSLIT | sed -r s/[~^]+//g | sed -r s/[^a-zA-Z0-9]+/-/g | sed -r s/^-+\|-+$//g | tr "[:upper:]" "[:lower:]"
}
