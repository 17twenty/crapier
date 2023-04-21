#!/usr/bin/env bash

# attach 'path_to_workspace/sqlite.db3' as toMerge;

# BEGIN;
#  sqlite> insert into StandardSchema select * from toMerge.StandardSchema;
# COMMIT;
# detach database toMerge;