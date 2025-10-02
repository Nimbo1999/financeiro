PHONY: compose-up compose-down compose-logs psql-auth psql-user psql-notification

compose-up:
	docker-compose up -d

compose-down:
	docker-compose down -v

compose-logs:
	docker-compose logs -f --tail=100

psql-auth:
	docker exec -it authentication_postgres_db psql -U matheus -d financeiro_authentication

psql-user:
	docker exec -it user_postgres_db psql -U matheus -d financeiro_user

psql-notification:
	docker exec -it notification_postgres_db psql -U matheus -d financeiro_notification
