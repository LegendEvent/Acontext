set -e
rm -rf ./db/dev_data
rm -rf ./db/dev_redis
rm -rf ./db/dev_rabbitmq
DATABASE_LOCATION="./db/dev_data" REDIS_LOCATION="./db/dev_redis" RABBITMQ_LOCATION="./db/dev_rabbitmq" docker compose up acontext-server-pg acontext-server-redis acontext-server-rabbitmq
rm -rf ./db/dev_data
rm -rf ./db/dev_redis
rm -rf ./db/dev_rabbitmq