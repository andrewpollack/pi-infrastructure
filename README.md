# Pi Infrastructure
Checkins for a raspberry pi setup and k3s cluster I'm playing around with.

## What's hosted?

### [recipe-maker](https://github.com/andrewpollack/pi-infrastructure/tree/main/containers/meals-go)

I got tired of having to pick what to eat for dinner each month, so thus [recipe-maker](https://github.com/andrewpollack/pi-infrastructure/tree/main/containers/meals-go)
was born. [recipe-maker](https://github.com/andrewpollack/pi-infrastructure/tree/main/containers/meals-go) selects and renders a month's worth of recipes, including related grocery list. Comes in two modes:
#### Server deployment:
Deployed on k3s using a Deployment+NodePort Service so I can view this month + next month's meals from my phone or laptop 
(whichever is more convenient at the time). Links are clickable for finding related recipes.
![Screenshot from 2024-09-28 15-54-16](https://github.com/user-attachments/assets/92b2241f-ee41-4184-aa17-0ba6494cf091)

#### Email CronJob:
Deployed on k3s using a CronJob. Every Friday, finds next week's recipes and compiles a grocery list for all items, combining like items by quantity, and assigning each item to its respective aisle. This is then formatted and emailed to me and my partner.
<img width="780" alt="image" src="https://github.com/user-attachments/assets/2e57dca2-dede-421a-b83b-1b44fb7f60d1">

#### Data:
Recipes are stored in a JSON file in a private repository shared with my partner.
This repo has CI/CD setup to validate the contents, and push the JSON file's latest state to S3. This state is then pulled
by the [recipe-maker](https://github.com/andrewpollack/pi-infrastructure/tree/main/containers/meals-go).

While we could just pull from the GitHub
repo itself, CD in this way is far more fun!
![Screenshot from 2024-09-28 13-27-38](https://github.com/user-attachments/assets/4b9abc7b-37e7-4730-8e1a-121b2c9d3536)

---

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

---

## Can I see the cluster?

Sure! Mind the dust 😹
![image_123650291](https://github.com/user-attachments/assets/8e0d6666-116b-43fb-9084-b6f32adf706a)
