#!/usr/bin/env bash

for EXAMPLE_DIR in `ls -d ./examples/*/`; do
	terraform validate $EXAMPLE_DIR
	EXIT_CODE=$?

	if [[ $EXIT_CODE -ne 0 ]]; then
		exit $EXIT_CODE
	fi
done

exit 0
