# Pi Infrastructure
Checkins for a raspberry pi setup and k3s cluster I'm playing around with.

## Setup
Ansible install:
`$ sudo apt install ansible-core`

Terraform install:
https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli#install-terraform

Required non-checked files:
* `ansible/inventory/pi`:
    ```
    [common]
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
* `ansible/roles/workers/vars/vars.yml`:
    ```
    k3s_token: <FILL IN>
    k3s_url: <FILL IN>
    ```
* `terraform/nexus/local.tfvars`:
    ```
    nexus_url = <FILL IN>
    nexus_username = <FILL IN>
    nexus_password = <FILL IN>
    nexus_docker_repo_port = <FILL IN>
    ```

## What's hosted?

### [recipe-maker.py](https://github.com/andrewpollack/pi-infrastructure/tree/main/containers/meals)

![Screenshot from 2024-09-28 15-54-16](https://github.com/user-attachments/assets/92b2241f-ee41-4184-aa17-0ba6494cf091)

Quick and dirty script for randoomly selecting and rendering a months worth of recipes.
Deployed as a Deployment+NodePort Service so I can bring up the website locally from my
phone or laptop (whichever is more convenient at the time).

Recipes are stored in a JSON file in a private repository shared with my partner.
This repo has CD setup to push the JSON file's latest state to S3, which is then pulled
by this script to make sure recipes are populated. While we could just pull from the GitHub
repo itself, CD in this way is far more fun!

![Screenshot from 2024-09-28 13-27-38](https://github.com/user-attachments/assets/4b9abc7b-37e7-4730-8e1a-121b2c9d3536)

## Can I see the cluster?

Sure! Mind the dust 😹
![image_123650291](https://github.com/user-attachments/assets/8e0d6666-116b-43fb-9084-b6f32adf706a)
