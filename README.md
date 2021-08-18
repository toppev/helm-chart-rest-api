# helm-chart-http-api

A simple HTTP API for installing kubernetes [Helm](https://helm.sh/) charts.  
The API can be secured with simple Basic Auth authentication.

## Deployment
See `./kubernetes-examples` for an example kubernetes deployment.  
The app listens on port `8080` by default.

## Building
Build the Docker image with `docker build . -t helm-api`.

## Endpoints

### Install a new chart
POST `/start-chart`
Example body:
```json
{
    "releaseName": "helm-api-test",
    "values": {
        "someValues": "values.yaml as json",
    }
}
```

### Uninstall a chart
DELETE `/uninstall-chart?release=helm-api-test`
