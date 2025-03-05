# Pi Infrastructure

Managing infrastructure as code and containerized applications on
a k3s-managed Raspberry Pi cluster dedicated to home automation.

<table>
  <tr>
    <th>Pi Diagram</th>
  </tr>
  <tr>
    <td align="center">
      <img width="600" alt="Pi Diagram" src="https://github.com/user-attachments/assets/9d8c3f3f-555b-4c0e-8bbd-aa6bda790870" />
    </td>
  </tr>
</table>

## What's hosted?

### [meals-go](https://github.com/andrewpollack/pi-infrastructure/tree/main/containers/meals-go) + [meals-frontend](https://github.com/andrewpollack/pi-infrastructure/tree/main/containers/meals-frontend)

I got tired of having to pick what to eat for dinner each month, thus
[meals-go](https://github.com/andrewpollack/pi-infrastructure/tree/main/containers/meals-go)
was born.
[meals-go](https://github.com/andrewpollack/pi-infrastructure/tree/main/containers/meals-go)
selects and renders a month's worth of recipes, including related grocery list.
Executes in two modes: frontend+backend service deployment, and email cronjob.

<table>
  <tr>
    <th>Meals Website</th>
  </tr>
  <tr>
    <td align="center">
      <img width="800" alt="Meals Website" src="https://github.com/user-attachments/assets/4271bbe3-b0bd-47e8-a858-0529da652c1a" />
    </td>
  </tr>
</table>

#### Meals Services:

<table>
  <tr>
    <th>Services Architecture Diagram</th>
  </tr>
  <tr>
    <td align="center">
      <img width="600" alt="Services Architecture Diagram" src="https://github.com/user-attachments/assets/5e68173c-f729-4eac-b710-cada1a3630c1" />
    </td>
  </tr>
</table>

Services on k3s using Deployments and a NodePort. This setup allows
viewing this month's meals from my phone, laptop, or tablet by hitting a Tailscale
URL.

* **[meals-frontend](https://github.com/andrewpollack/pi-infrastructure/tree/main/containers/meals-frontend)** written in Svelte, enables a few different functionalities:
  * Current month calendar view,
  * Enable/Disable meals form,
  * Trigger email form to trigger the same workflow as next section.
* **[meals-backend](https://github.com/andrewpollack/pi-infrastructure/tree/main/containers/meals-go)** written in GoLang and served using [Gin](https://github.com/gin-gonic/gin).
* **meals-db-sync-cronjob** A scheduled CronJob synchronizes the Postgres database with the source of truth JSON file for recipes stored in S3. Recipe data is maintained in a private repository shared with my partner, where a CI/CD pipeline validates updates, with a post submit that pushes the latest JSON state to S3. As a result, updating recipes is as simple as opening a PR.
* **email-cronjob** every Thursday, finds next week's recipes and compiles a grocery list for all items, combining like items by quantity, and assigning each item to its respective aisle. This is then formatted and emailed
to me and my partner. Email is sent using [Amazon Simple Email Service](https://aws.amazon.com/ses/)

<table>
  <tr>
    <th>Example Email</th>
    <th>Example PDF Attachment</th>
  </tr>
  <tr>
    <td><img width="700" alt="Example Email" src="https://github.com/user-attachments/assets/3ff39709-fbd2-470e-86fd-75d7b4c74b38" /></td>
    <td><img width="600" alt="Example PDF Attachment" src="https://github.com/user-attachments/assets/9c038346-114e-42f9-8691-5f237abdecec" /></td>
  </tr>
</table>


---

### [garage-go](https://github.com/andrewpollack/pi-infrastructure/tree/main/containers/garage-go)

I (too often) leave for work, round the corner, and immediately second guess
whether I closed the garage door. So, I fixed this concern.

This Garage door monitoring system uses a Raspberry Pi 2 W hooked up with a
magnetic door sensor. The magnetic door sensor detects when the garage door is
in the "closed" position. Metrics are exposed via a Prometheus exporter, which
are scraped by a Prometheus deployment I'm using across these projects. Finally,
a dashboard is available to view the yes/no question through Grafana.

<table>
  <tr>
    <th>Garage Diagram</th>
  </tr>
  <tr>
    <td align="center">
      <img width="650" alt="Garage Diagram" src="https://github.com/user-attachments/assets/5e800534-3013-4760-a3e6-b6288c50d657" />
    </td>
  </tr>
</table>

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
<table>
  <tr>
    <th>Pi Setup</th>
  </tr>
  <tr>
    <td align="center">
      <img width="500" alt="Pi Setup" src="https://github.com/user-attachments/assets/8e0d6666-116b-43fb-9084-b6f32adf706a" />
    </td>
  </tr>
</table>
