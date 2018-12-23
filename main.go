package main

import (
	"io/ioutil"
	"log"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

type column struct {
	TableName string `gorm:"column:TABLE_NAME"`
	Name      string `gorm:"column:COLUMN_NAME"`
	Type      string `gorm:"column:COLUMN_TYPE"`
	Null      string `gorm:"column:IS_NULLABLE"`
	Key       string `gorm:"column:COLUMN_KEY"`
	Default   string `gorm:"column:COLUMN_DEFAULT"`
	Extra     string `gorm:"column:EXTRA"`
	DataType  string `gorm:"column:DATA_TYPE"`
}

func main() {
	db = openDatabaseConn()
	defer db.Close()

	// router := mux.NewRouter()
	// http.ListenAndServe(":8080", router)

	tables2structs("2jours_calculatie")
}

func openDatabaseConn() (db *gorm.DB) {

	driver := "mysql"
	user := "devuser"
	pass := "5;9[,Mbg49yP6s!Q"
	name := "information_schema"

	db, err := gorm.Open(driver, user+":"+pass+"@/"+name+"?charset=utf8&parseTime=True")

	if err != nil {
		panic(err.Error())
	}

	db.LogMode(true)
	return db
}

func tables2structs(dbName string) {

	columns := []column{}

	// haal de informatie van de tabellen en kolommen op
	query := db.Where("TABLE_SCHEMA = ?", dbName).Find(&columns)
	if query.Error != nil {
		log.Fatal(query.Error)
	}

	// maak een array aan met als key de naam van de tabel
	// zo krijg je een lijst met tabellen die weer een array bevatten met kolommen
	tables := make(map[string][]column)
	for _, col := range columns {
		tables[col.TableName] = append(tables[col.TableName], col)
	}

	// maak de model files aan
	for name, columns := range tables {

		// if name != "admicom_groep" {
		// continue
		// }

		containsDateTime := containsDateTime(columns)

		content := "package models\r\n"
		content += "\r\n"

		if containsDateTime {
			content += "import \"time\"\r\n"
			content += "\r\n"
		}

		tableName := strings.Title(getFieldName(name))

		content += "// " + tableName + " model ...\r\n"
		content += "type " + tableName + " struct {"
		content += "\r\n"

		maxFieldNameSize := getMaxFieldNameSize(columns)
		maxDataTypeSize := getMaxDataTypeSize(columns)

		for _, col := range columns {

			tags := getTags(col)

			fieldName := getFieldName(col.Name)
			for i := len(fieldName); i <= maxFieldNameSize; i++ {
				fieldName += " "
			}

			columnType := getColumnType(col.DataType)
			for i := len(columnType); i <= maxDataTypeSize; i++ {
				columnType += " "
			}

			content += "\t" + fieldName + columnType + tags + "\r\n"
		}

		content += "}"

		bytes := []byte(content)
		err := ioutil.WriteFile("e:/temp/models/"+name+".txt", bytes, 0644)
		check(err)
	}
}

func getMaxFieldNameSize(columns []column) int {

	max := 0
	for _, col := range columns {
		size := len(col.Name)
		if size > max {
			max = size
		}
	}
	return max
}

func getMaxDataTypeSize(columns []column) int {

	max := 0
	for _, col := range columns {
		size := len(col.DataType)
		if size > max {
			max = size
		}
	}
	return max
}

func getTags(col column) string {

	var jsonTag = getFieldName(col.Name)
	jsonTag = strings.ToLower(jsonTag[0:1]) + jsonTag[1:len(jsonTag)]

	if col.Name == "Id" {
		jsonTag = "id"
	} else {
		// indien eindigt op id of Id dan ID van maken volgens GoLang
		if strings.HasSuffix(jsonTag, "ID") {
			runes := []rune(jsonTag)
			jsonTag = string(runes[0:len(runes)-2]) + "Id"
		}
	}

	var gormTag = ""

	if col.Key == "PRI" {
		gormTag += "primary_key;"
	}

	gormTag += "column:" + col.Name

	return " `gorm:\"" + gormTag + "\" json:\"" + jsonTag + "\"`"
}

func getFieldName(columnName string) string {

	columnName = strings.ToLower(columnName)
	columnName = strings.Title(columnName)

	if strings.Contains(columnName, "_") {

		fieldName := ""
		arr := strings.Split(strings.ToLower(columnName), "_")

		for _, s := range arr {
			fieldName += strings.Title(s)
		}

		columnName = fieldName
	}

	// indien eindigt op id of Id dan ID van maken volgens GoLang
	if strings.HasSuffix(columnName, "id") || strings.HasSuffix(columnName, "Id") {
		runes := []rune(columnName)
		columnName = string(runes[0:len(runes)-2]) + "ID"
	}

	return columnName
}

func getColumnType(columnType string) string {

	columnType = strings.ToLower(columnType)

	if strings.Contains(columnType, "int") {
		return "int"
	}

	if strings.Contains(columnType, "datetime") {
		return "time.Time"
	}

	if strings.Contains(columnType, "varchar") {
		return "string"
	}

	if strings.Contains(columnType, "text") {
		return "string"
	}

	if strings.Contains(columnType, "decimal") || strings.Contains(columnType, "double") {
		return "float64"
	}

	return columnType
}

func containsDateTime(columns []column) bool {

	for _, col := range columns {
		columnType := strings.ToLower(col.Type)
		if strings.Contains(columnType, "date") || strings.Contains(columnType, "time") {
			return true
		}
	}

	return false
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
