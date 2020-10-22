# Remote Log Collector

Collect logs from a remote service locally! Depending on your server resources, adding a logging
 stack to an instance might require a beefier server than you have or are willing to pay for. This
 lets you collect logs from a remote server to ingest them locally in the Elastic stack on a local
 computer you own, cutting down on costs.

## Configuration

This service polls for logs from the endpoints you configure, the endpoints should return
 an array of JSON format logs since the last time it was polled.

Create a `config.json` file that looks like this:

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

Even with bearer token authorization you should still be careful to only allow access to that
 `/logs` endpoint on every server behind a firewall.

Start with `docker-compose up --build -d`

You can then access Kibana at http://localhost:5601