# service rm

<!---MARKER_GEN_START-->
Remove one or more services

### Aliases

`docker service rm`, `docker service remove`


<!---MARKER_GEN_END-->

## Description

Removes the specified services from the swarm.

> [!NOTE]
> This is a cluster management command, and must be executed on a swarm
> manager node. To learn about managers and workers, refer to the
> [Swarm mode section](https://docs.docker.com/engine/swarm/) in the
> documentation.

## Examples

Remove the `redis` service:

```console
$ docker service rm redis

redis

$ docker service ls

ID  NAME  MODE  REPLICAS  IMAGE
```

> [!WARNING]
> Unlike `docker rm`, this command does not ask for confirmation before removing
> a running service.

## Related commands

* [service create](service_create.md)
* [service inspect](service_inspect.md)
* [service logs](service_logs.md)
* [service ls](service_ls.md)
* [service ps](service_ps.md)
* [service rollback](service_rollback.md)
* [service scale](service_scale.md)
* [service update](service_update.md)
