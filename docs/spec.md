# World Spec

A world is a means to present as a unit a set of components required to run DCL.
It cares about unifying otherwise isolated components but not about the
deployment of each one of them, this means two different worlds may use some of
the same components.

# Why worlds?

- In order to keep costs cheaper and to foster decentralization, we’re setting
up a mechanism that, enabling clusterization and filtering, could scale to
millions of users.

- Regional instances: To enhance the perceived performance (latency), the
  language of the users and possibly some law regulation, worlds can be
  clusterized by region. All major MMO games do this.

- Optimizes usage of resources Instances can vary in size allowing small deployments or huge clusters.

- There may be specialized instances

- Works for any parcel: Any parcel with a game server is able to recognize the
  world instance the user is connected to. That enables the same experience to
  work seamlessly in different regions with different users.

- Enables decentralization: Even if only one cluster is working somewhere, no
  matter who owns it, decentrand will continue working.

## Components

- World definition: public facing API that describes the world and it's components.
- World communication: this component is in charge of allowing users to see and
  interact with each other though public chat while exploring decentraland.
- Identity service: this component is in charge of authenticating users and
  providing a unique way of identifying users.
- Profile service: this component will serve as a way for users to backup their own profile.

## Word Definition Service

Public facing API, no authentication required.

### API

GET `/description`

Returns a json with the following schema:

```
{
    name: "",
    description: "",
    communication: {
        url: ""
    },
    identity: {
        url: ""
    },
    profile: {
        url: ""
    }
}
```

## Word Communication

This component is in charge of relaying information between clients. We use it to sends positions, public chat and other kinds of messages to nearby users.

It has two sub components: a coordinator and a set of communication servers. We use [webrtc-broker library](https://github.com/decentraland/webrtc-broker) for this work. 

- The communication server will not known about the specifics of the message content.
- Clients will subscribe to a set of topics, and only receive messages from topics they are subsribed to.
- Clients will send messages, and each message will include a topic.

Here is an extract from the library's doc:

### Components

![](docs/diagram.png?raw=true)

#### Clients

The whole point of this system is to provide connectivity to the clients. The
clients connect to a single communication server after credentials negotiation
with the Coordinator Server.

#### Coordinator Server

The coordinator server is the key entry point of our communications system.

- It exposes a WS endpoint for the clients to negotiate the communication with the communications server
    - Caveat: It’s very important to notice the coordinator server will choose a
      communication server randomly, which means all communication servers
      should be equivalent to each other. That is, a client connected to a
      cluster, should have a consistent latency no matter which communication
      server he ends connecting to.
- It exposes a WS endpoint for the communication servers to:
    - Negotiate connections with the clients
    - Discover others communication server in the cluster

#### Communication server

The communications server is the heavy lifter of the system.

- It connects to the Coordinator Server via WS and uses this connection to discover other servers and negotiate client connections
- Every communication server may be connected to clients and
    - It relays packets from the clients to other clients
    - It relays packets to all the connected servers
    - It handles the business logic of the packets (topics, etc)
- It has to keep the WS connection alive, always. If the connection is closed, it has to retry until success.
