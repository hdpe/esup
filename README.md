# esup

Elasticsearch schema migrator, inspired by Flyway, 
Liquibase etc., but as a standalone executable.

**Features:**

* Single-command  migration to a schema stored in
  flat config files (e.g. checked into source control -
  ideal for CI)
* Config file format mirrors request format of Elasticsearch
  REST API
* Enforces index access via aliases to facilitate atomic
  mapping changes
* Support for multiple "environments" on single ES cluster
  from single config file directory tree, with overridable 
  resource definitions, via resource prefixes
* Automatic reindexing from old to new indices
* Support for ingest pipelines in reindexing
* Includes for common configuration across resources 
  or environments

## Usage

```
$ esup migrate ENVIRONMENT
```

## Example

### Directory structure

```
schema/
├── esup.config.yml
├── includes/
├── indexSets/
│   ├── index1-default.json
│   ├── index2-default.json
│   ├── index2-default.meta.yml
│   └── index2-dev.json
└── pipelines/
    └── pipeline1-default.json
```

### Invocation

```
$ esup migrate dev
```

### Result

```
Planned changes:
- put pipeline dev-pipeline1
- create index dev-index1_20201207104530
- create alias dev-index1 -> dev-index1_20201207104530
- write index set changelog entry for index1
- create index dev-index2_20201207104530
- create alias dev-index2 -> dev-index2_20201207104530
- write index set changelog entry for index2

Confirm [Y/n]: Y
Complete
```


### Resources

Resource paths are of the format

`{resourceType}/{resourceIdentifier}-{environmentSelector}.json`

with attached meta

`{resourceType}/{resourceIdentifier}-{environmentSelector}.config.yml`

On execution, up to one resource and up to one meta are resolved
for each resource type and identifier with the following precedence:

* a file whose `environmentSelector` exactly matches the `ENVIRONMENT`; otherwise
* a file whose `environmentSelector` is `default`; otherwise
* nothing.

For the directory structure in the example above,
`esup migrate dev` would resolve the files `index1-default.json`,
`index2-dev.json`, `index2-default.meta.yml` 
and `pipeline1-default.json`.

## Resources

### Index Set

Corresponds to a timestamped _index_ and an associated _alias_.
Updates will create a new index, reindex to the new index and 
update the alias.

#### Resource

`{indexSetName}-{environment}.json`

Request body of Elasticsearch [Create index API](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-index.html).

```json
{
  "settings": { ... },
  "mappings": { ... }
}
```

#### Meta

`{indexSetName}-{environment}.meta.yml`

```yaml
index: ...
prototype:
  disabled: ...
  maxDocs: ...
reindex:
  pipeline: ...
```

|Key|Type|Description|Default|
|---|---|---|---|
|index|string|statically point the alias at this exact index, instead of managing a timestamped index set 
|prototype.disabled|bool|don't reindex documents from prototype environment on first index creation|`false`|
|prototype.maxDocs|int|only reindex this many documents from prototype environment on first index creation: `-1` reindexes all documents|`-1`|
|reindex.pipeline|string|ingest pipeline to use in reindexing||


### Pipeline

Corresponds to an _ingest pipeline_.

#### Resource

`{pipelineId}-{environment}.json`

Request body of Elasticsearch [Put pipeline API](https://www.elastic.co/guide/en/elasticsearch/reference/current/put-pipeline-api.html).

```json
{
  "description" : ...,
  "processors" : [ ... ]
}
```

#### Meta

No meta parameters supported.

## Includes

Simple resource includes are suppported via Go templates 
and the `include` function:

```
{{{ include 'NAME' }}}
```

We use triple braces `{{{` because double braces are common in
Elasticsearch ingest processor configuration. 
Includes are searched for in the `includes` directory
and expected to have a `.json` extension.

### Example

`includes/common-fields.json`

```json
"common1": {
  "type": "text"
},
"common2": {
  "type": "text"
}
```

`indexSets/index-default.json`

```json
{
  "mappings": {
    "field1": {
      "type": "text"
    },
    {{{ include "common-fields" }}}
  }
}
```

resolves to:

```json
{
  "mappings": {
    "field1": {
      "type": "text"
    },
    "common1": {
      "type": "text"
    },
    "common2": {
      "type": "text"
    }
  }
}
```

## Configuration

`esup.config.yml`

All properties overridable via environment variable.

```yaml
server:
  address: ...
  apiKey: ...
indexSets:
  directory: ...
pipelines:
  directory: ...
preprocess:
  includesDirectory: ...
```

|Key|Env Var|Type|Description|Default|
|---|---|---|---|---|
|server.address|SERVER_ADDRESS|string|address of Elasticsearch server|`"http://localhost:9200"`|
|server.apiKey|SERVER_APIKEY|string|api key for server access||
|prototype.environment|PROTOTYPE_ENVIRONMENT|string|reindex all new index sets from corresponding index in this environment||
|indexSets.directory|INDEXSETS_DIRECTORY|string|directory containing index set resources|`"./indexSets"`|
|pipelines.directory|PIPELINES_DIRECTORY|string|directory containing pipeline resources|`"./pipelines"`|
|preprocess.includesDirectory|PREPROCESS_INCLUDESDIRECTORY|string|directory containing resource includes|`"./includes"`|

## Development

Bring up local elasticsearch:

```
docker run --rm -p 9200:9200 -e 'discovery.type=single-node' elasticsearch:7.9.3
```