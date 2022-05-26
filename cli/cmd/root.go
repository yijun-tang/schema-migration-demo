/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed meta-info.json
var metaInfoStr string

type MetaInfo struct {
	Version string `json:"version"`
}

var (
	metaInfo        MetaInfo
	currentVersion  string
	rollbackVersion string
	targetVersion   string
	username        string
	password        string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cli",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Printf("Run: %s, %s\n", currentVersion, rollbackVersion)

		if rollbackVersion == "V-1" {
			if currentVersion >= targetVersion {
				fmt.Printf("Current Version: %s\n", currentVersion)
				return
			}

			files := listFiles("/home/tigergraph/yijun_ws/schema-migration-demo/database/schemas", "list scripts failed!!!")

			for _, f := range files {
				name := f.Name()
				version := name[:2]
				if version > targetVersion {
					break
				}
				if version > currentVersion {
					_, err := exec.Command("gsql", fmt.Sprintf("/home/tigergraph/yijun_ws/schema-migration-demo/database/schemas/%s", name)).Output()
					if err != nil {
						fmt.Printf("executing %s failed!!!", name)
						os.Exit(1)
					}
					currentVersion = version
					fmt.Printf("%s migrated...\n", name)
				}
			}

			saveVersion()
			fmt.Printf("Current Version: %s\n", currentVersion)
		} else {
			if rollbackVersion != currentVersion {
				fmt.Printf("Only supporting to rollback the latest version: %s\n", currentVersion)
				return
			}
			cookie := getCookie()
			client := http.Client{}
			var method string
			var url string
			var body io.Reader
			switch currentVersion {
			case "V1", "V2":
				method = http.MethodPut
				url = "http://localhost:14240/api/gsql-server/gsql/schema/change"
				if currentVersion == "V1" {
					body = strings.NewReader(`{"alterEdgeTypes":[],"alterVertexTypes":[],"dropEdgeTypes":[],"dropVertexTypes":["Comment","Post","Company","University","City","Country","Continent","Forum","Person","Tag","Tag_Class"],"addVertexTypes":[],"addEdgeTypes":[]}`)
				} else if currentVersion == "V2" {
					body = strings.NewReader(`{"alterEdgeTypes":[],"alterVertexTypes":[],"dropEdgeTypes":["Container_Of","Has_Creator","Has_Interest","Has_Member","Has_Moderator","Has_Tag","Has_Type","Is_Located_In","Is_Part_Of","Is_Subclass_Of","Knows","Likes","Reply_Of","Study_At","Work_At"],"dropVertexTypes":[],"addVertexTypes":[],"addEdgeTypes":[]}`)
				}
			case "V3":
				method = http.MethodDelete
				url = "http://localhost:14240/api/gsql-server/gsql/schema?graph=ldbc_snb"
				body = nil
			}

			req, _ := http.NewRequest(method, url, body)
			req.Header.Set("Cookie", "TigerGraphApp="+cookie)

			res, _ := client.Do(req)
			if _, err := io.ReadAll(res.Body); err != nil {
				fmt.Printf("[%s]rollback failed: %v\n", currentVersion, err)
				os.Exit(1)
			}
			fmt.Printf("Version %s has been revoked...\n", currentVersion)
			updateVersion()
			saveVersion()

			fmt.Printf("Current Version: %s\n", currentVersion)
		}

	},
}

func saveVersion() {
	metaInfo.Version = currentVersion
	f, err := os.Create("/home/tigergraph/yijun_ws/schema-migration-demo/cli/cmd/meta-info.json")
	if err != nil {
		fmt.Println("open meta-info.json file failed!!!")
		os.Exit(1)
	}
	defer f.Close()
	if m, err := json.Marshal(metaInfo); err == nil {
		_, e := f.Write(m)
		if e != nil {
			fmt.Println("write meta info failed!!!")
			os.Exit(1)
		}
	}
}

func updateVersion() {
	if i, err := strconv.Atoi(currentVersion[1:2]); err == nil {
		currentVersion = "V" + strconv.Itoa(i-1)
		return
	}
	fmt.Println("convert string to int failed!!!")
	os.Exit(1)
}

func listFiles(path, err string) []fs.FileInfo {
	files, e := ioutil.ReadDir(path)
	if e != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return files
}

func getCookie() string {
	client := http.Client{}
	cookieV := ""
	if req, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:14240/api/auth/login",
		strings.NewReader(fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)),
	); err == nil {
		if resp, err := client.Do(req); err == nil {
			for _, cookie := range resp.Cookies() {
				if cookie.Name == "TigerGraphApp" {
					cookieV = cookie.Value
					break
				}
			}
		}
	}

	return cookieV
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.Flags().StringVarP(&rollbackVersion, "rollbackVersion", "r", "V-1", "The rollback version, say V1, V2, ...")
	rootCmd.Flags().StringVarP(&targetVersion, "targetVersion", "g", "V-1", "The target version, say V1, V2, ...")
	rootCmd.Flags().StringVarP(&username, "username", "u", "tigergraph", "username")
	rootCmd.Flags().StringVarP(&password, "password", "p", "tigergraph", "password")

	// fmt.Printf("Execute(): %s, %s\n", currentVersion, rollbackVersion)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	if err := json.Unmarshal([]byte(metaInfoStr), &metaInfo); err != nil {
		fmt.Println("Parsing meta info failed!!!")
		os.Exit(1)
	}
	currentVersion = metaInfo.Version

	// fmt.Printf("init(): %s, %s\n", currentVersion, rollbackVersion)
}
