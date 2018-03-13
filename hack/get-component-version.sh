#!/usr/bin/env sh
set -e

get_hash()
{
    local DEP_DIRS="$1"
    # We use this to determine if there are any new additions
    local GIT_STATUS="$(git status $DEP_DIRS --porcelain)"
    # To determine if there are any changes
    local GIT_DIFF_INDEX="$(git diff-index -p HEAD -- $DEP_DIRS)"
    # Whether anything changed in the repo
    export GIT_DIRTY="$GIT_STATUS$GIT_DIFF_INDEX"
    if [ -n "$GIT_DIRTY" ]; then
        DIRTY_HASH_SHORT="$(echo $GIT_DIRTY | sha256sum | awk '{print $1}' | tail -c 9)"
        CHANGES_HASH="-dirty-$DIRTY_HASH_SHORT"
    else
        CHANGES_HASH=""
    fi
    # Get the current commit id
    local COMMIT_HASH_SHORT="$(git log -n 1 --pretty=format:%h -- $DEP_DIRS)"
    echo "$COMMIT_HASH_SHORT$CHANGES_HASH"
}

get_dep_dirs()
{
    local COMP=$1
    if [ ! -e "$COMP" ]; then
        echo "nothing at : $COMP"
        exit 1
    fi
    local COMP_DEPS="$COMP/BUILD_DEPS"
    if [ -e $COMP_DEPS ]; then
        EXTRA_DEPS=$(cat $COMP_DEPS)
        ALL_DEPS="$COMP $EXTRA_DEPS"
    else
        ALL_DEPS="$COMP"
    fi
    for DEP in $ALL_DEPS; do
        if [ ! -e "$DEP" ]; then
            echo "nothing at $DEP"
            exit 1
        fi
    done
    echo $ALL_DEPS
}

get_comp_hash()
{
    local COMP="$1"
    local DEP_DIRS="$(get_dep_dirs $COMP)"
    # echo $DEP_DIRS
    get_hash "$DEP_DIRS"
}

get_comp_hash $1
