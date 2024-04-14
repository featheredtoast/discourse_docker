# Launcher2

Build and run discourse images. Drop in replacement for launcher the shell script.

## Changes from launcher

No prereqs are checked. It assumes you have docker set up and whatever minimum requirements setup for Discourse: namely a recent enough version of docker, git.

## Migration from launcher

Some things are not implemented from launcher1.

* `DOCKER_HOST_IP` - container can use `host.docker.internal` in most cases. Supported on mac and windows... can also be [added on linux via docker args](https://stackoverflow.com/questions/72827527/what-is-running-on-host-docker-internal-host).
* stable `mac-address` - not implemented.

No debug. Under the hood, launcher2 uses docker build which does not allow images to be saved along the way.

## New features

### Separates bootstrap process into distinct build, configure, and migrate steps.

Separating the larger bootstrap process into separate steps allows us to break up the work. This is multipronged.

#### Easier creation for prebuilt docker images

Share built docker images by only running a build step - this build step does not need to connect to a database.
It does not need postgres or redis running. This makes for a simple way to install custom plugins to your Discourse image.

The resulting image is able to be used in Kubernetes and other docker environments.

This is done by deferring finishing the build step, to a later configure step -- which boostraps the db, and precompiles assets.
The configure and migrate steps can now be done on boot through use of env vars set in the `app.yml` config: `CREATE_DB_ON_BOOT`, `MIGRATE_ON_BOOT`, and `PRECOMPILE_ON_BOOT`

#### Adds support to *when* migrations are run

Build and Configure steps do not run migrations, allowing for external tooling to specify exactly when migrations are run.

Migrate, bootstrap, and rebuilt steps do run migrations.

#### Adds support for *how* migrations are run: `SKIP_POST_DEPLOYMENT_MIGRATIONS` support

Migrate (and consequently bootstrap) commands expose env var to turn on separate post deploy migration steps.

Allows the ability to turn on and skip post migration steps from launcher when running a migrate step.

#### Minimize downtime on rebuilds

Both standalone and multi-container setups' downtime have been minimized for rebuilds

On standalone builds, only stop the running container after the base build is done.
Standalone sites will only need to be offline during migration and configure steps.

On multi-container setups or setups with a configured external database, attempt to run migrations without stopping the container.
Multi-container app restarts trigger after migration and configuration steps are run.

#### Serve offline page during downtime

Adds the ability to build and run an image that finishes a build on boot, allowing the server to display an offline page.
For standalone builds above, this allows for the accrued downtime from migration and configure steps to happen more gracefully.

Additional env vars get turned on in `offline-page.template.yml`:
  * `CREATE_DB_ON_BOOT`
  * `MIGRATE_ON_BOOT`
  * `PRECOMPILE_ON_BOOT`

These allow containers to boot cleanly from a cold state, and complete db creation, migration, and precompile steps on boot.

During this time, nginx can be up which allows standalone builds to display an offline page.

Use of these variables may also be used for other applications where more bootstrapping is needed for alternative deployments such as in Kubernetes.

### Multiline env support

Allows the use of multiline env vars so this is valid config, and is passed through to the container as expected:
```
env:
  SECRET_KEY: |
    ---START OF SECRET KEY---
    123456
    78910
    ---END OF SECRET KEY---
```

### Killable

Launcher wraps docker run commands, which run as children in process trees. Launcher2 does the same, but attempts to kill or stop the underlying docker processes from interrupt signals.

### Docker compose generation.

Allows easier exporting of configuration from discourse's pups configuration to a docker compose configuration.

### Autocomplete support

Run `source <(./launcher2 sh)` to activate completions for the current shell, or add the results to your dotfiles

Autocompletes commands, subcommands, and suggests `app.yml` config files from your containers directory. Having a long site name should not feel like a pain to type.
