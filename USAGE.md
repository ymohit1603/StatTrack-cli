# Usage

Options can be passed to wakatime-cli via command line, or set in the `$WAKATIME_HOME/.wakatime.cfg` config file.
`$WAKATIME_HOME` defaults to your user's `$HOME` directory.
Command line arguments take precedence over config file settings.
Run `wakatime-cli --help` for available command line options.

## Project Detection

WakaTime auto-detects your projects.

The priority of projects detection is:

1. [.wakatime-project file](#wakatime-project-file)

2. [project map](#project-map-section)

3. Version control (Git)

4. IDE project

See the [source code](https://github.com/wakatime/wakatime-cli/blob/36f6372880d7113382e99453c2b94ff727788ae2/pkg/project/project.go#L145) for specifics.

### WakaTime Project File

To overwrite the auto-detected project, create a `.wakatime-project` file in your project’s root folder.
The first line of the file contents overwrites the project name, if present.
The second line, if present, overwrites the current branch name when working inside this folder.
When the `.wakatime-project` file is empty, the folder’s name is used as the project name.
Whenever a `.wakatime-project` file is found, it overwrites all other project detection.

## INI Config File

Here's an example `$WAKATIME_HOME/.wakatime.cfg` config file with all available options:

```ini
[settings]
debug = false
api_key = your-api-key
api_key_vault_cmd = command arg arg ... (space-separated, no shell syntax)
api_url = https://api.wakatime.com/api/v1
hide_file_names = false
hide_project_names = false
hide_branch_names = false
hide_dependencies = false
hide_project_folder = false
exclude =
    ^COMMIT_EDITMSG$
    ^TAG_EDITMSG$
    ^/var/(?!www/).*
    ^/etc/
include =
    .*
include_only_with_project_file = false
exclude_unknown_project = false
status_bar_enabled = true
status_bar_coding_activity = true
status_bar_hide_categories = false
offline = true
proxy = https://user:pass@localhost:8080
no_ssl_verify = false
ssl_certs_file =
timeout = 30
hostname = machinename
log_file =
import_cfg = /path/to/another/wakatime.cfg
metrics = true
guess_language = true

[projectmap]
projects/foo = new project name
^/home/user/projects/bar(\d+)/ = project{0}

[project_api_key]
projects/foo = your-api-key
^/home/user/projects/bar(\d+)/ = your-api-key

[git]
submodules_disabled = false
project_from_git_remote = false

[git_submodule_projectmap]
some/submodule/name = new project name
^/home/user/projects/bar(\d+)/ = project{0}
```

### Settings Section

| option                         | description | type | default value |
| ---                            | ---         | ---  | ---           |
| debug                          | Turns on debug messages in log file. | _bool_ | `false` |
| api_key                        | Your wakatime api key. | _string_ | |
| api_key_vault_cmd              | A command to get your api key, perhaps from some sort of secure vault. Actually a space-separated list of an executable and its arguments. Executables in PATH can be referred to by their basenames. Shell syntax not supported. | _string_ | |
| api_url                        | The WakaTime API base url. | _string_ | <https://api.wakatime.com/api/v1> |
| heartbeat_rate_limit_seconds   | Rate limit sending heartbeats to the API once per duration. Set to 0 to disable rate limiting. | _int_ | `120` |
| hide_file_names                | Obfuscate filenames. Will not send file names to api. | _bool_;_list_ | `false` |
| hide_project_names             | Obfuscate project names. When a project folder is detected instead of using the folder name as the project, a `.wakatime-project file` is created with a random project name. | _bool_;_list_ | `false` |
| hide_branch_names              | Obfuscate branch names. Will not send revision control branch names to api. | _bool_;_list_ | `false` |
| hide_dependencies              | Prevent sending imports/libraries/dependencies used in currently focused file to the api.  | _bool_;_list_ | `false` |
| hide_project_folder            | When set, send the file's path relative to the project folder. For ex: `/User/me/projects/bar/src/file.ts` is sent as `src/file.ts` so the server never sees the full path. When the project folder cannot be detected, only the file name is sent. For ex: `file.ts`. | _bool_ | `false` |
| exclude                        | Filename patterns to exclude from logging. POSIX regex syntax. | _bool_;_list_ | |
| include                        | Filename patterns to log. When used in combination with `exclude`, files matching `include` will still be logged. POSIX regex syntax | _bool_;_list_ | |
| include_only_with_project_file | Disables tracking folders unless they contain a `.wakatime-project file`. | _bool_ | `false` |
| exclude_unknown_project        | When set, any activity where the project cannot be detected will be ignored. | _bool_ | `false` |
| status_bar_enabled             | Turns on wakatime status bar for certain editors. | _bool_ | `true` |
| status_bar_coding_activity     | Enables displaying Today's code stats in the status bar of some editors. When false, only the WakaTime icon is displayed in the status bar. | _bool_ | `true` |
| status_bar_hide_categories     | When `true`, --today only displays the total code stats, never displaying Categories in the output. | _bool_ | `false` |
| offline                        | Enables saving code stats locally to ~/.wakatime/offline_heartbeats.bdb when offline, and syncing to the dashboard later when back online. | _bool_ | `true` |
| proxy                          | Optional proxy configuration. Supports HTTPS, SOCKS and NTLM proxies. For ex: `https://user:pass@host:port`, `socks5://user:pass@host:port`, `domain\\user:pass` | _string_ | |
| no_ssl_verify                  | Disables SSL certificate verification for HTTPS requests. By default, SSL certificates are verified. | _bool_ | `false` |
| ssl_certs_file                 | Path to a CA certs file. By default, uses bundled Letsencrypt CA cert along with system ca certs. | _filepath_ | |
| timeout                        | Connection timeout in seconds when communicating with the api. | _int_ | `120` |
| hostname                       | Optional name of local machine. By default, auto-detects the local machine’s hostname. | _string_ | |
| log_file                       | Optional log file path. | _filepath_ | `~/.wakatime/wakatime.log` |
| import_cfg                     | Optional path to another wakatime.cfg file to import. If set it will overwrite values loaded from $WAKATIME_HOME/.wakatime.cfg file. | _filepath_ | |
| metrics                        | When set, collects metrics usage in '~/.wakatime/metrics' folder. For further reference visit <https://go.dev/blog/pprof>. | _bool_ | `false` |
| guess_language                 | When `true`, enables detecting programming language from file contents. | _bool_ | `false` |

### Project Map Section

A key value pair list separated by new line, where the value before equal sign is the regex pattern and the latter is the override project name. Use when a project should be renamed to another when sent to the API. The order of the project map is important, the first match is used. Do not add any leading space before the regex pattern.

```ini
[projectmap]
projects/foo = new project name
^/home/user/projects/bar(\d+)/ = project{0}
```

### Project Api Key Section

A key value pair list separated by new line, where the value before equal sign is the regex pattern and the latter is the override api key. Use when a project should be sent using another api key other than the default on `settings.api_key`. The order of the project api key is important, the first match is used. Do not add any leading space before the regex pattern.

```ini
[project_api_key]
projects/foo = your-api-key
^/home/user/projects/bar(\d+)/ = your-api-key
```

### Api Key Environment Variable

If a `WAKATIME_API_KEY` env var exists, wakatime-cli will use its value as the api key.
This means you don’t need a `~/.wakatime.cfg` file, or you can omit or leave empty the `api_key` setting in your config file if using the env var.
However, if an api key exists in your `~/.wakatime.cfg` file then it takes precedence over the env var.

### Git Section

| option              | description | type | default value |
| ---                 | ---         | ---  | ---           |
| submodules_disabled | It will be matched against the submodule path and if matching, will skip it. | _bool_;_list_ | false |

### Git Submodule Project Map Section

A key value pair list separated by new line, where the value before equal sign is the regex pattern and the latter is the override project name for submodule. Use when a submodule project should be renamed to another when sent to the API. The order of the git submodule project map is important, the first match is used. Do not add any leading space before the regex pattern.

```ini
[git_submodule_projectmap]
some/submodule/name = new project name
^/home/user/projects/bar(\d+)/ = project{0}
```

For commonly used configuration options, see examples in the [FAQ](https://wakatime.com/faq).

## Internal INI Config File

The plugins and wakatime-cli use a separate internal INI file for things like caching auto-update requests to the GitHub releases API, and exponential backoff to the WakaTime API.
The default internal INI config file location is `$WAKATIME_HOME/.wakatime/wakatime-internal.cfg`.
