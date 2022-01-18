package cmd

import (
	"strconv"
	"strings"

	"github.com/civo/civogo"
	"github.com/civo/cli/config"
	"github.com/civo/cli/utility"

	"github.com/spf13/cobra"
)

var filterSize string

var sizeListCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list", "all"},
	Example: `civo size ls`,
	Short:   "List sizes",
	Long: `List all available sizes for instances or Kubernetes nodes.
If you wish to use a custom format, the available fields are:

	* name
	* description
	* type
	* cpu_cores
	* ram_mb
	* disk_gb
	* selectable

Example: civo size ls -o custom -f "Code: name (type)"`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := config.CivoAPIClient()
		if regionSet != "" {
			client.Region = regionSet
		}
		if err != nil {
			utility.Error("Creating the connection to Civo's API failed with %s", err)
			return
		}

		filter := []civogo.InstanceSize{}
		sizes, err := client.ListInstanceSizes()
		if err != nil {
			utility.Error("%s", err)
			return
		}

		if filterSize != "" {
			search := ""

			switch {
			case filterSize == "database" || filterSize == "Database":
				search = ".db."
			case filterSize == "kubernetes" || filterSize == "Kubernetes":
				search = ".kube,"
			case filterSize == "instance" || filterSize == "Instance":
				search = "iaas"
			}

			for _, size := range sizes {
				if search == "iaas" {
					if !strings.Contains(size.Name, ".db.") && !strings.Contains(size.Name, ".kube.") {
						filter = append(filter, size)
					}
				} else {
					if strings.Contains(size.Name, search) {
						filter = append(filter, size)
					}
				}
			}

			sizes = filter
		}

		ow := utility.NewOutputWriter()

		for _, size := range sizes {

			ow.StartLine()
			ow.AppendDataWithLabel("name", size.Name, "Name")
			ow.AppendDataWithLabel("description", size.Description, "Description")

			switch {
			case strings.Contains(size.Name, ".db."):
				ow.AppendDataWithLabel("Type", "Database", "Type")
			case strings.Contains(size.Name, ".k3s.") || strings.Contains(size.Name, ".kube."):
				ow.AppendDataWithLabel("Type", "Kubernetes", "Type")
			default:
				ow.AppendDataWithLabel("type", "Instance", "Type")
			}
			ow.AppendDataWithLabel("cpu_cores", strconv.Itoa(size.CPUCores), "CPU")
			ow.AppendDataWithLabel("ram_mb", strconv.Itoa(size.RAMMegabytes), "RAM")
			ow.AppendDataWithLabel("disk_gb", strconv.Itoa(size.DiskGigabytes), "SSD")
			ow.AppendDataWithLabel("selectable", utility.BoolToYesNo(size.Selectable), "Selectable")
		}

		switch outputFormat {
		case "json":
			ow.WriteMultipleObjectsJSON(prettySet)
		case "custom":
			ow.WriteCustomOutput(outputFields)
		default:
			ow.WriteTable()
		}
	},
}
