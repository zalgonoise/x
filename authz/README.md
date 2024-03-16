# authz


*A gRPC+HTTP asymmetric authentication and authorization service using ES512 elliptic-curve signatures, in Go.* 

__________

## Motivation

Looking into password-less, automated solutions for authentication in web-based services, it was clear that the 
way-to-go would be with asymmetric auth, using x509 certificate chains to easily manage and distribute authorization 
tokens across remote clients.

Using public (and private) keys in authentication allows a user to publicly share their public key, and *stamp* their 
identity by signing some data with their private key. In turn, other users can verify if the sender is legitimate using 
a copy of their public key. In the context of authentication with remote systems, clients can sign up by sharing their 
public key and respond to a challenge by signing it using their private key, as they log in. The remote system ensures 
the validity of the user by verifying their response using a stored copy of their public key.

To ensure that a chain-of-trust is created, x509 certificates are used; as a root entity (Certificate Authority) exists 
to provision new authentication services with a certificate; and in turn, these services are able to provision new 
clients with their own certificate as well. By doing so, it's possible to reduce the scope of impersonation of a client,
by verifying the signature in the provided x509 certificate. A registered end-user would possess their own certificate 
as emitted by the authorization service, as well as a copy of the authorization service's certificate, as emitted by 
the(ir parent) root entity / Certificate Authority.

The usefulness of this system peaks on systems which have the root entity (Certificate Authority) and authorization 
services isolated, but still need to manage remote access to data or services, and are capable of provisioning new 
remote clients ahead-of-time.

## Architecture and structure

This type of authentication system is described in
[this Microchip article](https://developerhelp.microchip.com/xwiki/bin/view/applications/security/asymmetric-use-case-example/),
covering all the steps required to make this system possible, ensuring that:
- there is a root entity (Certificate Authority) which provisions authorization services with x509 certificates,
  identified by an (ECDSA) public key.
- there is an authorization service (or more) which is registered in the root entity and holds a valid certificate, that
  is capable of issuing auth tokens in a JWT format.
- when signing up in an authorization service, a client shares their (ECDSA) public key with the service, and in return
  receives a certificate signed by the authorization service's private key (referencing the root entity as parent), and
  a copy of the authorization service's public key and certificate (as issued by the root entity).
- when logging in, a client must verify that:
    - their copy of the authorization service's public key and certificate are legitimate and valid.
    - their public key and issued certificate are legitimate and valid.
    - they respond to a random challenge by signing it with their private key, which will be verified with the
      authorization service's stored copy of the client's public key.

## Technologies

Considering the small dimension of the handled data, the persistence layer uses `SQLite` to store service, challenge and
token data. This ensures that there are no required remote connections during startup and runtime, as well as a fast and
reliable database solution. As for the containerized, Docker solution; the database file will be within the container
(if configured with a path). The Go `SQLite` library is `modernc.org`'s, which is a CGO-free SQLite driver.

As for network, this solution uses protocol buffers, gRPC and gRPC-gateway. This allows for a simple and readable API 
definition in proto-files and code generation for all gRPC and HTTP transport logic using `buf.build`. This choice is 
out of commodity, readability, and flexibility in the choice of transport(s) to use. I am aware these are big libraries,
but personally I enjoy this API-first development approach and transport logic workflow in my projects. Makes it simple,
and it's always a pleasure to use protocol buffers and gRPC.

## Setup

## Usage

## Components

## Disclaimer

*I wouldn't use this in production!*

By any means this should not replace any other more reliable form of authentication solution, considering this is a
personal experiment on this topic. While the key algorithms and hashing algorithms aim for a secure solution, it's
always better to focus on a widely-contributed and reputable library for these kinds of topics, instead of a personal
library from a single user on GitHub.