package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/ivotron/aver"
	"github.com/spf13/cobra"
)

var dbConfig string
var dataFile string
var fileType string
var printVersion bool
var toStdout bool

func main() {

	var cmd = &cobra.Command{
		Use:   "aver \"<statement(s)>\"",
		Short: "Aver helps you automatically validate assertions on data",
		Long:  ``,
		Run:   Execute,
	}

	cmd.Flags().BoolVarP(&toStdout, "stdout", "s", false, `Print result (true|false)
			to stdout. By default, the program returns an exit code of 1 if the validation
			fails, and 0 if it holds. With this option, 1 is returned only if there's an error
			in the execution of the program; the result of the validation is printed
			to stdout ('false' if it fails; 'true' if it holds). Thus, when --stdout is given,
			the exit code will always be 0 (unless there's an error) regardless of wheter the
			validation holds or not.`)
	cmd.Flags().BoolVarP(&printVersion, "version", "v", false, `Print program version.`)
	cmd.Flags().StringVarP(&dbConfig, "dbconf", "c", "", `Name of file containing
			database configuration. Format is JSON where only top-level elements are
			considered. See http://github.com/ivotron/aver for supported drivers and
			configuration examples.`)
	cmd.Flags().StringVarP(&dataFile, "input", "i", "", `File to read data from.
			Format is inferred from file extension. 'csv' and 'json' supported; 'csv'
			is assumed for files without extension.`)

	cmd.Execute()
}

func Execute(cmd *cobra.Command, args []string) {
	if printVersion {
		fmt.Println("Aver validation engine v0.3.0 -- HEAD")
		os.Exit(0)
	}
	if len(args) == 0 {
		log.Fatalln(cmd.UsageString())
	}
	if dataFile != "" && dbConfig != "" {
		log.Fatalln("ERROR: Options 'dbconf' and 'input' cannot be used simultaneously.")
	}
	if len(args) != 1 {
		log.Fatalln("ERROR: Expecting one double-quoted string as argument.")
	}
	if strings.HasSuffix(dataFile, ".json") {
		fileType = "json"
		log.Fatalln("ERROR: JSON not supported yet.")
	} else {
		fileType = "csv"
	}

	if dbConfig != "" {
		dataFile = dbConfig
		fileType = "config"
	}

	db, tblName, err := aver.MakeDb(dataFile, fileType)
	if err != nil {
		log.Fatalln("ERROR: " + err.Error())
	}

	holds, err := aver.Holds(args[0], db, tblName)
	if err != nil {
		var stack [4096]byte
		runtime.Stack(stack[:], true)
		log.Printf("%q\n%s\n", err, stack[:])
		log.Fatalln("ERROR: " + err.Error())
	}

	db.Close()

	if toStdout {
		fmt.Printf("%t\n", holds)
	} else if !holds {
		os.Exit(1)
	}
}
