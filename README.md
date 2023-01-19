## civo-opencontrolplane
The Civo OpenControlPlane is the implematation of the OpenControlPlane specification for Civo.

## What is the OpenControlPlane?
The OpenControlPlane is a specification for a control plane that is designed to be portable across multiple cloud providers. It is designed to be a simple, easy to use, and easy to implement.

## How to run the OpenControlPlane
You can run the Civo OpenControlPlane by running the following command, which will start the OpenControlPlane on port 8080.
also you need pass the ENV variable REGION to the container, this is the region you want to manage.

```
docker run -d -p 8080:8080 -e REGION=lon1 civo/opencontrolplane
``` 


