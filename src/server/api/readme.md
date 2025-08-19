acontext-api/
├─ cmd/
│  └─ server/
│     └─ main.go
├─ internal/
│  ├─ di/
│  │  └─ container.go          # samber/do
│  ├─ config/
│  │  └─ config.go             # Viper
│  ├─ logger/
│  │  └─ logger.go             # Zap
│  ├─ db/
│  │  └─ db.go                 # GORM
│  ├─ model/                   # database schema
│  ├─ repo/                    # database access layer
│  ├─ service/                 # business logic layer
│  ├─ handler/                 # handler
│  └─ router/                  # router
├─ configs/
│  └─ config.yaml              # config file
├─ docs/                       # swag docs
├─ go.mod
├─ Makefile
└─ README.md
