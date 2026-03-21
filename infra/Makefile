TF_DIR := terraform
PEM     := pill-doser.pem

.PHONY: up down redeploy ssh

up:
	cd $(TF_DIR) && terraform init && terraform apply

down:
	cd $(TF_DIR) && terraform destroy

redeploy:
	cd $(TF_DIR) && terraform taint aws_instance.pill_doser && terraform apply

ssh:
	ssh -i $(PEM) -o StrictHostKeyChecking=no ec2-user@$$(cd $(TF_DIR) && terraform output -raw elastic_ip)
