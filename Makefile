PHONY: up down down-volumes up-build logs psql-auth psql-user psql-notification pods svc

up: down
	docker-compose up -d

down:
	docker-compose down

down-volumes:
	docker-compose down -v

up-build: down
	docker-compose up -d --build

logs:
	docker-compose logs -f --tail=100

psql-auth:
	docker exec -it authentication_postgres_db psql -U matheus -d financeiro_authentication

psql-user:
	docker exec -it user_postgres_db psql -U matheus -d financeiro_user

psql-transaction:
	docker exec -it transaction_postgres_db psql -U matheus -d financeiro_transaction

psql-notification:
	docker exec -it notification_postgres_db psql -U matheus -d financeiro_notification

pods:
	kubectl get pods -n financeiro

svc:
	kubectl get svc -n financeiro

# Logs from production deployments

logs-auth:
	kubectl logs -n financeiro -f --tail=5 Deployments/authentication

logs-users:
	kubectl logs -n financeiro -f --tail=5 Deployments/users

logs-notification:
	kubectl logs -n financeiro -f --tail=5 Deployments/notification

logs-frontend:
	kubectl logs -n financeiro -f --tail=5 Deployments/frontend
