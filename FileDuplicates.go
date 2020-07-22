package main

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"
	"sync"
	"github.com/iafan/cwalk"
)

var lock = sync.RWMutex{}

type fileInfo struct {
	name       string
	path       string
	size       int64
	isDir      bool
	sha256hash string
	fileType   string
	location   string
}

func getSHA256Hash(path string) string {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		fmt.Printf("Error '%s' opening file at '%s'\n", err, path)
	}

	data, _ := ioutil.ReadAll(file)
	sha256Sum := sha256.Sum256(data)
	sha256String := hex.EncodeToString(sha256Sum[:])

	data = nil

	return sha256String
}

func contains(sliceToSearch []string, searchString string) bool {
	for _, val := range sliceToSearch {
		match, _ := regexp.MatchString(val, searchString)
		if match {
			return true
		}
	}

	return false
}

func walkTheDirectory(directoryToSearch string, exclusionList []string) (map[string][]fileInfo, []fileInfo) {
	allFiles := make(map[string][]fileInfo)
	var allDirectories []fileInfo

	var walkDirPath = func(pathX string, infoX os.FileInfo, errX error) error {
		lock.Lock()
		defer lock.Unlock()
		fileHash := ""

		if errX != nil {
			fmt.Printf("\n Error '%v' at path '%q'\n", errX, pathX)
			return errX
		}

		if contains(exclusionList, pathX) {
			return nil
		}

		fi := fileInfo{
			name:  infoX.Name(),
			path:  pathX,
			size:  infoX.Size(),
			isDir: infoX.IsDir(),
		}

		if !(fi.isDir) {
			fileLoc := fmt.Sprintf("%s\\%s", directoryToSearch, pathX)
			fileHash := getSHA256Hash(fileLoc)

			// fi.sha256hash = fileHash
			fi.fileType = "File"

			// hashSlice = append(hashSlice, fileHash)
			allFiles[fileHash] = append(allFiles[fileHash], fi)
		} else {
			fi.fileType = "Directory"
			fi.sha256hash = fileHash
			allDirectories = append(allDirectories, fi)
		}

		return nil
	}

	cwalk.Walk(directoryToSearch, walkDirPath)

	return allFiles, allDirectories
}

func writeCSVFile(files []fileInfo, csvSaveLocation string) {
	fmt.Printf("Writing data to CSV file '%s'\n", csvSaveLocation)

	csvData := [][]string{
		{"FileName", "RelativePath", "FileSize", "Type", "FileHash", "Location"},
	}

	csvFile, _ := os.Create(csvSaveLocation)
	csvFileWriter := csv.NewWriter(csvFile)

	for _, file := range files {
		filedata := []string{file.name, file.path, strconv.FormatInt(file.size, 10), file.fileType, file.sha256hash, file.location}
		csvData = append(csvData, filedata)
	}

	csvFileWriter.WriteAll(csvData)
}

func countFilesAndDirectories(files []fileInfo, folder string) {
	directoryCount := 0
	fileCount := 0

	for _, val := range files {
		if val.isDir {
			directoryCount++
		} else {
			fileCount++
		}
	}

	fmt.Printf("\nNumber of Directories found in '%s': %d", folder, directoryCount)
	fmt.Printf("\nNumber of Files found in '%s':       %d\n\n", folder, fileCount)
}

func isInSlice(referenceSlice []string, differenceString string) bool {
	sort.Strings(referenceSlice)
	i := sort.SearchStrings(referenceSlice, differenceString)
	if i < len(referenceSlice) && referenceSlice[i] == differenceString {
		return true
	}

	return false
}

func getDifferences(references, differences []string) map[string]string {
	var allDifferences = make(map[string]string)

	for _, diff := range differences {
		if !(isInSlice(references, diff)) {
			allDifferences[diff] = "<="
		}
	}

	for _, ref := range references {
		if !(isInSlice(differences, ref)) {
			allDifferences[ref] = "=>"
		}
	}

	return allDifferences
}

func assignComparison(diffList map[string]string, refFiles *[]fileInfo, diffFiles *[]fileInfo) {
	for key, value := range diffList {
		for _, ref := range *refFiles {
			if key == ref.sha256hash {
				ref.location = value
			}
		}
		for _, diff := range *diffFiles {
			if key == diff.sha256hash {
				diff.location = value
			}
		}
	}
}

func main() {
	startTime := time.Now()

	fmt.Printf("Start time: %s\n", startTime)

	var referenceFolder string
	// var comparisonFolder string
	// var csvSaveLocation string
	// var comparisonCsvSaveLocation string

	flag.StringVar(&referenceFolder, "ReferenceFolder", "C:\\Users\\gouda", "Specify a path to search")
	// flag.StringVar(&comparisonFolder, "ComparisonFolder", `"C:\Testing"`, "Specify a path to comapare against")
	// flag.StringVar(&csvSaveLocation, "CSVSaveLocation", `"C:\Temp\GoTest.csv"`, "Specify a path to save the output CSV")
	// flag.StringVar(&comparisonCsvSaveLocation, "ComparisonCSVSaveLocaiton", `"C:\Temp\GoTestComparison.csv"`, "Specify a path to save the output CSV")

	flag.Parse()

	exclusionList := []string{}

	fmt.Printf("Gathering directory and file data from reference folder '%s'...\n", referenceFolder)
	referenceFiles, _ := walkTheDirectory(referenceFolder, exclusionList)

	duplicates := 0

	for key := range referenceFiles {
		if len(referenceFiles[key]) > 1 {
			duplicates++
			fmt.Println(referenceFiles[key])
		}
	}

	fmt.Printf("\nTotal number of duplicate SHA256 hashes: %d\n", duplicates)

	elapsedTime := time.Since(startTime)

	fmt.Printf("\nElapsed time: %s\n\n", elapsedTime)

	fmt.Println("Program Complete")
}
