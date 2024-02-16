# ipwatcher - Address Watchdog

`ipwatcher` is a robust IPv4/IPv6 address watchdog implemented in `Go`. It comes equipped with event actions, notifications, and exposes a REST API, enabling seamless integration with various systems and workflows.

### Installation & Usage

Installing `ipwatcher` is pretty straightforward, clone the repository and then run `make` to generate the executable:
```bash
> git clone git@github.com:gweebg/ipwatcher.git
> cd ipwatcher
> make
```
This should generate the executable at `ipwatcher/build`, with the name `ipwatcher`.

To use the application,  just execute `ipwatcher` with:
```bash
> ./ipwatcher --version (v4|v6)
```
This will start the service and check for changes for the specified IP version according to the parameters set in the configuration file (see [Configuration](configuring-the-service)). You also have access to some general flags that toggle certain behaviours of the application:

| Flag                  | Default      | Description                                                                               |
|-----------------------|--------------|-------------------------------------------------------------------------------------------|
| `--api`               | `false`      | Enable API mode, exposing a REST service on the port specified in the configuration file. |
| `--config <path>`     | `config.yml` | Set the `path` for the configuration file                                                 |
| `--exec`              | `true`       | Enable the execution of the actions defined in the configuration file.                    |
| `--notify`            | `false`      | Enable notifications via email.                                                           |
| `--version <v4\|v6>`  | `v4`         | Set the IP `version` for the watcher.                                                     |
| `--quiet`             | `false`      | Set the log level to `info` instead of `debug`.                                           |



### Configuring the Service

The configuration of the application is made via a YAML file, and allows configuring the different aspects that make the application stand out. By default, the service assumes the configuration path of 
`config.yml` in the root of the project, however, this behaviour can be changed by using the flag `--config <config_path>` when running the application. 

Let's explore the configuration file, section by section.

#### Sources Definition

`ipwatcher` relies on already existing and public API's to retrieve your public address, for redundancy purposes you can define one or more fall-back alternatives, used in case of failure of the previous one. 

```yaml
sources:
    - field: ip
      name: "ipify"
      type: json
      url:
        v4: https://api.ipify.org?format=json
        v6: https://api6.ipify.org?format=json

    - field: ip
      name: "myip"
      type: json
      url:
        v4: https://api4.my-ip.io/v2/ip.json
        v6: https://api6.my-ip.io/v2/ip.json
```

On the above example, we can see that each source is defined by a set of attributes:
- `name` corresponds to the name of the source, names are not unique and don't need to match the actual source name
- `type` defines the expected response type from the API (`json`, `text`, etc.)
- `field` is only used **when `type` is `json`** and dictates the field where the address is included on the `json` response
- `url` represents both v4 and v6 versions of the API `url` (at least one must be included)

Note that at least one source is needed for the application to run.

### Watcher Specific 

```yaml
watcher:

  timeout: 30 # in seconds
  force_source: "ipify" # should not be included if not used
  max_execution_time: 120 # in seconds
```

On the `watcher` section of the configuration file, you can specify execution related settings, such as:
- `timeout`, the time to wait between address checks (and consequently API calls)
- `force_source`, forces only a source (by its name) to be used
- `max_execution_time`, specifies the maximum time an action can be run for

### Event Handling

With `ipwatcher` you can act upon some events, like when the address is updated `on_change`, when the address stays the same `on_match` or when an error occurs `on_error`. For each event
you can define if you want to be notified and/or execute an action, for example, by running a Python script. My personal use-case is to update DNS records with the new address.

```yaml
watcher:
  ...
  events:
    on_change:
      notify: true # true | false, enables email notifications for this action

      actions:
        - type: "python" # to execute
          path: "scripts/script.py" # what to execute 
          args: "-x" # execution arguments

    on_match:
      ...

    on_error:
      ...
  ...
```

Event handlers don't need to be defined, as they are completely optional. Note that event actions have, by default, 60 seconds to execute, this behaviour can be changed by updating `watcher.max_execution_time`.

### Notification Settings

Notifications are, for now, only sent via email, thus when enabling notification for the events you need to define the `smtp` settings.

```yaml
  smtp:
    smtp_server: "smtp.gmail.com"
    smtp_port: 587

    username: "your_username"
    password: "your_app_password"

    from_address: "example@gmail.com"

    recipients:
      - address: "person_one@gmail.com"
        name: "Person One"

      - address: "person_two@proton.me"
        name: "Person Two"
```

When defining the `smtp` settings:
- `smtp_server` represents the `stmp` server (for example, `smtp.gmail.com` for Gmail)
- `stmp_port` is the port on which the `smtp_server` answers
- `username` and `password` are the credentials for the `smtp_server`
- `from_address` in this case should match the `username` and represents the sender email address
- `recipients` define the recipients of the notifications, represented by their `address` and `name` (both mandatory)

## API Settings

In Progress...
