#!/bin/bash
if [ -z "${XRF_Q2_BID_PG_DB_URL}" ]; then # -z to test if the length of a string is zero.
  echo "Error: Environment variable XRF_Q2_BID_PG_DB_URL is not set."
  echo "Please set this environment variable before running migrations."
  exit 1 # Exit with error code 1
fi

# Run go-migrate migrations
echo "Running database migrations..."

if migrate -database "${XRF_Q2_BID_PG_DB_URL}" -path ../db/migrations up; then
  echo "Database migrations completed successfully."
else
  echo "Error: Database migrations failed. Please check the logs and your go-migrate setup."
  exit 1 # Exit with error code if migrations failed
fi

exit 0 # Exit with success code
