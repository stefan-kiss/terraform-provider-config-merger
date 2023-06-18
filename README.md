<!-- TOC -->
* [Terraform Provider Config Merger](#terraform-provider-config-merger)
  * [how to use](#how-to-use)
  * [yaml merging engine](#yaml-merging-engine)
* [Security](#security)
<!-- TOC -->
# Terraform Provider Config Merger

_Please see the security section. There is no obvious insecure part, however currently functionality was considered over security._

This is a Terraform provider for merging configuration files.
The end-goal is to provide a way to configure `organisation defaults` while still allowing a high degree of per `environment customisation`.

The organising of the files is inspired by hiera. However there is at least one key difference:

- facts are not sent to the provider but discovered by it using the path to the most specific file. facts are then injected into the configuration inside the configured key.

## how to use

- first you need to define a hierarchical structure that reflects your environment
  - easiest way to envision this is as a directory structure
    - ```shell
        config
        ├── development
        │   └── us-east-2
        │         └── s3bucket
        └── production
            └── us-east-2
            │      └── s3bucket
            └── us-west-2
                   └── s3bucket
      ```

The goal here is to describe the most common configuration in the top level directories and then override it in the lower level directories.
The provider gets several inputs:

- a description of the structure
  - including the names of the keys where the discovered values are added
- the most specific path to the configuration file
- the `glob` patterns of what is considered a configuration file
 

```terraform
terraform {
  required_providers {
    merger = {
      source  = "registry.terraform.io/stefan-kiss/config-merger"
      version = "=1.0.0"
    }
  }
}

provider "config-merger" {
  project_config = "config/{{facts.environment}}/{{facts.region}}/{{facts.project}}"
  config_globs = [
    "config.yaml",
    "*.config.yaml",
  ]
}

data "config-merger_merged" "example" {
  config_path = "config/production/us-west-2/s3bucket"
}
```

The above configuration will look for all files matching the glob patterns in each of the directories down the path starting with the top one:
- `config/(config.yaml|*.config.yaml)`
- `config/production/(config.yaml|*.config.yaml)`
- `config/production/us-west-2/(config.yaml|*.config.yaml)`
- `config/production/us-west-2/s3bucket/(config.yaml|*.config.yaml)`

Keys found in files on a lower level will always override keys found in files on a higher level.

On top of that the result wil also include the `facts` that were discovered. Each fact will stay in it's own key as indicated by the directory strucure.

```shell
config/production           /us-west-2       /s3bucket
            |                     |                |  
config/{{facts.environment}}/{{facts.region}}/{{facts.project}}
    
```

keys added to the result: 
```yaml
facts:
  environment: production
  region: us-west-2
  project: s3bucket
```


## yaml merging engine

yaml merging is done using spruce with the default options:
[https://github.com/geofffranks/spruce](https://github.com/geofffranks/spruce)

Spruce is implemented as a library, so there is no need to have spruce installed. This also allows this provider to work with terraform enterprise.

# Security

**The provider is intended to work with yaml files that are fully under your control.**
**Allowing end-users/external users to input yaml into the config tree may have serious security implications.**
**On top of any processing and potential exploits, this engine also allows integration with vault. (so your secrets are potentially exposed)** 

**See details here: [https://github.com/geofffranks/spruce/blob/main/doc/pulling-creds-from-vault.md](https://github.com/geofffranks/spruce/blob/main/doc/pulling-creds-from-vault.md)**

Please use this software only if you fully understand how it works and what are the implications.
