#!/bin/sh

# Run some pre commit checks on the Go source code. Prevent the commit if any errors are found
echo "Running pre-commit checks on your code..."

FILES=$(go list ./...  | grep -v /vendor/)

# Format the Go code
go fmt ${FILES}

# Check all files for errors
{
	errcheck -ignoretests ${FILES}
} || {
	exitStatus=$?

	if [ $exitStatus ]; then
		printf "\nErrors found in your code, please fix them and try again."
		exit 1
	fi
}

# Check all files for suspicious constructs
{
	go vet ${FILES}
} || {
	exitStatus=$?

	if [ $exitStatus ]; then
		printf "\nIssues found in your code, please fix them and try again."
		exit 1
	fi
}

{
	make test
} || {
	exitStatus=$?

	if [ $exitStatus ]; then
		printf "\nTest Errors found in your code, please fix them and try again."
		exit 1
	fi
}