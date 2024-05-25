# 💉 syringe.sh

## Proposed API

Allow omission of any arguments which can be inferred. For example, if current project is already linked then allow omission of `--project` and `--env` arguments when setting a variable.

| Command (Alias)  | Arguments                                                | Description                                                                                   | Example                                    |
| ---------------- | -------------------------------------------------------- | --------------------------------------------------------------------------------------------- | ------------------------------------------ |
| `install`<br>`i` | `--config-dir` <br>`--db-only` <br>`--config-only`       | Create config directory and database, if doesn't exist.                                       | `syringe install`                          |
| `link`<br>`l`    | `--project` (`-p`) <br>`--env` (`-e`)                    | Link up current directory. Implicitly create project and environment, if don't already exist. | `syringe link`                             |
| `unlink`<br>`u`  |                                                          | Remove link between current directory and project/environment.                                | `syringe unlink`                           |
| `set`<br>`s`     | `--project` (`-p`) <br>`--env` (`-e`) <br>`--var` (`-v`) | Set environment variable. Implicitly create project/env, if don't already exist.              | `syringe set --var DB_PASSWORD=p4ssw0rd`   |
| `get`<br>`g`     | `--project` (`-p`) <br>`--env` (`-e`) <br>`--var` (`-v`) | Get environment variable.                                                                     | `syringe get --var DB_PASSWORD`            |
| `current`<br>`c` |                                                          | Get details of current link to project/env.                                                   | `syringe current`                          |
| `all`<br>`a`     |                                                          | Get details of all links of all projects to system.                                           | `syringe all`                              |
| `remove`<br>`r`  | `--project` (`-p`) <br>`--env` (`-e`) <br>`--var` (`-v`) | Remove a project/environment/variable from database.                                          | `syringe remove --project dunce --env dev` |
