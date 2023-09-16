# PubSub

**Contents**
1. [Overview](#overview)
1. [Usage](#usage)

## Overview
PubSub is a TCP-based [Pub/Sub](https://en.wikipedia.org/wiki/Publish%E2%80%93subscribe_pattern) server. It lets clients create/listen to topics and data that is published to specific topics by other clients.

<!--
For more on how this project works, visit my [portfolio](https://sidney-bernardin.github.io/project/?id=pubsub).
-->

## Usage

### Install and run

Clone this repository.
``` bash
git clone https://github.com/Sidney-Bernardin/PubSub.git
cd PubSub
```
Then run directly on your machine or with Docker.

#### Run directly on your machine.
``` bash
go run . -addr=0.0.0.0:8080
```

#### Run with Docker.
You can use the [pre-built image](https://hub.docker.com/r/sidneybernardin/pubsub) that's based on the [Dockerfile](https://github.com/Sidney-Bernardin/PubSub/blob/main/Dockerfile) at the root of this repository.

``` bash
docker run -p 8080:8080 sidneybernardin/pubsub:latest
```

### Sending Commands
Using a TCP client of your choice (like Telnet), you can send two types of commands to the server.

To create and listen to topics, use a subscribe command.

```
sub topic_a topic_b topic_c
```

To publish data to topics, use a publish command.

```
pub topic_a Hello, World!
```

All errors are returned in a JSON format and are easy to parse.

``` json
{"type":"invalid_command","detail":"Operation 'pub' requires a topic and topic-message as arguments."}
```
