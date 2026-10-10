# traefik-empty-gif-plugin

Traefik middleware equivalent to nginx `empty_gif`. Responds to `GET`/`HEAD` on the configured paths with a 43-byte 1x1 transparent GIF (other methods get 405). Requests to other paths are passed to the next handler untouched.

Intended for local plugin use only. It is not published to the Plugin Catalog.

## Usage (local plugin)

Place this directory at `<traefik cwd>/plugins-local/src/github.com/suer/traefik-empty-gif-plugin`.

Static config:

```yaml
experimental:
  localPlugins:
    emptygif:
      moduleName: github.com/suer/traefik-empty-gif-plugin
```

Dynamic config:

```yaml
http:
  routers:
    app:
      rule: PathPrefix(`/`)
      service: app
      middlewares: [emptygif]
  middlewares:
    emptygif:
      plugin:
        emptygif:
          paths:
            - /favicon.ico
```

`paths` is required and matched exactly against the request path. Other paths go to the router's service (`app` above, defined elsewhere).
