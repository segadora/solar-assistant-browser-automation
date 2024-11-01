# Solar assistant browser automation

[![Actions Status](https://github.com/segadora/solar-assistant-browser-automation/workflows/CI/badge.svg)](https://github.com/segadora/solar-assistant-browser-automation/actions?query=workflow%3ACI)

## Description

This go projects emulates a chrome browser and updates the work mode schedule.

This is relevant on Growatt inverters since solar assistant do not have any ways of updating these values from MQTT.

## Http endpoints

Updates the from time to 02:00, end time to 03:00 and sets enabled to true.

```
http://localhost:8080/work-mode-schedule?schedule1[from]=02&schedule1[to]=03&schedule1[priority]=Load first&schedule1[enabled]=1
```

## Docker compose

```
services:
  solar-assistant-browser-automation:
    image: ghcr.io/segadora/solar-assistant-browser-automation:latest
    container_name: "solar-assistant-browser-automation"
    restart: on-failure
    environment:
      SOLAR_ASSISTANT_URL: "https://xxx.eu.solar-assistant.io/"
      SOLAR_ASSISTANT_USER: "xxx"
      SOLAR_ASSISTANT_PASS: "xxx"
    ports:
      - 8080:8080
```
