# Butler

Butler came out of frustration with a local Jenkins setup and from curiosity: how hard can it be to make a build server?

## Running

Choose a directory to host projects and run `butler` there.

## Adding a project

Create a directory `projects/<projectname>/src` and put the checked out source code there (so that the path `projects/<projectname>/src/.git` exists). The new project will be discovered and the builds will start automatically.

## Getting the builds

The builds are served over HTTP at the address `http://localhost:8080/<projectname>`.

On disk the builds are stored in the `projects/<projectname>/builds` directory and grouped by branches. For example, builds from master branch are put in `projects/<projectname>/master/`. Commits with version tags (like "1.1.0") are treated specially, their builds are stored in the `projects/<projectname>/releases` directory.

## Passing environment variables to build commands

To define additional environment variables for a project create a file `projects/<projectname>/.env`. For example, a .env file for an Android project might look like this (provided its build script inspects these variables):

```
ANDROID_KEYSTORE_FILE=/home/butler/Android/keystore.jks
ANDROID_KEYSTORE_PASSWORD=123456
ANDROID_KEY_ALIAS=debug
ANDROID_KEY_PASSWORD=123456
```

## Specifying multiple build variants

To have multiple build variants for the same source add a `butler.json` file to the source root and commit it. The file might look like this:

```json
{
  "versions": {
    "dev": {
      "env": {
        "ENVFILE": ".env",
        "RPC": "https://rpc.dev.myproject.com"
      }
    },
    "staging": {
      "env": {
        "ENVFILE": ".env.staging",
        "RPC": "https://rpc.myproject.com"
      }
    }
  }
}
```

In this example two variants are specified, called "dev" and "staging", each with its own environment variables.
