# Comparador de Bancos de Dados PostgreSQL

Este aplicativo em Go compara dois bancos de dados PostgreSQL e gera um relatório detalhado das diferenças entre suas estruturas, incluindo tabelas ausentes, colunas ausentes, diferenças de tipos de dados e configurações de nullable.

## Requisitos

- Go 1.16 ou superior
- Acesso aos bancos de dados PostgreSQL que deseja comparar
- Driver PostgreSQL para Go (`github.com/lib/pq`)

## Instalação

1. Clone ou baixe este repositório
2. Instale as dependências necessárias:

```bash
go mod init dbcompare
go get github.com/lib/pq
```

## Uso

Você pode executar o aplicativo de duas maneiras:

### 1. Usando argumentos de linha de comando

Execute o aplicativo fornecendo as strings de conexão e nomes para os dois bancos de dados que deseja comparar:

```bash
go run main.go "host=localhost port=5432 user=postgres password=senha dbname=db1 sslmode=disable" "DB1" "host=localhost port=5432 user=postgres password=senha dbname=db2 sslmode=disable" "DB2"
```

Substitua os parâmetros de conexão pelos valores apropriados para seus bancos de dados.

### 2. Usando arquivo de configuração JSON

Alternativamente, você pode usar um arquivo de configuração JSON para especificar os parâmetros de conexão e opções de comparação:

```bash
go run main.go -c config.json
```

O aplicativo também tentará carregar automaticamente um arquivo `config.json` no diretório atual se nenhum argumento for fornecido.

## Arquivo de Configuração

O arquivo de configuração JSON deve seguir o seguinte formato:

```json
{
  "database1": {
    "connection": "host=localhost port=5432 user=postgres password=senha dbname=db1 sslmode=disable",
    "name": "DB1"
  },
  "database2": {
    "connection": "host=localhost port=5432 user=postgres password=senha dbname=db2 sslmode=disable",
    "name": "DB2"
  },
  "options": {
    "compare_tables": true,
    "compare_columns": true,
    "compare_types": true,
    "compare_nullable": true
  }
}
```

As opções permitem controlar quais tipos de diferenças serão incluídas no relatório.

## Saída

O aplicativo gerará um relatório detalhado no console, mostrando:

1. Tabelas que existem em um banco de dados mas não no outro
2. Colunas que existem em um banco de dados mas não no outro
3. Colunas com tipos de dados diferentes entre os bancos
4. Colunas com configurações de nullable diferentes
5. Um resumo com a contagem de cada tipo de diferença

## Exemplo de Saída

```
Relatório de Diferenças entre DB1 e DB2

1. TABELAS AUSENTES (2):
Tabela                 DB1        DB2       
--------------------   ----------  ----------
clientes               presente    ausente   
usuarios               ausente     presente  

2. COLUNAS AUSENTES (3):
Tabela                 Coluna                 DB1        DB2       
--------------------   --------------------   ----------  ----------
produtos               descricao              presente    ausente   
produtos               categoria              ausente     presente  
pedidos                data_entrega           ausente     presente  

3. DIFERENÇAS DE TIPO DE COLUNA (2):
Tabela                 Coluna                 DB1        DB2       
--------------------   --------------------   ----------  ----------
produtos               preco                  numeric     real      
pedidos                id_cliente            integer     bigint    

4. DIFERENÇAS DE NULLABLE (1):
Tabela                 Coluna                 DB1        DB2       
--------------------   --------------------   ----------  ----------
produtos               estoque                NOT NULL    NULL      

RESUMO:
- Total de diferenças: 8
- Tabelas ausentes: 2
- Colunas ausentes: 3
- Diferenças de tipo de coluna: 2
- Diferenças de nullable: 1
```

## Limitações

- O aplicativo compara apenas tabelas no esquema 'public'
- Não compara índices, chaves estrangeiras, restrições ou outros objetos do banco de dados
- Não compara dados, apenas a estrutura do banco de dados