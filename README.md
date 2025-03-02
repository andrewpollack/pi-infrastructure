# Pi Infrastructure

Repository for managing infrastructure as code and containerized applications in
a k3s cluster deployed on a Raspberry Pi setup. Experimenting with lightweight
Kubernetes and localized development.

| ![image](https://github.com/user-attachments/assets/3b9e1048-1b2b-49bd-9887-1338c3262ce7) |
| :---------------------------------------------------------------------------------------: |

## What's hosted?

### [meals-go](https://github.com/andrewpollack/pi-infrastructure/tree/main/containers/meals-go) + [meals-frontend](https://github.com/andrewpollack/pi-infrastructure/tree/main/containers/meals-frontend)

I got tired of having to pick what to eat for dinner each month, thus
[meals-go](https://github.com/andrewpollack/pi-infrastructure/tree/main/containers/meals-go)
was born.
[meals-go](https://github.com/andrewpollack/pi-infrastructure/tree/main/containers/meals-go)
selects and renders a month's worth of recipes, including related grocery list.
Executes in two modes: frontend+backend service deployment, and email cronjob.

| <img width="1241" alt="image" src="https://github.com/user-attachments/assets/438e51af-75c5-4d3b-bcf5-1834bba4d36a" /> |
| :---------------------------------------------------------------------------------------------------------------------: |

#### Meals Services:

| <img width="675" alt="image" src="https://github.com/user-attachments/assets/98329cdc-2b0a-424b-b436-9badc10857a4" /> |
| :---------------------------------------------------------------------------------------------------------------------: |

Services on k3s using Deployments and a NodePort. This setup allows
viewing this month's meals from my phone, laptop, or tablet by hitting a Tailscale
URL.

* **[meals-frontend](https://github.com/andrewpollack/pi-infrastructure/tree/main/containers/meals-frontend)** written in Svelte, enables a few different functionalities:
  * Current month calendar view,
  * Enable/Disable meals form,
  * Trigger email form to trigger the same workflow as next section.
* **[meals-go](https://github.com/andrewpollack/pi-infrastructure/tree/main/containers/meals-go)** backend written in GoLang and served using [Gin](https://github.com/gin-gonic/gin).
* **meals-updater-cronjob** A scheduled CronJob synchronizes the Postgres database with the source of truth JSON file for recipes stored in S3. Recipe data is maintained in a private repository shared with my partner, where a CI/CD pipeline validates updates, with a post submit that pushes the latest JSON state to S3. As a result, updating recipes is as simple as opening a PR.
* **email-cronjob** every Thursday, finds next week's recipes and compiles a grocery list for all items, combining like items by quantity, and assigning each item to its respective aisle. This is then formatted and emailed
to me and my partner. Email is sent using [Amazon Simple Email Service](https://aws.amazon.com/ses/)

| <img width="742" alt="Screenshot 2025-03-01 at 6 16 48 PM" src="https://github.com/user-attachments/assets/3ff39709-fbd2-470e-86fd-75d7b4c74b38" /> |
| :---------------------------------------------------------------------------------------: |

---

### [garage-go](https://github.com/andrewpollack/pi-infrastructure/tree/main/containers/garage-go)

I (too often) leave for work, round the corner, and immediately second guess
whether I closed the garage door. So, I fixed this concern.

This Garage door monitoring system uses a Raspberry Pi 2 W hooked up with a
magnetic door sensor. The magnetic door sensor detects when the garage door is
in the "closed" position. Metrics are exposed via a Prometheus exporter, which
are scraped by a Prometheus deployment I'm using across these projects. Finally,
a dashboard is available to view the yes/no question through Grafana.

| ![image](https://github.com/user-attachments/assets/5e800534-3013-4760-a3e6-b6288c50d657) |
| :---------------------------------------------------------------------------------------: |

_Note: Getting started with the Hardware portion, particularly reading the state
of the GPIO pin, was super doable thanks to this book
<ins>[Automate Your Home Using Go](https://pragprog.com/titles/gohome/automate-your-home-using-go/)</ins>_

---

## Setup

Ansible install: `$ sudo apt install ansible-core`

Terraform install:
https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli#install-terraform

Required non-checked files:

- `ansible/inventory/pi`:

  ```
  [common]
  <FILL IN>

  [artifactors]
  <FILL IN>

  [all:vars]
  ansible_python_interpreter=/usr/bin/python3

  ```

- `ansible/roles/common/vars/vars.yml`:
  ```
  pub_ssh_key_path: <FILL IN>
  ```
- `ansible/roles/workers/vars/vars.yml`:
  ```
  k3s_token: <FILL IN>
  k3s_url: <FILL IN>
  ```
- `terraform/nexus/local.tfvars`:
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
