export tmp_dir=$(mktemp -d)
export project="e2e"
export git_dir="$tmp_dir/$project"

function check_equal() {
    if [ "$1" != "$2" ]; then
        ink -r "Expected ciux output ($1) to equal $2"
	exit 1
    else
        ink -g "Correct ciux output: $2"
    fi
}

