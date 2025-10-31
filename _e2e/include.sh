export tmp_dir=$(mktemp -d)
export git_tag_v1="v1.0.0"

# project_dir is the absolute name of the git repository
project_dir=$DIR/..
CIUXCONFIG=$(ciux get configpath -l ci "$project_dir")
. $CIUXCONFIG

function check_equal() {
    if [ "$1" != "$2" ]; then
        ink -r "Expected ciux output ($1) to equal $2"
	    exit 1
    else
        ink -g "Correct ciux output: '$2'"
    fi
}

