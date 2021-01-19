# UPP - Concept RW Elasticsearch

Concept RW ElasticSearch is an application that writes concepts into Amazon Elasticsearch cluster in batches. It is also used to update concepts metrics as part of a regular cron job that keeps concepts search up to date.

## Code

up-crwes

## Primary URL

<https://upp-prod-delivery-glb.upp.ft.com/__concept-rw-elasticsearch>

## Service Tier

Bronze

## Lifecycle Stage

Production

## Delivered By

content

## Supported By

content

## Known About By

- ivan.nikolov
- marina.chompalova
- miroslav.gatsanoga
- elitsa.pavlova
- kalin.arsov

## Host Platform

AWS

## Architecture

Concept RW ElasticSearch provides read/write access for Concepts Elasticsearch cluster. The service is deployed in both Delivery EU and US with two replicas per region.

## Contains Personal Data

No

## Contains Sensitive Data

No

## Dependencies

- concept-elasticsearch-cluster

## Failover Architecture Type

ActiveActive

## Failover Process Type

FullyAutomated

## Failback Process Type

PartiallyAutomated

## Failover Details

See the [failover guide](https://github.com/Financial-Times/upp-docs/tree/master/failover-guides/delivery-cluster) for more details.

## Data Recovery Process Type

NotApplicable

## Data Recovery Details

The service does not store data, so it does not require any data recovery steps.

## Release Process Type

PartiallyAutomated

## Rollback Process Type

Manual

## Release Details

The release is triggered by making a Github release which is then picked up by a Jenkins multibranch pipeline. The Jenkins pipeline should be manually started in order for it to deploy the helm package to the Kubernetes clusters.

## Key Management Process Type

Manual

## Key Management Details

In order to get access to ElasticSearch the service uses credentials from the `content-containers-apps` AWS IAM user. The credentials are rotated by following a standard manual procedure.

## Monitoring

Pod health:

- <https://upp-prod-delivery-eu.ft.com/__health/__pods-health?service-name=concept-rw-elasticsearch>
- <https://upp-prod-delivery-us.ft.com/__health/__pods-health?service-name=concept-rw-elasticsearch>

## First Line Troubleshooting

<https://github.com/Financial-Times/upp-docs/tree/master/guides/ops/first-line-troubleshooting>

## Second Line Troubleshooting

Please refer to the GitHub repository README for troubleshooting information.
