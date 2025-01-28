#!/bin/bash

# Check if argument is provided and if it equals "true"
INVERT_BEHAVIOR=${1:-false}

# List of required secrets
REQUIRED_SECRETS=(
  "MY_SECRET"
  "MY_OTHER_SECRET"
  "HELLO_WORLD"
)

missing_secrets=()
defined_secrets=()

echo "Checking for required secrets..."

for secret in "${REQUIRED_SECRETS[@]}"; do
  if [ -z "${!secret}" ]; then
    missing_secrets+=("$secret")
    echo "❌ Missing secret: $secret"
  else
    defined_secrets+=("$secret")
    if [ "$INVERT_BEHAVIOR" = "true" ]; then
      echo "❌ Found secret when it should be undefined: $secret"
    else
      echo "✅ Found secret: $secret"
    fi
  fi
done

if [ "$INVERT_BEHAVIOR" = "true" ]; then
  if [ ${#defined_secrets[@]} -ne 0 ]; then
    echo "Error: Found ${#defined_secrets[@]} secrets when none should be defined"
    exit 1
  fi
  echo "Success: No secrets are defined (as required)"
else
  if [ ${#missing_secrets[@]} -ne 0 ]; then
    echo "Error: Missing ${#missing_secrets[@]} required secrets"
    exit 1
  fi
  echo "Success: All required secrets are available"
fi
