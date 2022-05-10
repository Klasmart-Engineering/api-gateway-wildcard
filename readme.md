# Wildcard Plugin

By default, KrakenD expects you to explicitly state every single route that is exposed by the gateway. This is good for
a number of reasons, one of the main one's being that it's future proof. It's unknown how your API might evolve in the
future, what plugins might be needed, how you might need to reach out to multiple backends to resolve a request. Maybe
you want to add some specific caching or authorization to it.

If each of your endpoints is exposed individually, then you have complete control over when and how you make these
changes without having to worry about affecting other endpoints. However if you effectively use the gateway as a simple
proxy ie. `/foo/** -> Foo Service`, then you cannot apply configuration to independent endpoints, things either apply to
everything or nothing.

However, despite all that, sometimes there can be use cases for supporting this wildcard style routing. As such, this
plugin is the methodology to do so.

## Configuration

In order to use this plugin, you need to define two sets of configuration:

1. At the gateway level, you need to define which routes you want to expose
2. At the backend level, you need to specify that you want wildcard routing to occur

This is because this plugin hooks into the request lifecycle at two points in time, first as a server based plugin
_(intercepting the request before KrakenD processes it)_ and finally as a client plugin _(resolving the final request)_.

### Example

Lets say that we want to create wildcard routing to an endpoint `/users`. So the public API that our microgateway will expose
will be `/users/**`
- Some example endpoints might be:
  - `/users/me`
  - `/users/user/123/profile`
- These would then resolve to the final service as
  - `/me`
  - `/user/123/profile`

#### Server level configuration

At the root `extra_config` key, we need to specify that we want to expose this route:

```json
{
  "$schema": "https://www.krakend.io/schema/v3.json",
  "version": 3,
  "extra_config": {
    ...,
    "plugin/http-server": {
      "name": ["wildcard"],
      "wildcard": {
        "endpoints": {
          "/__wildcard/users": ["/users"]
        }
      }
    }
  }
  ...,
}
```

Notice how we:
1. Add `"wildcard"` to `$.extra_config.plugin/http-server.name`
2. Under the `wildcard` key, add the endpoints we want to expose _(you can add as many as you want)_
  - This is added in the format `{Internal URL (used for routing) - MUST start with /__wildcard/ }: [ Array of routes we want to expose publicly ]`
  - _Note: it is possible to define multiple wildcard routes which use the same backend_
3. Make sure you keep note of the `key` of each endpoint eg. the `/__wildcard/{path}` as we will need this later

#### Endpoint level configuration

Once we have set up what routes we want to expose publicly, we then need to set up the actual backends for these routes

In the same way you would define a normal endpoints:

```json
{
  "$schema": "https://www.krakend.io/schema/v3.json",
  "@comment": "Test configuration when working on the wildcard plugin",
  "version": 3,
  ...,
  "endpoints": [
    {
      "endpoint": "/__wildcard/users",
      "input_headers": [ "X-KidsLoop-Wildcard" ],
      "method": "GET",
      "output_encoding": "json",
      "backend": [
        {
          "method": "GET",
          "host": ["https://users.kidsloop.live"],
          "url_pattern": "",
          "extra_config": {
            "plugin/http-client": {
                "name": "wildcard"
            }
          }
        }
      ]
    },
    {
      "endpoint": "/__wildcard/users",
      "input_headers": [ "X-KidsLoop-Wildcard", "Content-Length", "Content-Type" ],
      "method": "POST",
      "output_encoding": "json",
      "backend": [
        {
          "method": "POST",
          "encoding": "json",
          "host": ["https://users.kidsloop.live"],
          "url_pattern": "",
          "extra_config": {
            "plugin/http-client": {
                "name": "wildcard"
            }
          }
        }
      ]
    }
  ]
}
```
There are 3 core things we have to do in order to correctly configure the request chain
1. The `$.endpoints.endpoint` value is the key you previously placed in the server configuration stage. _This is how we know which backends to use to resolve the request_
2. You must add the `X-KidsLoop-Wildcard` header to the `input_headers` array. The plugin uses this under the hood
3. To the `backend.[x].extra_config` key, you need to add the block below.

```json
"plugin/http-client": {
    "name": "wildcard"
}
```

A few things to take note of:
- Each endpoint will only match a single HTTP Method
  - If you want to support multiple HTTP Methods, you MUST have an endpoint defined for each one
    - _In the example above, we have support for both GET and POST requests - they're both using the same backend_
- For security reasons we explictly deny direct access to the `/__wildcard/users` endpoint, the only way to hit it is to go
  through the intended flow ie. `/users/**`


## Makefile

The [Makefile](Makefile) has a number of helper commands to get you up and running quickly. There are a number of
**protected** commands _(these are protected as they're used in build pipelines)_. If you think there is a reason to change
them specific to your use-case please contact the API management team to confirm.

For any targets that are not `PROTECTED` please free to edit and add to the `Makefile` as you wish

| Command       | Description                                                                                                                                       | Status      |
| ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- | ----------- |
| `make build`  | Builds the plugin in the intended format _<name>.so_                                                                                              | `PROTECTED` |
| `make login`  | Logs you into GitHub Container Registry - must have `GH_PAT` environment variable set in your shell - see [GitHub Personal Access Token](#GH-PAT) |             |
| `make b`      | Builds a microgateway with the configuration found in `krakend.json` - this is useful for local testing                                           |             |
| `make r`      | Runs the microgateway built by the step above                                                                                                     |             |
| `make br`     | Alias to run both the build and run commands in a single step                                                                                     |             |
| `make run-ci` | Command to run the POSTMAN integration tests in the CI environment                                                                                | `PROTECTED` |

## GH PAT

This refers to the GitHub Personal Access Token. This is a token that is specific to your own GitHub account and grants
access to GitHub resources. This PAT will need to have permissions that relate to `repos` and `packages`. You will also
need to enable SSO on the GITHUB PAT.

### Authenticating manually

1. Authenticate to the [github container registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry#authenticating-to-the-container-registry)

```sh
export GH_PAT="your personal github access token here - must have read packages scope at a minimum"

echo $GH_PAT | docker login ghcr.io -u USERNAME --password-stdin
```

## Integration Tests

We opted to use Postman as an integration testing tool, primarily because it's reasonably language agonistic and
well-known in the industry. Similarly it's useful for running both locally and in CI.

Please make sure you commit both the postman `collection` and `environment` files - _environment file should be for
CI/local development_

You can then use the [reuseable workflow](https://github.com/KL-Engineering/central-microgateway-configuration/blob/main/.github/workflows/plugin-integration-test.yaml) to easily set up a CI pipeline for your plugin.

## KrakenD Docs

[Intro to plugins](https://www.krakend.io/docs/extending/introduction/)

There are four different types of plugins you can write:

1. **[HTTP server plugins](https://www.krakend.io/docs/extending/http-server-plugins/)** (or handler plugins): They belong to the router layer and let you modify the request before KrakenD starts processing it, or modify the final response. You can have several plugins at once.
2. **[HTTP client plugins](https://www.krakend.io/docs/extending/http-client-plugins/)** (or proxy client plugins): They belong to the proxy layer and let you change how KrakenD interacts (as a client) with your backend services. You can have one plugin per backend.
3. **[Response Modifier plugins](https://www.krakend.io/docs/extending/plugin-modifiers/)**: They are strictly modifiers and let you change the responses received from your backends.
4. **[Request Modififer plugins](https://www.krakend.io/docs/extending/plugin-modifiers/)**: They are strictly modifiers and let you change the requests sent to your backends.
