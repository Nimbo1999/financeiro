# On the VPS, basic load test

ssh -i ~/.ssh/your-private-key <USERNAME>@<VPS_IP>

# Install apache bench if needed

sudo apt-get update && sudo apt-get install -y apache2-utils

# Load test on the Gateway (100 requests, 10 concurrent)

ab -n 100 -c 10 http://localhost:30083/health

# Check pods during load

watch -n 1 'kubectl get pods -n financeiro'
