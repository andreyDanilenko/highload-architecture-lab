module worker-pool

go 1.27

toolchain go1.27.1

require (
	github.com/joho/godotenv v1.5.1
	labshared v0.0.0
)

replace labshared => ../../shared/go
