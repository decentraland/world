# World Spec

A world is a means to present as a unit a set of components required to run DCL. It cares about unifying otherwise isolated components but not about the deployment of each one of them, this means two different worlds may use some of the same components.

## Components

- World definition: public facing API that describes the world and it's components.
- World communication: this component is in charge of allowing users to see and interact with each other though public chat while exploring decentraland.
- Identity service: this component is in charge of authenticating users and providing a unique way of identifying users.
- Profile service: this component will serve as a way for users to backup their own profile.

## Word Definition Service

GET `/description`

Returns a json with the following schema:

```
{
	name: ""
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
