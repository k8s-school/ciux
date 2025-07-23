      if ciuxconfig=$(ciux get configpath --selector "build" "$git_dir" 2>&1); then
        source "$ciuxconfig"
      else
        echo "Error while loading ciux config : $ciuxconfig" >&2
        exit 1
      fi
