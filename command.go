package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Mitu217/tamate/util"

	"github.com/Mitu217/tamate/differ"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/Mitu217/tamate/datasource"
	"github.com/Mitu217/tamate/schema"

	"github.com/urfave/cli"
	"io"
	"text/tabwriter"
)

func main() {
	app := cli.NewApp()
	app.Name = "tamate"
	app.Usage = "tamate commands"
	app.Version = "0.1.0"

	// Commands
	app.Commands = []cli.Command{
		{
			Name:   "generate:config",
			Usage:  "Generate config file.",
			Action: generateConfigAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "type, t",
					Usage: "",
				},
				cli.StringFlag{
					Name:  "output, o",
					Usage: "",
				},
			},
		},
		{
			Name:   "generate:schema",
			Usage:  "Generate schema file.",
			Action: generateSchemaAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "type, t",
					Usage: "",
				},
				cli.StringFlag{
					Name:  "config, c",
					Usage: "",
				},
				cli.StringFlag{
					Name:  "output, o",
					Usage: "",
				},
			},
		},
		{
			Name:   "dump",
			Usage:  "Dump Command.",
			Action: dumpAction,
		},
		{
			Name:   "diff",
			Usage:  "Diff between 2 schema.",
			Action: diffAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "schema, s",
					Usage: "schema file path.",
				},
			},
		},
	}

	// Run Commands
	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}

func generateConfigAction(c *cli.Context) {
	// Override output path
	var w io.Writer
	if c.String("output") != "" {
		f, err := os.OpenFile(c.String("output"), os.O_CREATE, 0644)
		defer f.Close()
		if err != nil {
			log.Fatalln(err)
		}
		w = f
	} else {
		w = os.Stdout
	}

	if err := generateConfig(w, c.String("type")); err != nil {
		log.Fatalln(err)
	}
}

func generateSchemaAction(c *cli.Context) {
	// Override output path
	var w io.Writer
	inputType := strings.ToLower(c.String("type"))
	configPath := ""
	if c.String("config") == "" {
		log.Fatalln("must specify -c (--config) option")
	}
	configPath = c.String("config")
	if c.String("output") != "" {
		f, err := os.OpenFile(c.String("output"), os.O_CREATE, 0644)
		defer f.Close()
		if err != nil {
			log.Fatalln(err)
		}
		w = f
	} else {
		w = os.Stdout
	}

	switch inputType {
	case "sql":
		conf := &datasource.SQLDatasourceConfig{}
		if err := datasource.NewConfigFromJSONFile(configPath, conf); err != nil {
			log.Fatalf("Unable to read config file: %s\n%v", configPath, err)
		}
		ds, err := datasource.NewSQLDataSource(conf)
		if err != nil {
			log.Fatalln(err)
		}
		sc, err := ds.GetSchema()
		if err != nil {
			log.Fatalln(err)
		}
		if err := sc.ToJSON(w); err != nil {
			log.Fatalln(err)
		}
		break
	default:
		log.Fatalln("Not defined input type. type:" + inputType)
	}
}

func dumpAction(c *cli.Context) {
	// Get InputDataSource
	if len(c.Args()) < 1 {
		log.Fatalf("Please specify the input type and datasource config path.")
	}
	inputConfigPath := c.Args()[0]
	inputDatasource, err := util.GetConfigDataSource(inputConfigPath)
	if err != nil {
		log.Fatalln(err)
	}

	// Standard Output
	rows, err := inputDatasource.GetRows()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("[Input DataSource]")
	printRows(rows)
}

func diffAction(c *cli.Context) {
	schemaFilePath := c.String("schema")

	// read schema
	var diffSchema *schema.Schema
	if schemaFilePath != "" {
		sc, err := schema.NewJSONFileSchema(schemaFilePath)
		if err != nil {
			log.Fatalf("Unable to read schema file: %v", err)
		}
		diffSchema = sc
	}

	// left datasource
	if len(c.Args()) < 1 {
		log.Fatalf("Please specify the left type and datasource config path.")
	}
	leftConfigPath := c.Args()[0]
	leftDatasource, err := util.GetConfigDataSource(leftConfigPath)
	if err != nil {
		log.Fatalln(err)
	}
	if err := leftDatasource.SetSchema(diffSchema); err != nil {
		log.Fatalln(err)
	}

	// right datasource
	if len(c.Args()) < 2 {
		log.Fatalf("Please specify the right type and datasource config path.")
	}
	rightConfigPath := c.Args()[1]
	rightDatasource, err := util.GetConfigDataSource(rightConfigPath)
	if err != nil {
		log.Fatalln(err)
	}
	if err := rightDatasource.SetSchema(diffSchema); err != nil {
		log.Fatalln(err)
	}

	var d *differ.Differ
	if diffSchema == nil {
		rowsDiffer, err := differ.NewRowsDiffer(leftDatasource, rightDatasource)
		if err != nil {
			log.Fatalln(err)
		}
		d = rowsDiffer
	} else {
		schemaDiffer, err := differ.NewSchemaDiffer(diffSchema, leftDatasource, rightDatasource)
		if err != nil {
			log.Fatalln(err)
		}
		d = schemaDiffer
	}
	if err != nil {
		log.Fatalln(err)
	}
	diff, err := d.DiffRows()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("[Add]")
	printRows(diff.Add)
	fmt.Println("[Delete]")
	printRows(diff.Delete)
	fmt.Println("[Modify]")
	printRows(diff.Modify)
}

func generateConfig(w io.Writer, configType string) error {
	t := strings.ToLower(configType)
	isStdinTerm := terminal.IsTerminal(0) // fd0: stdin
	switch t {
	case "sql":
		conf := &datasource.SQLDatasourceConfig{Type: t}
		if isStdinTerm {
			fmt.Print("DriverName: ")
			fmt.Scan(&conf.DriverName)
		}
		if isStdinTerm {
			fmt.Print("DSN: ")
			fmt.Scan(&conf.DSN)
		}
		if isStdinTerm {
			fmt.Print("DatabaseName: ")
			fmt.Scan(&conf.DatabaseName)
		}
		if isStdinTerm {
			fmt.Print("TableName: ")
			fmt.Scan(&conf.TableName)
		}

		if err := datasource.ConfigToJSON(w, conf); err != nil {
			return err
		}
		return nil
	case "spreadsheets":
		conf := &datasource.SpreadSheetsDatasourceConfig{Type: t}
		if isStdinTerm {
			fmt.Print("SpreadSheetsID: ")
			fmt.Scan(&conf.SpreadSheetsID)
		}
		// TODO: スペース入りの文字列が対応不可
		if isStdinTerm {
			fmt.Print("SheetName: ")
			fmt.Scan(&conf.SheetName)
		}
		if isStdinTerm {
			fmt.Print("Range: ")
			fmt.Scan(&conf.Range)
		}
		if err := datasource.ConfigToJSON(w, conf); err != nil {
			return err
		}
		return nil
	case "csv":
		conf := &datasource.CSVDatasourceConfig{Type: t}
		if isStdinTerm {
			fmt.Print("FilePath: ")
			fmt.Scan(&conf.Path)
		}
		if err := datasource.ConfigToJSON(w, conf); err != nil {
			return err
		}
		return nil
	default:
		return errors.New("Not defined input type. type:" + configType)
	}
}

func printRows(rows *datasource.Rows) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', tabwriter.TabIndent)
	// Columns
	w.Write([]byte(strings.Join(rows.Columns, "\t") + "\n"))
	// Values
	for i := range rows.Values {
		w.Write([]byte(strings.Join(rows.Values[i], "\t") + "\n"))
	}
	w.Flush()
}
