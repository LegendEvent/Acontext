set -e
rm -rf ./db/dev_data
rm -rf ./db/dev_redis
rm -rf ./db/dev_clickhouse
DATABASE_LOCATION="./db/dev_data" REDIS_LOCATION="./db/dev_redis" CLICKHOUSE_LOCATION="./db/dev_clickhouse" docker compose up acontext-server-pg acontext-server-redis acontext-server-clickhouse
rm -rf ./db/dev_data
rm -rf ./db/dev_redis
rm -rf ./db/dev_clickhouse