package main

type ServerConfig struct {
	Port int
}

type DatabaseConfig struct {
	Host     string
	User     string
	Password string
	DBName   string
	Port     int
}
