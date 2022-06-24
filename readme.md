# Harbinger

Collect logs from a remote service locally! Depending on your server resources, adding a logging stack to an instance might require a beefier server than you have or are willing to pay for. This lets you collect logs from a remote server to ingest them locally in the Elastic stack on a local computer you own, cutting down on costs.

There are two parts to this project. The first is the remote log collector, the second is a service ("Harbinger") that monitors site uptime and reports that (as well as the logs collected by the remote log collector) to Discord using webhooks.

## Remote Log Configuration

This service polls for logs from the endpoints you configure, the endpoints should return
 an array of JSON format logs since the last time it was polled.

Create a `src/log-collector/config.json` file that looks like this:

```json
{
  "servers": [{
    "path": "https://example.com/logs",
    "bearer": "an optional bearer token to authenticate to your log endpoint",
    "name": "an ES index name to create/use for this service"
  }]
}
```

Configured names need to follow the [Elasticsearch naming restrictions for indices](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-index.html#indices-create-api-path-params).

Even with bearer token authorization you should still be careful to only allow access to that `/logs` endpoint on every server behind a firewall.

You can access Kibana at http://localhost:5601

## Harbinger Configuration

Harbinger is configured by a file you create at `./src/harbinger/config.json`. It should look like this:

```json
{
  "services": [{
    "serviceName": "my-site",
    "displayName": "My Site",
    "endpoint": "https://example.com/health",
    "webhooks": ["a discord webhook url", "another webhook url if you want to notify multiple channels or servers"]
  }],
  "harbinger": {
    "name": "a name for this instance of harbinger",
    "webhook": "a discord webhook url"
  }
}
```

It's recommended to create separate webhooks for each service you wish to monitor, as well has Harbinger itself so you can customize the image and name in Discord.

Health checks call your configured endpoint every five minutes and look for a 200 status response (response body is ignored) and will notify you over Discord if a site responds with any other status code.

For added monitoring it's a good idea to run separate instances of Harbinger on two machines that monitor something on each other's machine. You should monitor services from a different machine than the one the services run on so if the entire machine goes down Harbinger doesn't go down with it and you're left without monitoring.

## Log message schema

Log messages are generally expected to have a schema like this:

```typescript
interface LogMessage {
    // the name of the service this originated from, should match a serviceName in the harbinger config
    service: string;
    // the part of the app this is from, e.g. app, http, auth, etc.
    concern: string
    // the actual message logged
    message: string;
}
```

If your log messages have a `service` string that matches a `serviceName` in the Harbinger config it will use that service's webhook to send the message, otherwise it will use the general Harbinger webhook.

## Run

Run with `docker-compose -f <a docker compose file> up --build -d`. `docker-compose.harbinger.yml` just runs Harbinger, `docker-compose.all.yml` runs Harbinger as well as the remote log collector and the Elastic Stack.

