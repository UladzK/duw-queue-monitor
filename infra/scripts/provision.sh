#!/bin/bash

set -euo pipefail

# --- Validate Inputs ---

# Usage check
if [[ $# -ne 2 ]]; then
  echo "Usage: $0 <module> <env>"
  exit 1
fi

# --- Input Arguments ---
MODULE="$1"
ENV="$2"
SCRIPT_DIR="$(dirname "$(realpath "$0")")"
MODULE_DIR="$SCRIPT_DIR/../terraform/$MODULE"

# --- Constants ---
readonly AZURE_SUB="77a70a5e-2230-43b7-8983-61e7497498a8"

# --- Script workflow ---

if [[ ! -d "$MODULE_DIR" ]]; then
  echo "❌ Error: Module directory '$MODULE_DIR' does not exist."
  exit 1
fi

echo "🔧 Module dir: $MODULE_DIR"
echo "🌍 Environment: $ENV"
cd "$MODULE_DIR"

# --- Check Infisical Login ---
if ! infisical user switch; then
  echo "Log in to infisical first using (infisical login) command"
  exit 1
fi

# --- Check Azure CLI Login ---
if ! az account show; then
  echo "☁️ Logging in to Azure..."
  az login
fi

# --- Setting the subscription ---
az account set --subscription $AZURE_SUB

# --- Terraform Init ---
echo "📦 Initializing Terraform backend..."
terraform init -backend-config="envs/$ENV/backend.hcl"

# --- Terraform Plan ---
echo "🛠️ Planning infrastructure changes..."
PLAN_OUT="${ENV}.tfplan"

infisical run --env="$ENV" -- terraform plan \
  -var-file="envs/$ENV/$ENV.tfvars" \
  -out="$PLAN_OUT"

echo -e "\n✅ Terraform plan created: $PLAN_OUT"
echo -n "❓ Do you want to apply this plan? (y/n): "
read -r CONFIRM

if [[ "$CONFIRM" == "y" ]]; then
  echo "🚀 Applying Terraform plan..."
  terraform apply "$PLAN_OUT"
else
  echo "❌ Apply aborted."
fi
