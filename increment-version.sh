#! /bin/bash

function version() {
	version=$(cat VERSION.txt)
	a=( ${version//./ } )
	major=${a[0]}
	minor=${a[1]}
	patch=${a[2]}

	while true; do
		read -p "Increment version [(M)ajor/m(I)nor/(P)atch]?" part
		if [[ "$part" == "" ]]; then
			part="M"
		fi
		case "$part" in
			[Mm])
				major=$((major + 1))
				minor=0
				patch=0
				break
			;;
			[Ii])
				minor=$((minor + 1))
				patch=0
				break
			;;
			[Pp])
				patch=$((patch + 1))
				break
			;;
			*)
				echo "Enter M to increment major version, I for minor version, or P for patch"
			;;
		esac
	done

	version="$major.$minor.$patch"
	echo "$version"
	echo "$version" > VERSION.txt
	git tag "v$version"
}

version $@
