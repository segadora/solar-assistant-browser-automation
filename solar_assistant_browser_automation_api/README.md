Solar Assistant Browser Automation
===

Add the following to your **configuration.yaml**

```yaml
rest_command:
  solar_assistant_work_mode_schedule_update:
    url: 'http://localhost:8080/work-mode-schedule?schedule1[from]={{ from }}&schedule1[to]={{ to }}&schedule1[priority]={{ priority }}&schedule1[enabled]={{ enable }}'
    method: GET
    content_type:  'application/json; charset=utf-8'
    timeout: 120
```

The rest api can be called with the following method:

### Enable

```yaml
action: rest_command.solar_assistant_work_mode_schedule_update
data:
  from: "0500"
  to: "0600"
  priority: Battery first
  enable: "1"
```

| from                              | to                                | priority                                    | enable                                    |
|-----------------------------------|-----------------------------------|---------------------------------------------|-------------------------------------------|
| Timestamp with 4 digits 0000-2359 | Timestamp with 4 digits 0000-2359 | Battery first<br/>Load first<br/>Grid first | To enable use `1`<br/>To disable use `-1` |
