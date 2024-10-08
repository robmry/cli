# service logs

<!---MARKER_GEN_START-->
Fetch the logs of a service or task

### Options

| Name                 | Type     | Default | Description                                                                                     |
|:---------------------|:---------|:--------|:------------------------------------------------------------------------------------------------|
| `--details`          | `bool`   |         | Show extra details provided to logs                                                             |
| `-f`, `--follow`     | `bool`   |         | Follow log output                                                                               |
| `--no-resolve`       | `bool`   |         | Do not map IDs to Names in output                                                               |
| `--no-task-ids`      | `bool`   |         | Do not include task IDs in output                                                               |
| `--no-trunc`         | `bool`   |         | Do not truncate output                                                                          |
| `--raw`              | `bool`   |         | Do not neatly format logs                                                                       |
| `--since`            | `string` |         | Show logs since timestamp (e.g. `2013-01-02T13:23:37Z`) or relative (e.g. `42m` for 42 minutes) |
| `-n`, `--tail`       | `string` | `all`   | Number of lines to show from the end of the logs                                                |
| `-t`, `--timestamps` | `bool`   |         | Show timestamps                                                                                 |


<!---MARKER_GEN_END-->

## Description

The `docker service logs` command batch-retrieves logs present at the time of execution.

> [!NOTE]
> This is a cluster management command, and must be executed on a swarm
> manager node. To learn about managers and workers, refer to the
> [Swarm mode section](https://docs.docker.com/engine/swarm/) in the
> documentation.

The `docker service logs` command can be used with either the name or ID of a
service, or with the ID of a task. If a service is passed, it will display logs
for all of the containers in that service. If a task is passed, it will only
display logs from that particular task.

> [!NOTE]
> This command is only functional for services that are started with
> the `json-file` or `journald` logging driver.

For more information about selecting and configuring logging drivers, refer to
[Configure logging drivers](https://docs.docker.com/engine/logging/configure/).

The `docker service logs --follow` command will continue streaming the new output from
the service's `STDOUT` and `STDERR`.

Passing a negative number or a non-integer to `--tail` is invalid and the
value is set to `all` in that case.

The `docker service logs --timestamps` command will add an [RFC3339Nano timestamp](https://pkg.go.dev/time#RFC3339Nano)
, for example `2014-09-16T06:17:46.000000000Z`, to each
log entry. To ensure that the timestamps are aligned the
nano-second part of the timestamp will be padded with zero when necessary.

The `docker service logs --details` command will add on extra attributes, such as
environment variables and labels, provided to `--log-opt` when creating the
service.

The `--since` option shows only the service logs generated after
a given date. You can specify the date as an RFC 3339 date, a UNIX
timestamp, or a Go duration string (e.g. `1m30s`, `3h`). Besides RFC3339 date
format you may also use RFC3339Nano, `2006-01-02T15:04:05`,
`2006-01-02T15:04:05.999999999`, `2006-01-02T07:00`, and `2006-01-02`. The local
timezone on the client will be used if you do not provide either a `Z` or a
`+-00:00` timezone offset at the end of the timestamp. When providing Unix
timestamps enter seconds[.nanoseconds], where seconds is the number of seconds
that have elapsed since January 1, 1970 (midnight UTC/GMT), not counting leap
seconds (aka Unix epoch or Unix time), and the optional .nanoseconds field is a
fraction of a second no more than nine digits long. You can combine the
`--since` option with either or both of the `--follow` or `--tail` options.

## Related commands

* [service create](service_create.md)
* [service inspect](service_inspect.md)
* [service ls](service_ls.md)
* [service ps](service_ps.md)
* [service rm](service_rm.md)
* [service rollback](service_rollback.md)
* [service scale](service_scale.md)
* [service update](service_update.md)
