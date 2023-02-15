# civo-opencontrolplane

The Civo Open Control Plane is the implementation of the [OpenControlPlane](https://www.github.com/opencontrolplane) specification for [Civo](https://www.civo.com).

## What is Open Control Plane?

Open Control Plane (OpenCP) is a specification for a control plane that is designed to be portable across multiple cloud providers. It is designed to be simple, easy to use, and easy to implement.

## How to run the Civo Open Control Plane

You can run the Civo OpenCP by running the following command, which will start the Open Control Plane on port 8080.

You would need to pass an `ENV` variable `REGION` to the container. This is the Civo region where you want to manage resources, `lon1` in the below example.

```console
docker run -d -p 8080:8080 -e REGION=lon1 civo/opencontrolplane
``` 

## Dependencies

The Civo OpenControlPlane depends on the following projects:
- [Opencp-shim](https://github.com/opencontrolplane/opencp-shim)
