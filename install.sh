#!/bin/bash
# Install terraform provider from gorelease archive
# (c) 2017 Brightbox Systems

set -e

for word in terraform-provider-*.tar.gz
do
	target="${word%.tar.gz}"
	elements=(${target//_/ })
	source_filename=${elements[0]?}
	target_dir=~/.terraform.d/plugins/${elements[2]?}_${elements[3]?}
	target_filename=${elements[0]?}_v${elements[1]?}
	if [ -f "${source_filename}" ]
	then
		echo "Detected plugin ${source_filename} at version ${elements[1]?}"
		echo "Installing..."
		mkdir -v -p ${target_dir}
		mv -v -u ${source_filename} ${target_dir}/${target_filename}
		echo "Done."
	else
		echo "Plugin ${source_filename} is missing"
		echo "Extract the tar file and run again"
	fi
done

