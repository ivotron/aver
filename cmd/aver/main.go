package main

import (
	"log"
	"os"
	"strings"

	"github.com/ivotron/aver"
	"github.com/spf13/cobra"
)

var dbConfig string
var dataFile string
var fileType string

func main() {

	var cmd = &cobra.Command{
		Use:   "aver [OPTIONS] \"<statement(s)>\"",
		Short: "Aver helps you automatically validate assertions on data",
		Long:  ``,
		Run:   Execute,
	}

	cmd.Flags().StringVarP(&dbConfig, "dbconf", "c", "", `Name of file
			containing database configuration. Format is JSON where only top-level
			elements are considered. See http://github.com/ivotron/aver for supported
			drivers and configuration examples.`)
	cmd.Flags().StringVarP(&dataFile, "input", "i", "", `File to read data from.
			Format is inferred from file extension ('csv' and 'json' supported). 'csv'
			is assumed for files without extension`)

	cmd.Execute()
}

func Execute(cmd *cobra.Command, args []string) {
	if dataFile != "" && dbConfig != "" {
		log.Fatal("ERROR: Only one two options 'dbconf' and 'input' allowed.")
	}
	if len(args) > 1 {
		log.Fatal("ERROR: Expecting one double-quoted string as argument.")
	}
	if strings.HasSuffix(dataFile, ".json") {
		fileType = "json"
		log.Fatal("ERROR: JSON not supported yet.")
	} else {
		fileType = "csv"
	}

	if dbConfig != "" {
		dataFile = dbConfig
		fileType = "config"
	}

	db, tblName, err := aver.MakeDb(dataFile, fileType)
	if err != nil {
		log.Fatal("ERROR: " + err.Error())
	}

	holds, err := aver.Holds(args[0], db, tblName)
	if err != nil {
		log.Fatal("ERROR: " + err.Error())
	}

	db.Close()

	if !holds {
		os.Exit(1)
	}
}
