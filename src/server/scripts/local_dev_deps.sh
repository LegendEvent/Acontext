set -e
rm -rf ./db/dev_data
rm -rf ./db/dev_redis
DATABASE_LOCATION="./db/dev_data" REDIS_LOCATION="./db/dev_redis" docker compose up acontext-server-pg acontext-server-redis
rm -rf ./db/dev_data
rm -rf ./db/dev_redis