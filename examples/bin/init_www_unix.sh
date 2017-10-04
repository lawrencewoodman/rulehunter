#!/usr/bin/env bash
#
#  This command prepares the www directory for the examples
#

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SUPPORT_DIR="$SCRIPT_DIR/../../support"
WWW_DIR="$SCRIPT_DIR/../www"
mkdir -p $WWW_DIR

cp -r $SUPPORT_DIR/bootstrap/* $WWW_DIR
cp -r $SUPPORT_DIR/jquery/* $WWW_DIR
cp -r $SUPPORT_DIR/rulehunter/* $WWW_DIR
cp -r $SUPPORT_DIR/html5shiv/js/* $WWW_DIR/js/
cp -r $SUPPORT_DIR/respond/js/* $WWW_DIR/js/
