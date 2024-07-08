# Running ansible playbooks

INVENTORY = inventory

# Default target
all: playbooks

# All
playbooks:
	ansible-playbook main-playbook.yml -i $(INVENTORY)

# Role: Primary
nexus:
	ansible-playbook roles/primary/tasks/nexus.yml -i $(INVENTORY)

prometheus:
	ansible-playbook roles/primary/tasks/prometheus-install.yml -i $(INVENTORY)

# Role: Workers
k3s-agent:
	ansible-playbook roles/workers/tasks/install-k3s-agent.yml -i $(INVENTORY)

# Role: Common
docker:
	ansible-playbook roles/common/tasks/docker-install.yml -i $(INVENTORY)

ssh:
	ansible-playbook roles/common/tasks/ssh-setup.yml -i $(INVENTORY)
