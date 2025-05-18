#!/bin/bash

# Check if the 'migrate' command is installed
if ! command -v migrate &> /dev/null; then
  # command -v migrate: command checks if the command migrate is in the PATH and is executable
  # -v option for command makes it print the full path of the command migrate if found
  echo "Error: 'migrate' command not found. Please install golang-migrate: follow instructions here:"
  echo "https://github.com/golang-migrate/migrate/tree/master/cmd/migrate"
  exit 1 # Exit with error code 1
fi

if [ -z "${XRF_Q2_BID_PG_DB_URL}" ]; then # -z to test if the length of a string is zero.
  echo "Error: Environment variable XRF_Q2_BID_PG_DB_URL is not set."
  exit 1 # Exit with error code 1
fi

# Run go-migrate migrations
echo "Running database migrations..."
echo "DB URL ==> ${XRF_Q2_BID_PG_DB_URL}"

if migrate -database "${XRF_Q2_BID_PG_DB_URL}" -path ../storage/migrations -verbose up; then
  echo "Database migrations completed successfully."
else
  echo "Error: Database migrations failed. Please check the logs and your go-migrate setup."
  exit 1 # Exit with error code if migrations failed
fi

exit 0 # Exit with success code
