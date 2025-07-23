if [ -z "$selector" ]; then
  echo "Error: Selector (\$selector)is required." >&2
  exit 1
else
  echo "Using selector: $selector"
  selector_opt="--selector $selector"
fi

git_dir=$(git rev-parse --show-toplevel 2>/dev/null)
if [ -z "$git_dir" ]; then
  echo "Not a git repository. Please run this command inside a git repository." >&2
  exit 1
fi

if ciuxconfig=$(ciux get configpath $selector_opt "$git_dir" 2>&1); then
  source "$ciuxconfig"
else
  echo "Error while loading ciux config : $ciuxconfig" >&2
  exit 1
fi
