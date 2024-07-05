# Running ansible playbooks

INVENTORY = inventory

# Default target
all: playbooks

# All
playbooks:
	ansible-playbook main-playbook.yml -i $(INVENTORY)

# Role: Artifactors
nexus:
	ansible-playbook roles/artifactors/tasks/nexus.yml -i $(INVENTORY)

# Role: Common
docker:
	ansible-playbook roles/common/tasks/docker-install.yml -i $(INVENTORY)

prometheus:
	ansible-playbook roles/common/tasks/prometheus-install.yml -i $(INVENTORY)

ssh:
	ansible-playbook roles/common/tasks/ssh-setup.yml -i $(INVENTORY)
