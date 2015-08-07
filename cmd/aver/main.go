package main

import (
	"log"
	"os"

	"github.com/ivotron/aver"
	"github.com/spf13/cobra"
)

var dbConfig string
var tblName string

func main() {

	var cmd = &cobra.Command{
		Use:   "aver [OPTIONS] \"<statement(s)>\"",
		Short: "Aver helps you automatically validate assertions on data",
		Long:  ``,
		Run:   Execute,
	}

	cmd.Flags().StringVarP(&dbConfig, "dbconf", "c", "db.json", `Name of file
			containing database configuration. Format is JSON where only top-level
			elements are considered. See http://github.com/ivotron/aver for supported
			drivers and configuration examples.`)
	cmd.Flags().StringVarP(&tblName, "table", "t", "data", `Name of the table that
			will be queried when validating statements`)

	cmd.Execute()
}

func Execute(cmd *cobra.Command, args []string) {
	if len(args) > 1 {
		log.Fatal("ERROR: Expecting one double-quoted string as argument.")
	}

	db, err := aver.MakeDb(dbConfig)
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
