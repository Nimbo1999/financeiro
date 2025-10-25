PHONY: compose-up compose-down compose-down-volumes compose-up-build compose-logs psql-auth psql-user psql-notification pods svc

compose-up: compose-down
	docker-compose up -d

compose-down:
	docker-compose down

compose-down-volumes:
	docker-compose down -v

compose-up-build: compose-down
	docker-compose up -d --build

compose-logs:
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
