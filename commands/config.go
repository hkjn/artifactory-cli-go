package commands

import (
    "os"
    "fmt"
    "bytes"
    "syscall"
    "io/ioutil"
    "encoding/json"
    "github.com/JFrogDev/artifactory-cli-go/utils"
    "github.com/JFrogDev/artifactory-cli-go/Godeps/_workspace/src/golang.org/x/crypto/ssh/terminal"
)

func Config(details *utils.ArtifactoryDetails, interactive, shouldEncPassword bool) {
    var bytePassword []byte
    if interactive {
        if details.Url == "" {
            print("Artifactory Url: ")
            fmt.Scanln(&details.Url)
        }
        if details.User == "" {
            print("User: ")
            fmt.Scanln(&details.User)
        }
        if details.Password == "" {
            print("Password: ")
            var err error
            bytePassword, err = terminal.ReadPassword(int(syscall.Stdin))
            details.Password = string(bytePassword)
            utils.CheckError(err)
        }
    }
    details.Url = utils.AddTrailingSlashIfNeeded(details.Url)
    if shouldEncPassword {
        details = encryptPassword(details)
    }
    writeConfFile(details)
}

func ShowConfig() {
    details := readConfFile()
    if details.Url != "" {
        fmt.Println("Url: " + details.Url)
    }
    if details.User != "" {
        fmt.Println("User: " + details.User)
    }
    if details.Password != "" {
        fmt.Println("Password: " + details.Password)
    }
}

func ClearConfig() {
    writeConfFile(new(utils.ArtifactoryDetails))
}

func GetConfig() *utils.ArtifactoryDetails {
    return readConfFile()
}

func encryptPassword(details *utils.ArtifactoryDetails) *utils.ArtifactoryDetails {
    if details.Password == "" {
        return details
    }
    response, encPassword := utils.GetEncryptedPasswordFromArtifactory(details)
    switch response.StatusCode {
        case 409:
            utils.Exit(utils.ExitCodeError, "\nYour Artifactory server is not configured to encrypt passwords.\n" +
                "You may use \"art config --enc-password=false\"")
        case 200:
            details.Password = encPassword
        default:
            utils.Exit(utils.ExitCodeError, "\nArtifactory response: " + response.Status)
    }
    return details
}

func getConFilePath() string {
    userDir := utils.GetHomeDir()
    if userDir == "" {
        utils.Exit(utils.ExitCodeError, "Couldn't find home directory. Make sure your HOME environment variable is set.")
    }
    confPath := userDir + "/.jfrog/"
    os.MkdirAll(confPath ,0777)
    return confPath + "art-cli.conf"
}

func writeConfFile(details *utils.ArtifactoryDetails) {
    confFilePath := getConFilePath()
    if !utils.IsFileExists(confFilePath) {
        out, err := os.Create(confFilePath)
        utils.CheckError(err)
        defer out.Close()
    }

    b, err := json.Marshal(&details)
    utils.CheckError(err)
    var content bytes.Buffer
    err = json.Indent(&content, b, "", "  ")
    utils.CheckError(err)

    ioutil.WriteFile(confFilePath,[]byte(content.String()), 0x777)
}

func readConfFile() *utils.ArtifactoryDetails {
    confFilePath := getConFilePath()
    details := new(utils.ArtifactoryDetails)
    if !utils.IsFileExists(confFilePath) {
        return details
    }
    content := utils.ReadFile(confFilePath)
    json.Unmarshal(content, &details)

    return details
}