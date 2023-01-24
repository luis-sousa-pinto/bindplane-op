---
title: Filter HTTP Status
category: 636c08e1212e49001e7a3032
parentDoc: 636c0a0cddea2a005d14423a
slug: filter-http-status
hidden: false
---

##  HTTP Status Filter Processor

The HTTP Status processor can be used to filter out logs that contain a status code between a minimum and a maximum status code.

## Supported Types

| Metrics | Logs | Traces |
| :--- | :--- | :--- |
|  | ✓ |  |

## Configuration Table

| Parameter  | Type    | Default  | Description |
| :---       | :---    | :---     | :--- |
| minimum    | `enum`  | `100` | Minimum Status to match. Log entries with lower status codes will be filtered. |
| maximum    | `enum`  | `599` | Maximum Status to match. Log entries with higher status codes will be filtered. |


Valid Minimum Status Codes:
- 100
- 200
- 300
- 400
- 500

Valid Maximum Status Codes:
- 199
- 299
- 399
- 499
- 599


## Example Configuration

Filter out all 1xx status codes and 2xx status codes.
 
**Web Interface**

![Filter_HTTP_Status](https://storage.googleapis.com/bindplane-op-doc-images/resources/processor-types/filter_http_status.png)

**Standalone Processor**

```yaml
apiVersion: bindplane.observiq.com/v1
kind: Processor
metadata:
  id: http-status
  name: http-status
spec:
  type: filter_http_status
  parameters:
    - name: maximum
      value: 599
    - name: minimum
      value: 300
```

**Configuration with Embedded Processor**

```yaml
apiVersion: bindplane.observiq.com/v1
kind: Configuration
metadata:
  id: http-status
  name: http-status
  labels:
    platform: linux
spec:
  sources:
    - type: journald
      parameters:
        - name: units
          value: []
        - name: directory
          value: ""
        - name: priority
          value: info
        - name: start_at
          value: end
      processors:
        - type: filter_http_status
          parameters:
            - name: maximum
              value: 599
            - name: minimum
              value: 300
  selector:
    matchLabels:
      configuration: http-status
```
