#!/bin/bash

# Telegram Notification Script for Vbitez
# Usage: ./scripts/telegram-notify.sh <type> <message>
# Types: app-deployment, library-update

set -e

TYPE=$1
MESSAGE=$2

if [ -z "$TYPE" ] || [ -z "$MESSAGE" ]; then
    echo "Usage: $0 <type> <message>"
    echo "Types: app-deployment, library-update"
    exit 1
fi

# Telegram configuration
TELEGRAM_BOT_TOKEN="7250945940:AAGxOe0ahPZy2UfzGjQfXR4XC5vjThrnvow"
TELEGRAM_CHAT_ID="-4810256950"

# Send to Telegram
curl -s -X POST \
    "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage" \
    -H "Content-Type: application/json" \
    -d "{
        \"chat_id\": \"${TELEGRAM_CHAT_ID}\",
        \"text\": \"${MESSAGE}\"
    }"

echo "✅ Telegram notification sent for: $TYPE" 