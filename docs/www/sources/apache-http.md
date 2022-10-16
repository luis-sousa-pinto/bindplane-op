---
title: "Apache HTTP"
slug: "apache-http"
hidden: false
createdAt: "2022-08-02T13:44:01.491Z"
updatedAt: "2022-08-10T15:31:03.366Z"
---
## Supported Platforms

| Platform | Metrics | Logs | Traces |
| :------- | :------ | :--- | :----- |
| Linux    | ✓       | ✓    |        |
| Windows  | ✓       | ✓    |        |
| macOS    | ✓       | ✓    |        |

## Configuration Table

| Parameter           | Type       | Default                         | Description                                                                                                                                         |
| :------------------ | :--------- | :------------------------------ | :-------------------------------------------------------------------------------------------------------------------------------------------------- |
| enable_metrics      | `bool`     | true                            | Enable to send metrics.                                                                                                                             |
| hostname\*          | `string`   | localhost                       | The hostname or IP address of the Apache HTTP system.                                                                                               |
| port                | `int`      | 3000                            | The TCP port of the Apache HTTP system.                                                                                                             |
| collection_interval | `int`      | 60                              | How often (seconds) to scrape for metrics.                                                                                                          |
| enable_tls          | `bool`     | false                           | Whether or not to use TLS when connecting to the Apache HTTP server.                                                                                |
| strict_tls_verify   | `bool`     | false                           | Enable to require TLS certificate verification.                                                                                                     |
| ca_file             | `string`   |                                 | Certificate authority used to validate TLS certificates. Not required if the collector's operating system already trusts the certificate authority. |
| mutual_tls          | `bool`     | false                           | Enable to require TLS mutual authentication.                                                                                                        |
| cert_file           | `string`   |                                 | A TLS certificate used for client authentication, if mutual TLS is enabled.                                                                         |
| key_file            | `string`   |                                 | A TLS private key used for client authentication, if mutual TLS is enabled.                                                                         |
| enable_logs         | `bool`     | true                            | Enable to collect Apache HTTP logs.                                                                                                                 |
| start_at            | `enum`     | end                             | Start reading logs from 'beginning' or 'end'.                                                                                                       |
| access_log_path     | `strings`  | ["/var/log/apache2/access.log"] | Access Log File paths to tail for logs.                                                                                                             |
| error_log_path      | `strings`  | ["/var/log/apache2/error.log"]  | Error Log File paths to tail for logs.                                                                                                              |
| timezone            | `timezone` | "UTC"                           | The timezone to use when parsing timestamps.                                                                                                        |

<span style="color:red">\*_required field_</span>