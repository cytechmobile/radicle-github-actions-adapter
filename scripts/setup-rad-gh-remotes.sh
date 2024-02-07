#!/bin/bash
# Sets up a repo fot the Radicle Github Actions Adapter
# usage:
# ./setup-rad-gh-remotes.sh

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN="\033[0;36m"
ERROR='\033[41m'
NC='\033[0m'

trap 'last_command=$current_command; current_command=$BASH_COMMAND' DEBUG
trap 'catch' EXIT
catch() {
if [ $? -ne 0 ]; then
        printf "${ERROR}${last_command} command failed with exit code $?.${NC}\n"
fi
}

printf "\n${GREEN}This script will setup the repository for the Radicle Github Actions Adapter.${NC}\n"

printf "\n${CYAN}    Please provide the Github's workspace name: ${NC}"
read workspace_name
printf "\n${CYAN}    Please provide the Github's repo name: ${NC}"
read repo_name

printf "\n${GREEN}Populating .radicle/github-action.yaml sile.${NC}\n"
mkdir -p .radicle
echo "github_username: $workspace_name
github_repo: $repo_name" > .radicle/github-action.yaml

printf "\n${GREEN}Updating git remote.${NC}\n"
REPO_ID=$(rad .)
NODE_ID=$(rad self --nid)
git remote add both rad://$REPO_ID
git remote set-url --push both https://github.com/$workspace_name/$repo_name.git
git remote set-url --add --push both rad://$REPO_ID/$NODE_ID

printf "\n${GREEN}Setup completed successfully${NC}\n"
