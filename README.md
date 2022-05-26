# Demo for Schema Migration Tools

## Introduction

Schema Migration is the version control tool for database. It helps you manage your graph schemas as code. 

Two popular tools (support most of common relational database systems):

* Flyway: [website](https://flywaydb.org/), [repo](https://github.com/flyway/flyway)
* Liquibase: [website](https://www.liquibase.org/), [repo](https://github.com/liquibase/liquibase), [Neo4j extension](https://liquibase.jira.com/wiki/spaces/CONTRIB/pages/2936537089/Neo4J+Extension)

## How to run this demo tool?

_Note: you should install tigergraph & golang first._

_*To be careful: migration would drop all resources first!!!*_

```sh
# migrate to specified version, say V1, V2, V3...
go run main.go -g [version]

# rollback to specified version (currently, just supporting the latest one)
go run main.go -r [version]
```
