# Full Cycle Auction (Go Expert)

Sistema de leiloes em Go com fechamento automatico de leiloes por expiracao.

## Requisitos

- Go 1.20+
- Docker e Docker Compose

## Variaveis de ambiente

As variaveis ficam em `cmd/auction/.env`.

Principais variaveis:

- `MONGODB_URL`: string de conexao do MongoDB.
- `MONGODB_DB`: nome do banco.
- `AUCTION_DURATION`: duracao do leilao (exemplos: `20s`, `1m`, `5m`).
- `BATCH_INSERT_INTERVAL`: intervalo do processamento em lote de bids.
- `MAX_BATCH_SIZE`: tamanho maximo do lote de bids.

Observacao:

- Se `AUCTION_DURATION` nao for definida ou for invalida, o sistema usa `5m` como default.

## Rodando com Docker Compose

Suba toda a stack:

```bash
docker compose up --build
```

A API sobe na porta `8080`.

## Rodando localmente (sem Docker para app)

1. Suba apenas o MongoDB (via Compose):

```bash
docker compose up -d mongodb
```

2. Ajuste `cmd/auction/.env` para apontar para seu Mongo.
3. Execute a aplicacao:

```bash
go run cmd/auction/main.go
```

## Endpoint relevante do desafio

- `POST /auction`: cria um leilao e agenda o fechamento automatico.

Quando o tempo de `AUCTION_DURATION` expira, o status do leilao e atualizado para `Closed` em background (goroutine).

## Testes manuais com HTTP

O arquivo [test/auction-tests.http](test/auction-tests.http) é util para validar rapidamente o fluxo completo da API (criar usuario, criar leilao, enviar lances e verificar expiracao).

Como usar:

1. Inicie a API (`docker compose up --build` ou `go run cmd/auction/main.go`).
2. Abra [test/auction-tests.http](test/auction-tests.http) no VS Code com a extensao REST Client.
3. Execute os requests na ordem do arquivo.
4. Substitua os placeholders `<USER_ID>` e `<AUCTION_ID>` apos criar os recursos.

Observacao sobre status do leilao:

- `status=0` -> `Active`
- `status=1` -> `Closed`

## Teste automatizado de expiracao

Foi adicionado teste automatizado para validar o fechamento automatico:

- Arquivo: `internal/infra/database/auction/create_auction_test.go`
- Cenario validado:
  1. Cria leilao
  2. Verifica status inicial `Active`
  3. Aguarda expiracao
  4. Verifica status final `Closed`

### Executar apenas o teste de expiracao

```bash
go test ./internal/infra/database/auction -run TestCreateAuction_ExpiresAutomatically -v
```

### Configuracao de conexao Mongo para teste

O teste tenta conectar usando, em ordem:

1. `AUCTION_TEST_MONGODB_URL`
2. `MONGODB_URL`
3. default: `mongodb://admin:admin@localhost:27017/?authSource=admin`

Banco para teste (em ordem):

1. `AUCTION_TEST_MONGODB_DB`
2. `MONGODB_DB`
3. default: `auctions`

Se o Mongo nao estiver acessivel, o teste e marcado como skipped.

## Testes gerais

```bash
go test ./...
```
