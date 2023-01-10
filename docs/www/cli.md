---
title: CLI
category: 636c08d51eb043000f8ce20e
slug: cli
hidden: false
---

You can access the BindPlane CLI by using the `bindplane` command from the install directory or preceded by the absolute path of the install directory.

## Installing BindPlane Client (Remote CLI)

BindPlane CLI can be run remotely (from a workstation), packages and binary releases are available on the [Downloads](doc:downloads) page.

For macOS and Windows, place the BindPlane binary / exe in the user's path or execute it directly. For Debian and RHEL platforms, the installation is the same as BindPlane server, however, only the BindPlane binary is installed (User and config, log, storage directories are not created).

## CLI Commands

| Command      | Description                                                |
| :----------- | :--------------------------------------------------------- |
| `apply`      | Apply resources                                            |
| `completion` | Generate the autocompletion script for the specified shell |
| `delete`     | Delete bindplane resources                                 |
| `get`        | Display one or more resources                              |
| `help`       | Help about any command                                     |
| `init`       | Initialize an installation                                 |
| `install`    | Install a new agent                                        |
| `label`      | List or modify the labels of a resource                    |
| `profile`    | Profile commands.                                          |
| `serve`      | Starts the server                                          |
| `sync`       | Sync an agent-version from github                          |
| `update`     | Update an existing agent                                   |
| `validate`   | validate the current profile                               |
| `version`    | Prints BindPlane version                                   |

| Flags                            | Description                                                                                |
| :------------------------------- | :----------------------------------------------------------------------------------------- |
| `-c`, `--config string`          | full path to configuration file                                                            |
| `--env string`                   | BindPlane environment. One of test|development|production (default "production")           |
| `-h`, `--help`                   | help for bindplane                                                                         |
| `--host string`                  | domain on which the BindPlane server will run (default "localhost")                        |
| `--log-file-path string`         | full path of the BindPlane log file, defaults to $HOME/.bindplane/bindplane.log            |
| `--log-output string`            | output of the log. One of: file|stdout                                                     |
| `--otlp-tracing-endpoint string` | endpoint to send OTLP traces to                                                            |
| `--otlp-tracing-insecure-tls`    | set true to allow insecure TLS                                                             |
| `-o`, `--output string`          | output format. One of: json\|table\|yaml\|raw (default "table")                            |
| `--password string`              | password to use with Basic auth (default "admin")                                          |
| `--port string`                  | port on which the rest server is listening (default "3001")                                |
| `--profile string`               | configuration profile name to use                                                          |
| `--server-url string`            | http url that clients use to connect to the server                                         |
| `--tls-ca strings`               | TLS certificate authority file(s) for mutual TLS authentication                            |
| `--tls-cert string`              | TLS certificate file                                                                       |
| `--tls-key string`               | TLS private key file                                                                       |
| `--tls-skip-verify`              | Whether to verify the server's certificate chain and host name when making client requests |
| `--trace-type string`            | type of trace to use for tracing requests, either 'otlp' or 'google'                       |
| `--username string`              | username to use with Basic auth (default "admin")                                          |
