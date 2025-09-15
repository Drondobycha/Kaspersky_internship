package main

import (
	"Kaspersky_internship/internal/config"
	"Kaspersky_internship/internal/server"
	"Kaspersky_internship/internal/worker"
	"flag"
	"log"
	"os"
)

func main() {
	//Парсим флаги командной строки
	var filename string
	flag.StringVar(&filename, "filename", "config.env", "File name")
	//Подгружаем конфигурацию
	if err := config.LoadEnv(filename); err != nil {
		panic("Failed to set environment variables")
	}
	cfg := config.Load()

	// Создаем экземпляр сервера с TaskHendler
	srv := server.NewServer(cfg, worker.TaskHandler)

	// Запускаем сервер
	if err := srv.Start(); err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}

	log.Printf("Application exited")
}
