#!/bin/sh

# make sure gofmt is installed
command -v gofmt >/dev/null 2>&1 || { echo >&2 "gofmt is required but it's not installed."; exit 1; }

# (over)write the pre-commit hook
cat > .git/hooks/pre-commit <<'EOF'
#!/usr/bin/env bash
declare -a source_files=()
files=$(git diff-index --name-only HEAD)
for file in $files; do
    if [ ! -f "${file}" ]; then
        continue
    fi
    if [[ "${file}" == *.rs ]]; then
        source_files+=("${file}")
    fi
done

if [ ${#source_files[@]} -ne 0 ]; then
    command -v gofmt >/dev/null 2>&1 || { echo >&2 "gofmt is not installed. Aborting."; exit 1; }
    $(command -v gofmt) -w ${source_files[@]} &
fi
wait
if [ ${#source_files[@]} -ne 0 ]; then
    git add ${source_files[@]}
    echo "adjusted source formatting in ${source_files[@]}"
else
    echo "no formatting was needed"
fi
EOF

chmod +x .git/hooks/pre-commit

echo "hook updated"
