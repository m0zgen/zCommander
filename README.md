# zCommander

Send mass or single command to remote server via Web requests.

## Configuration

- `groups` - list of groups to send command to.
- `creation_timeout` - timeout for command creation, actually for `-mass` argument.
- `users.txt` - list of users (or anything) to send command to.
- Build in commands `add_command` (`-add` flag), `remove_command` (`-remove` flag)

## Usage

Single operation:

```shell
./zCommander -config=config-prod.yml -group=testbld -command='/add?user=user_t4'
```

Mass add:

```shell
./zCommander -config=config-prod.yml -group=testbld -mass -add
````

Mass remove:

```shell
./zCommander -config=config-prod.yml -group=testbld -mass -remove
```