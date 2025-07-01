#!/bin/bash

# ARK API Configuration Update Script
# Updates deep-coding-agent configuration to use ARK API

set -e

# ARK API Configuration
ARK_API_KEY="xxx"
ARK_MODEL="xxx"
ARK_BASE_URL="https://ark.cn-beijing.volces.com/api/v3"

echo "🔧 Updating deep-coding-agent configuration for ARK API..."

# Update main configuration
./deep-coding-agent config set api_key "$ARK_API_KEY"
./deep-coding-agent config set base_url "$ARK_BASE_URL"
./deep-coding-agent config set model "$ARK_MODEL"

# Update models.basic configuration
./deep-coding-agent config set models.basic.api_key "$ARK_API_KEY"
./deep-coding-agent config set models.basic.base_url "$ARK_BASE_URL"
./deep-coding-agent config set models.basic.model "$ARK_MODEL"

# Update models.reasoning configuration
./deep-coding-agent config set models.reasoning.api_key "$ARK_API_KEY"
./deep-coding-agent config set models.reasoning.base_url "$ARK_BASE_URL"
./deep-coding-agent config set models.reasoning.model "$ARK_MODEL"

echo "✅ Configuration updated successfully!"
echo ""
echo "📋 Current configuration:"
./deep-coding-agent config show