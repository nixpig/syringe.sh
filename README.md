# 💉 syringe.sh

## Proposed API

Allow omission of any arguments which can be inferred. For example, if current project is already linked then allow omission of `--project` and `--env` arguments when setting a variable.

| Command (Alias) | Arguments                                          | Description                                                                                  | Example                                    |
| --------------- | -------------------------------------------------- | -------------------------------------------------------------------------------------------- | ------------------------------------------ |
| `install` (`i`) | `--config-dir`, `--db-only`, `--config-only`       | Create config directory and database, if doesn't exist.                                      | `syringe install`                          |
| `link` (`l`)    | `--project` (`-p`), `--env` (`-e`)                 | Link up current directory. Implicitly creat project and environment, if don't already exist. | `syringe link`                             |
| `unlink` (`u`)  |                                                    | Remove link between current directory and project/environment.                               | `syringe unlink`                           |
| `set` (`s`)     | `--project` (`-p`), `--env` (`-e`), `--var` (`-v`) | Set environment variable. Implicitly create project/env, if don't already exist.             | `syringe set --var DB_PASSWORD=p4ssw0rd`   |
| `get` (`g`)     | `--project` (`-p`), `--env` (`-e`), `--var` (`-v`) | Get environment variable.                                                                    | `syringe get --var DB_PASSWORD`            |
| `current` (`c`) |                                                    | Get details of current link to project/env.                                                  | `syringe current`                          |
| `all` (`a`)     |                                                    | Get details of all links of all projects to system.                                          | `syringe all`                              |
| `remove` (`r`)  | `--project` (`-p`), `--env` (`-e`), `--var` (`-v`) | Remove a project/environment/variable from database.                                         | `syringe remove --project dunce --env dev` |
