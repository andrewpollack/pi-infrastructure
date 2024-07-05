# Pi Infrastructure
Checkins for a raspberry pi setup I'm playing around with.

## Setup
Ansible install:
`$ sudo apt install ansible-core`

Terraform install:
https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli#install-terraform

Required non-checked files:
* `ansible/inventory/pi`:
    ```
    [piservers]
    <FILL IN>

    [artifactors]
    <FILL IN>

    [all:vars]
    ansible_python_interpreter=/usr/bin/python3

    ```
* `ansible/roles/common/vars/vars.yml`:
    ```
    pub_ssh_key_path: <FILL IN>
    ```
* `terraform/nexus/local.tfvars`:
    ```
    nexus_url = <FILL IN>
    nexus_username = <FILL IN>
    nexus_password = <FILL IN>
    nexus_docker_repo_port = <FILL IN>
    ```
