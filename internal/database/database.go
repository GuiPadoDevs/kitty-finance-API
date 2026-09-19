package database

import (
	"context"
	"embed"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB guarda o pool de conexões do PostgreSQL
type DB struct {
	Pool *pgxpool.Pool
}

// ConnectDatabase cria e inicializa o pool de conexões pgxpool
func ConnectDatabase(dbURL string) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("falha ao analisar DATABASE_URL: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 2
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar no PostgreSQL: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("banco de dados inacessível: %w", err)
	}

	log.Println("🎀 Conectado com sucesso ao PostgreSQL!")
	return &DB{Pool: pool}, nil
}

// RunMigrations executa scripts de migração SQL embutidos
func (db *DB) RunMigrations(migrationFS embed.FS) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	entries, err := migrationFS.ReadDir(".")
	if err != nil {
		return fmt.Errorf("erro ao ler diretório de migrações: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "embed.go" {
			continue
		}

		sqlContent, err := migrationFS.ReadFile(entry.Name())
		if err != nil {
			return fmt.Errorf("erro ao ler arquivo %s: %w", entry.Name(), err)
		}

		log.Printf("🌸 Executando migração: %s...", entry.Name())
		_, err = db.Pool.Exec(ctx, string(sqlContent))
		if err != nil {
			return fmt.Errorf("falha ao executar migração %s: %w", entry.Name(), err)
		}
	}

	log.Println("✨ Todas as migrações foram aplicadas com sucesso!")
	return nil
}

// Close fecha o pool de conexões
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
