package main

import (
  "strings"
  "os"
  "strconv"
  "github.com/codegangsta/cli"
  "github.com/JFrogDev/artifactory-cli-go/commands"
  "github.com/JFrogDev/artifactory-cli-go/utils"
)

var dryRun bool
var url string
var username string
var password string
var props string
var recursive bool
var flat bool
var useRegExp bool
var threads int
var minSplitSize int64
var splitCount int
var confFile string

func main() {
    defer utils.RemoveTempDir()

    app := cli.NewApp()
    app.Name = "art"
    app.Usage = "See https://github.com/JFrogDev/artifactory-cli-go for usage instructions."

    app.Commands = []cli.Command{
        {
            Name: "config",
            Flags: getFlags(),
            Aliases: []string{"c"},
            Usage: "config",
            Action: func(c *cli.Context) {
                config(c)
            },
        },
        {
            Name: "upload",
            Flags: getUploadFlags(),
            Aliases: []string{"u"},
            Usage: "upload <local path> <repo path>",
            Action: func(c *cli.Context) {
                upload(c)
            },
        },
        {
            Name: "download",
            Flags: getDownloadFlags(),
            Aliases: []string{"d"},
            Usage: "download <repo path>",
            Action: func(c *cli.Context) {
                download(c)
            },
        },
    }

    app.Run(os.Args)
}

func getFlags() []cli.Flag {
    return []cli.Flag{
        cli.StringFlag{
         Name:  "url",
         Usage: "[Mandatory] Artifactory URL",
        },
        cli.StringFlag{
         Name:  "user",
         Usage: "[Optional] Artifactory username",
        },
        cli.StringFlag{
         Name:  "password",
         Usage: "[Optional] Artifactory password",
        },
    }
}

func getUploadFlags() []cli.Flag {
    flags := []cli.Flag{
        nil,nil,nil,nil,nil,nil,nil,nil,nil,
    }
    copy(flags[0:3], getFlags())
    flags[3] = cli.StringFlag{
         Name:  "props",
         Usage: "[Optional] List of properties in the form of key1=value1;key2=value2,... to be attached to the uploaded artifacts.",
    }
    flags[4] = cli.StringFlag{
        Name:  "recursive",
        Value:  "",
        Usage: "[Default: true] Set to false if you do not wish to collect artifacts in sub-folders to be uploaded to Artifactory.",
    }
    flags[5] = cli.StringFlag{
        Name:  "flat",
        Value:  "",
        Usage: "[Default: true] If not set to true, and the upload path ends with a slash, files are uploaded according to their file system hierarchy.",
    }
    flags[6] = cli.BoolFlag{
         Name:  "regexp",
         Usage: "[Default: false] Set to true to use a regular expression instead of wildcards expression to collect files to upload.",
    }
    flags[7] = cli.StringFlag{
         Name:  "threads",
         Value:  "",
         Usage: "[Default: 3] Number of artifacts to upload in parallel.",
    }
    flags[8] = cli.BoolFlag{
         Name:  "dry-run",
         Usage: "[Default: false] Set to true to disable communication with Artifactory.",
    }
    return flags
}

func getDownloadFlags() []cli.Flag {
    flags := []cli.Flag{
        nil,nil,nil,nil,nil,nil,nil,nil,nil,
    }
    copy(flags[0:3], getFlags())
    flags[3] = cli.StringFlag{
         Name:  "props",
         Usage: "[Optional] List of properties in the form of key1=value1;key2=value2,... Only artifacts with these properties will be downloaded.",
    }
    flags[4] = cli.StringFlag{
        Name:  "recursive",
        Value:  "",
        Usage: "[Default: true] Set to false if you do not wish to include the download of artifacts inside sub-folders in Artifactory.",
    }
    flags[5] = cli.StringFlag{
        Name:  "flat",
        Value:  "",
        Usage: "[Default: false] Set to true if you do not wish to have the Artifactory repository path structure created locally for your downloaded files.",
    }
    flags[6] = cli.StringFlag{
        Name:  "min-split",
        Value:  "",
        Usage: "[Default: 5120] Minimum file size in KB to split into ranges when downloading. Set to -1 for no splits.",
    }
    flags[7] = cli.StringFlag{
        Name:  "split-count",
        Value:  "",
        Usage: "[Default: 3] Number of parts to split a file when downloading. Set to 0 for no splits.",
    }
    flags[8] = cli.StringFlag{
         Name:  "threads",
         Value:  "",
         Usage: "[Default: 3] Number of artifacts to download in parallel.",
    }
    return flags
}

func initFlags(c *cli.Context, cmd string) {
    url = getMandatoryFlag(c, "url")
    if !strings.HasSuffix(url, "/") {
        url += "/"
    }

    strFlat := c.String("flat")
    if cmd == "upload" {
        if strFlat == "" {
            flat = true
        }
    } else
    if cmd == "download" {
        if strFlat == "" {
            flat = false
        }
    }

    username = c.String("user")
    password = c.String("password")
    props = c.String("props")
    dryRun = c.Bool("dry-run")
    useRegExp = c.Bool("regexp")
    var err error
    if c.String("threads") == "" {
        threads = 3
    } else {
        threads, err = strconv.Atoi(c.String("threads"))
        if err != nil || threads < 1 || threads > 30 {
            utils.Exit("The '--threads' option should have a numeric value between 1 and 30. Try 'art download --help'.")
        }
    }
    if c.String("min-split") == "" {
        minSplitSize = 5120
    } else {
        minSplitSize, err = strconv.ParseInt(c.String("min-split"), 10, 64)
        if err != nil {
            utils.Exit("The '--min-split' option should have a numeric value. Try 'art download --help'.")
        }
    }
    if c.String("split-count") == "" {
        splitCount = 3
    } else {
        splitCount, err = strconv.Atoi(c.String("split-count"))
        if err != nil {
            utils.Exit("The '--split-count' option should have a numeric value. Try 'art download --help'.")
        }
        if splitCount > 15 {
            utils.Exit("The '--split-count' option value is limitted to a maximum of 15.")
        }
        if splitCount < 0 {
            utils.Exit("The '--split-count' option cannot have a negative value.")
        }
    }

    if c.String("recursive") == "" {
        recursive = true
    } else {
        recursive = c.Bool("recursive")
    }
}

func config(c *cli.Context) {
    commands.Config()
}

func download(c *cli.Context) {
    initFlags(c, "download")
    if len(c.Args()) != 1 {
        utils.Exit("Wrong number of arguments. Try 'art download --help'.")
    }
    pattern := c.Args()[0]
    commands.Download(url, pattern, props, username, password, recursive, flat, dryRun, minSplitSize, splitCount, threads)
}

func upload(c *cli.Context) {
    initFlags(c, "upload")
    size := len(c.Args())
    if size != 2 {
        utils.Exit("Wrong number of arguments. Try 'art upload --help'.")
    }
    localPath := c.Args()[0]
    targetPath := c.Args()[1]
    commands.Upload(url, localPath, targetPath, recursive, flat, props, username, password, threads, useRegExp, dryRun)
}

// Get a CLI flagg. If the flag does not exist, utils.Exit with a message.
func getMandatoryFlag(c *cli.Context, flag string) string {
    value := c.String(flag)
    if value == "" {
        utils.Exit("The --" + flag + " flag is mandatory")
    }
    return value
}