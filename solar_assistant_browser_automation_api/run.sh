#!/usr/bin/with-contenv bashio

CONFIG_PATH=/data/options.json

echo "SOLAR_ASSISTANT_URL=$(bashio::config 'url')" > .env
echo "SOLAR_ASSISTANT_USER=$(bashio::config 'user')" >> .env
echo "SOLAR_ASSISTANT_PASS=$(bashio::config 'pass')" >> .env

./solar-assistant-browser-automation
