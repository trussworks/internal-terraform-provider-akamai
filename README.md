Terraform Provider for Akamai FastDNS
==================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Requirements
------------

- [Terraform](https://www.terraform.io/downloads.html) 0.10+
- [Go](https://golang.org/doc/install) 1.12 (to build the provider plugin)

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/trussworks/terraform-provider-akamai`

```sh
$ mkdir -p $GOPATH/src/github.com/trussworks; cd $GOPATH/src/github.com/trussworks
$ git clone git@github.com:trussworks/terraform-provider-akamai
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/trussworks/terraform-provider-akamai
$ make build
```

Using the provider
----------------------
If you're building the provider, follow the instructions to [install it as a plugin.](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin) After placing it into your plugins directory,  run `terraform init` to initialize it. Documentation about the provider specific configuration options can be found on the [provider's website](https://www.terraform.io/docs/providers/akamai/index.html).

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.12+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-akamai
...
```

In order to test the provider, you can simply run `make test`.


```sh
$ make test
```

If you need to add a new package in the vendor directory under `github.com/trussworks/akamai-sdk-go`, create a separate PR handling _only_ the update of the vendor for your new requirement. Make sure to pin your dependency to a specific version, and that all versions of `github.com/trussworks/akamai-sdk-go/*` are pinned to the same version.
