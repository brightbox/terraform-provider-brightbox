#!/bin/bash

set -e

[ "$#" -eq 1 ] || { echo "Specify tag as argument"; exit 1; }

current_branch=$(git rev-parse --abbrev-ref HEAD)

[ "${current_branch}" = "master" ] || { echo "Merge down to master branch first"; exit 1; }

tag=${1#v}
changelog=CHANGELOG.md
changedate=$(date "+%B %-d, %Y")

if grep "${tag}" "${changelog}"
then
	echo "Tag already released"
	exit 1
fi

#Prepend git log since last tag to Changelog
cat <<-EOF >"${changelog}"
## ${tag} (${changedate})

$(git log `git describe --tags --abbrev=0`..HEAD --oneline |
sed 's/^......./-/' |
egrep -v '^- (Typo|Merge)')

$(cat "${changelog}")
EOF

git commit -m "Update Changelog for release ${tag}" "${changelog}"
git tag -s -m "Release ${tag}" v${tag}
git push --follow-tags
