# Contributing

Pull requests and Bug reports are always welcome.
When contributing to this repository, please first discuss the change you wish to make by opening an issue.

Please note we have a code of conduct, please follow it in all your interactions with the project.

## Getting started
1. Clone the repo using `git clone https://github.com/sivaramasubramanian/csvprocessor.git`
2. Install `go` and `make` if not already installed.
3. Run `make install` to install the dev dependencies.
4. Make the changes as per [coding guidelines](#coding-guidelines) and run the unit tests using `make test`
5. To run benchmarks, use `make bench`
6. After verifying the test results, commit the code with commit as per the [commit message guidelines](#commit-messages)
7. Push the code and raise a PR for review and follow the [Pull request process](#pull-request-process)

## Coding guidelines
1. Ensure to add test cases for any changes, to keep the code coverage at a maximum.
2. Always run lint (golang-ci lint) before commiting any changes.
3. Follow the guidelines given in [Golang CodeReviewComments]("https://github.com/golang/go/wiki/CodeReviewComments")
4. Add comments judiciously explaining any changes, wherever necessary.

## Commit messages
Please follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for commit messages

## Pull Request Process
1. Ensure that the build pipeline succeeds and if there are any breakages, please change the PR to draft state till the issues are fixed.
2. Update the README.md with details of changes to the interface.
3. Increase the version numbers in any examples files and the README.md to the new version that this
   Pull Request would represent. The versioning scheme we use is [SemVer](http://semver.org/).
4. You may merge the Pull Request in once you have the sign-off of the maintainers, or if you 
   do not have permission to do that, you may request the maintainer to merge it for you.
