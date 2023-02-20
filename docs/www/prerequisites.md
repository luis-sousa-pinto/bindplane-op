---
title: Prerequisites
excerpt: Installation Prerequisites
category: 636c08d51eb043000f8ce20e
slug: prerequisites
hidden: false
---

# Prerequisites

## Resource Requirements

BindPlane OP's resource requirements will differ based on the number of
managed agents. CPU, Memory, Disk throughput / iops, and network consumption
will scale linearly with the number of managed agents.

### Instance Sizing

Follow this table for CPU, memory, and storage capacity sizing.

| Agent Count  | CPU   | Memory | Storage Capacity |
| ------------ | ----- | ------ | ----------- |
| 10           | 1\*   | 1GB    | 10GB    |
| 100          | 1     | 2GB    | 10GB    |
| 2,000        | 2     | 8GB    | 20GB    |

\* A shared "burstable" core is suitable for 1-10 agents.

### Disk Performance Requirements

When using the default storage backend (`bbolt`), disk throughput and operations per second
will increase linearly with the number of managed agents. Enterprise deployments which are not using `bbolt`
can safely ignore this section.

To prevent disk performance bottlenecking, ensure that the underlying storage solution
can provide enough disk throughput and operations per second. Generally, cloud providers
will limit disk performance based on provisioned disk capacity.

| Agent Count | Read / Write Throughput | Read / Write IOPS |
| ----------- | ----------------------- | ----------------- |
| 10          | 1MB/s                   | 72/s   |
| 100         | 2 MB/s                  | 400/s  |
| 1,000       | 113 MB/s                | 5000/s |

### Network Requirements

BindPlane OP maintains network connections for the following:
- Agent Management
- Agent Throughput Measurements
- CLI and Web Interfaces

While BindPlane's observed network throughput is very low (less than 1mbps at 2,000 agents),
it is recommended to use a low latency network. Generally this means a modern network
interface supporting 1gpbs or greater speeds.
