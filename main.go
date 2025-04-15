package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	_ "github.com/lib/pq"
)

// Estrutura para armazenar informações da tabela
type Table struct {
	Name    string
	Columns []Column
}

// Estrutura para armazenar informações da coluna
type Column struct {
	Name     string
	DataType string
	Nullable bool
}

// Estrutura para armazenar diferenças entre bancos de dados
type Difference struct {
	Type        string // "table_missing", "column_missing", "column_type_different", "column_nullable_different"
	TableName   string
	ColumnName  string
	DB1Value    string
	DB2Value    string
	Description string
}

// Estrutura para armazenar configurações do arquivo JSON
type Config struct {
	Database1 DatabaseConfig `json:"database1"`
	Database2 DatabaseConfig `json:"database2"`
	Options   OptionsConfig  `json:"options"`
}

// Estrutura para armazenar configurações de banco de dados
type DatabaseConfig struct {
	Connection string `json:"connection"`
	Name       string `json:"name"`
}

// Estrutura para armazenar opções de comparação
type OptionsConfig struct {
	CompareTables   bool `json:"compare_tables"`
	CompareColumns  bool `json:"compare_columns"`
	CompareTypes    bool `json:"compare_types"`
	CompareNullable bool `json:"compare_nullable"`
}

// Função para carregar configurações do arquivo JSON
func loadConfig(filePath string) (Config, error) {
	var config Config

	// Ler o arquivo de configuração
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return config, fmt.Errorf("erro ao ler arquivo de configuração: %v", err)
	}

	// Decodificar o JSON
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("erro ao decodificar JSON: %v", err)
	}

	return config, nil
}

func main() {
	var db1ConnStr, db1Name, db2ConnStr, db2Name string
	var config Config
	var err error

	// Verificar se foi fornecido um arquivo de configuração
	if len(os.Args) == 2 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Println("Uso: go run main.go -c <arquivo_config>")
		fmt.Println("   ou: go run main.go <db1_connection_string> <db1_name> <db2_connection_string> <db2_name>")
		fmt.Println("Exemplo: go run main.go -c config.json")
		os.Exit(1)
	} else if len(os.Args) == 3 && (os.Args[1] == "-c" || os.Args[1] == "--config") {
		// Carregar configurações do arquivo JSON
		configFile := os.Args[2]
		config, err = loadConfig(configFile)
		if err != nil {
			log.Fatalf("Erro ao carregar configurações: %v", err)
		}

		// Usar configurações do arquivo
		db1ConnStr = config.Database1.Connection
		db1Name = config.Database1.Name
		db2ConnStr = config.Database2.Connection
		db2Name = config.Database2.Name

		fmt.Printf("Configurações carregadas do arquivo: %s\n", configFile)
	} else if len(os.Args) < 5 {
		// Tentar carregar configurações do arquivo padrão
		config, err = loadConfig("config.json")
		if err != nil {
			fmt.Println("Uso: go run main.go <db1_connection_string> <db1_name> <db2_connection_string> <db2_name>")
			fmt.Println("   ou: go run main.go -c <arquivo_config>")
			fmt.Println("Exemplo: go run main.go 'host=localhost port=5432 user=postgres password=senha dbname=db1 sslmode=disable' 'DB1' 'host=localhost port=5432 user=postgres password=senha dbname=db2 sslmode=disable' 'DB2'")
			fmt.Println("   ou: go run main.go -c config.json")
			os.Exit(1)
		}

		// Usar configurações do arquivo padrão
		db1ConnStr = config.Database1.Connection
		db1Name = config.Database1.Name
		db2ConnStr = config.Database2.Connection
		db2Name = config.Database2.Name

		fmt.Printf("Configurações carregadas do arquivo padrão: config.json\n")
	} else {
		// Usar argumentos da linha de comando
		db1ConnStr = os.Args[1]
		db1Name = os.Args[2]
		db2ConnStr = os.Args[3]
		db2Name = os.Args[4]
	}

	// Conectar ao primeiro banco de dados
	db1, err := sql.Open("postgres", db1ConnStr)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados 1: %v", err)
	}
	defer db1.Close()

	// Verificar conexão
	err = db1.Ping()
	if err != nil {
		log.Fatalf("Erro ao verificar conexão com banco de dados 1: %v", err)
	}

	// Conectar ao segundo banco de dados
	db2, err := sql.Open("postgres", db2ConnStr)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados 2: %v", err)
	}
	defer db2.Close()

	// Verificar conexão
	err = db2.Ping()
	if err != nil {
		log.Fatalf("Erro ao verificar conexão com banco de dados 2: %v", err)
	}

	fmt.Printf("Conectado com sucesso aos bancos de dados %s e %s\n\n", db1Name, db2Name)

	// Obter tabelas e colunas dos bancos de dados
	db1Tables, err := getTables(db1)
	if err != nil {
		log.Fatalf("Erro ao obter tabelas do banco de dados 1: %v", err)
	}

	db2Tables, err := getTables(db2)
	if err != nil {
		log.Fatalf("Erro ao obter tabelas do banco de dados 2: %v", err)
	}

	// Comparar os bancos de dados
	differences := compareDatabases(db1Tables, db2Tables, db1Name, db2Name)

	// Aplicar filtros de opções se estiver usando arquivo de configuração
	if len(os.Args) >= 2 && (os.Args[1] == "-c" || os.Args[1] == "--config") {
		differences = applyConfigFilters(differences, config.Options)
	}

	// Gerar relatório
	generateReport(differences, db1Name, db2Name)
}

// Função para obter todas as tabelas e colunas de um banco de dados
func getTables(db *sql.DB) ([]Table, error) {
	// Consulta para obter todas as tabelas do esquema public
	rows, err := db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}

		// Obter colunas para esta tabela
		columns, err := getColumns(db, tableName)
		if err != nil {
			return nil, err
		}

		tables = append(tables, Table{
			Name:    tableName,
			Columns: columns,
		})
	}

	return tables, nil
}

// Função para obter todas as colunas de uma tabela
func getColumns(db *sql.DB, tableName string) ([]Column, error) {
	// Consulta para obter todas as colunas de uma tabela
	rows, err := db.Query(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_schema = 'public' AND table_name = $1 
		ORDER BY ordinal_position
	`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var columnName, dataType, isNullable string
		if err := rows.Scan(&columnName, &dataType, &isNullable); err != nil {
			return nil, err
		}

		columns = append(columns, Column{
			Name:     columnName,
			DataType: dataType,
			Nullable: isNullable == "YES",
		})
	}

	return columns, nil
}

// Função para comparar dois bancos de dados
// Função para aplicar filtros de opções às diferenças encontradas
func applyConfigFilters(differences []Difference, options OptionsConfig) []Difference {
	var filteredDifferences []Difference

	for _, diff := range differences {
		switch diff.Type {
		case "table_missing":
			if options.CompareTables {
				filteredDifferences = append(filteredDifferences, diff)
			}
		case "column_missing":
			if options.CompareColumns {
				filteredDifferences = append(filteredDifferences, diff)
			}
		case "column_type_different":
			if options.CompareTypes {
				filteredDifferences = append(filteredDifferences, diff)
			}
		case "column_nullable_different":
			if options.CompareNullable {
				filteredDifferences = append(filteredDifferences, diff)
			}
		}
	}

	return filteredDifferences
}

func compareDatabases(db1Tables, db2Tables []Table, db1Name, db2Name string) []Difference {
	var differences []Difference

	// Mapear tabelas do DB1 por nome para facilitar a busca
	db1TablesMap := make(map[string]Table)
	for _, table := range db1Tables {
		db1TablesMap[table.Name] = table
	}

	// Mapear tabelas do DB2 por nome para facilitar a busca
	db2TablesMap := make(map[string]Table)
	for _, table := range db2Tables {
		db2TablesMap[table.Name] = table
	}

	// Verificar tabelas que existem em DB1 mas não em DB2
	for _, table := range db1Tables {
		if _, exists := db2TablesMap[table.Name]; !exists {
			differences = append(differences, Difference{
				Type:        "table_missing",
				TableName:   table.Name,
				DB1Value:    "presente",
				DB2Value:    "ausente",
				Description: fmt.Sprintf("Tabela '%s' existe em %s mas não em %s", table.Name, db1Name, db2Name),
			})
		}
	}

	// Verificar tabelas que existem em DB2 mas não em DB1
	for _, table := range db2Tables {
		if _, exists := db1TablesMap[table.Name]; !exists {
			differences = append(differences, Difference{
				Type:        "table_missing",
				TableName:   table.Name,
				DB1Value:    "ausente",
				DB2Value:    "presente",
				Description: fmt.Sprintf("Tabela '%s' existe em %s mas não em %s", table.Name, db2Name, db1Name),
			})
		}
	}

	// Comparar colunas das tabelas que existem em ambos os bancos
	for _, db1Table := range db1Tables {
		db2Table, exists := db2TablesMap[db1Table.Name]
		if !exists {
			continue // Tabela não existe no DB2, já foi registrada acima
		}

		// Mapear colunas do DB1 por nome para facilitar a busca
		db1ColumnsMap := make(map[string]Column)
		for _, column := range db1Table.Columns {
			db1ColumnsMap[column.Name] = column
		}

		// Mapear colunas do DB2 por nome para facilitar a busca
		db2ColumnsMap := make(map[string]Column)
		for _, column := range db2Table.Columns {
			db2ColumnsMap[column.Name] = column
		}

		// Verificar colunas que existem em DB1 mas não em DB2
		for _, column := range db1Table.Columns {
			if _, exists := db2ColumnsMap[column.Name]; !exists {
				differences = append(differences, Difference{
					Type:        "column_missing",
					TableName:   db1Table.Name,
					ColumnName:  column.Name,
					DB1Value:    "presente",
					DB2Value:    "ausente",
					Description: fmt.Sprintf("Coluna '%s.%s' existe em %s mas não em %s", db1Table.Name, column.Name, db1Name, db2Name),
				})
			}
		}

		// Verificar colunas que existem em DB2 mas não em DB1
		for _, column := range db2Table.Columns {
			if _, exists := db1ColumnsMap[column.Name]; !exists {
				differences = append(differences, Difference{
					Type:        "column_missing",
					TableName:   db2Table.Name,
					ColumnName:  column.Name,
					DB1Value:    "ausente",
					DB2Value:    "presente",
					Description: fmt.Sprintf("Coluna '%s.%s' existe em %s mas não em %s", db2Table.Name, column.Name, db2Name, db1Name),
				})
			}
		}

		// Comparar colunas que existem em ambos os bancos
		for _, db1Column := range db1Table.Columns {
			db2Column, exists := db2ColumnsMap[db1Column.Name]
			if !exists {
				continue // Coluna não existe no DB2, já foi registrada acima
			}

			// Verificar se o tipo de dados é diferente
			if db1Column.DataType != db2Column.DataType {
				differences = append(differences, Difference{
					Type:       "column_type_different",
					TableName:  db1Table.Name,
					ColumnName: db1Column.Name,
					DB1Value:   db1Column.DataType,
					DB2Value:   db2Column.DataType,
					Description: fmt.Sprintf("Coluna '%s.%s' tem tipo '%s' em %s e '%s' em %s",
						db1Table.Name, db1Column.Name, db1Column.DataType, db1Name, db2Column.DataType, db2Name),
				})
			}

			// Verificar se a propriedade nullable é diferente
			if db1Column.Nullable != db2Column.Nullable {
				db1Nullable := "NOT NULL"
				if db1Column.Nullable {
					db1Nullable = "NULL"
				}

				db2Nullable := "NOT NULL"
				if db2Column.Nullable {
					db2Nullable = "NULL"
				}

				differences = append(differences, Difference{
					Type:       "column_nullable_different",
					TableName:  db1Table.Name,
					ColumnName: db1Column.Name,
					DB1Value:   db1Nullable,
					DB2Value:   db2Nullable,
					Description: fmt.Sprintf("Coluna '%s.%s' é %s em %s e %s em %s",
						db1Table.Name, db1Column.Name, db1Nullable, db1Name, db2Nullable, db2Name),
				})
			}
		}
	}

	return differences
}

// Função para gerar o relatório de diferenças
func generateReport(differences []Difference, db1Name, db2Name string) {
	if len(differences) == 0 {
		fmt.Println("Nenhuma diferença encontrada entre os bancos de dados.")
		return
	}

	fmt.Printf("Relatório de Diferenças entre %s e %s\n\n", db1Name, db2Name)

	// Agrupar diferenças por tipo
	tableMissing := []Difference{}
	columnMissing := []Difference{}
	columnTypeDifferent := []Difference{}
	columnNullableDifferent := []Difference{}

	for _, diff := range differences {
		switch diff.Type {
		case "table_missing":
			tableMissing = append(tableMissing, diff)
		case "column_missing":
			columnMissing = append(columnMissing, diff)
		case "column_type_different":
			columnTypeDifferent = append(columnTypeDifferent, diff)
		case "column_nullable_different":
			columnNullableDifferent = append(columnNullableDifferent, diff)
		}
	}

	// Criar um writer formatado para saída tabular
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.TabIndent)

	// Imprimir tabelas ausentes
	if len(tableMissing) > 0 {
		fmt.Fprintf(w, "\n1. TABELAS AUSENTES (%d):\n", len(tableMissing))
		fmt.Fprintf(w, "Tabela\t%s\t%s\n", db1Name, db2Name)
		fmt.Fprintf(w, "%s\t%s\t%s\n", strings.Repeat("-", 20), strings.Repeat("-", 10), strings.Repeat("-", 10))
		for _, diff := range tableMissing {
			fmt.Fprintf(w, "%s\t%s\t%s\n", diff.TableName, diff.DB1Value, diff.DB2Value)
		}
		w.Flush()
	}

	// Imprimir colunas ausentes
	if len(columnMissing) > 0 {
		fmt.Fprintf(w, "\n2. COLUNAS AUSENTES (%d):\n", len(columnMissing))
		fmt.Fprintf(w, "Tabela\tColuna\t%s\t%s\n", db1Name, db2Name)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", strings.Repeat("-", 20), strings.Repeat("-", 20), strings.Repeat("-", 10), strings.Repeat("-", 10))
		for _, diff := range columnMissing {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", diff.TableName, diff.ColumnName, diff.DB1Value, diff.DB2Value)
		}
		w.Flush()
	}

	// Imprimir diferenças de tipo de coluna
	if len(columnTypeDifferent) > 0 {
		fmt.Fprintf(w, "\n3. DIFERENÇAS DE TIPO DE COLUNA (%d):\n", len(columnTypeDifferent))
		fmt.Fprintf(w, "Tabela\tColuna\t%s\t%s\n", db1Name, db2Name)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", strings.Repeat("-", 20), strings.Repeat("-", 20), strings.Repeat("-", 10), strings.Repeat("-", 10))
		for _, diff := range columnTypeDifferent {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", diff.TableName, diff.ColumnName, diff.DB1Value, diff.DB2Value)
		}
		w.Flush()
	}

	// Imprimir diferenças de nullable
	if len(columnNullableDifferent) > 0 {
		fmt.Fprintf(w, "\n4. DIFERENÇAS DE NULLABLE (%d):\n", len(columnNullableDifferent))
		fmt.Fprintf(w, "Tabela\tColuna\t%s\t%s\n", db1Name, db2Name)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", strings.Repeat("-", 20), strings.Repeat("-", 20), strings.Repeat("-", 10), strings.Repeat("-", 10))
		for _, diff := range columnNullableDifferent {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", diff.TableName, diff.ColumnName, diff.DB1Value, diff.DB2Value)
		}
		w.Flush()
	}

	// Resumo
	fmt.Printf("\nRESUMO:\n")
	fmt.Printf("- Total de diferenças: %d\n", len(differences))
	fmt.Printf("- Tabelas ausentes: %d\n", len(tableMissing))
	fmt.Printf("- Colunas ausentes: %d\n", len(columnMissing))
	fmt.Printf("- Diferenças de tipo de coluna: %d\n", len(columnTypeDifferent))
	fmt.Printf("- Diferenças de nullable: %d\n", len(columnNullableDifferent))
}
