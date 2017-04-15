package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/user"
	"sort"
	"strconv"
	"time"
	"strings"

	"github.com/dnote-io/cli/utils"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Book   string
	APIKey string
}

// Deprecated. See upgrade/migrate.go
type YAMLDnote map[string][]string

type Dnote map[string]Book
type Book []Note
type Note struct {
	UID     string
	Name 	string
	Content string
	AddedOn int64
}

const configFilename = ".dnoterc"
const DnoteUpdateFilename = ".dnote-upgrade"
const dnoteFilename = ".dnote"
const Version = "0.1.0-beta.1"

const letterRunes = "abcdefghipqrstuvwxyz0123456789"
const interfix = "_note_"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenerateNoteID() string {
	result := make([]byte, 8)
	for i := range result {
		result[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(result)
}

// Get the amount of note for auto-generated note names.
func GetAutoGeneratedNoteNumber(note_name_list []string) (int, error) {
	var highest_number int

	book, err := GetCurrentBook()
	if err != nil {
		return highest_number, err
	}

	var note_numbers []int
	for _, note_name := range note_name_list {
		trimmed_string := strings.TrimPrefix(note_name, book + interfix)

		if trimmed_string != "" {
			note_number, err := strconv.Atoi(trimmed_string)
			if err != nil {
				return highest_number, err
			}

			note_numbers = append(note_numbers, note_number)
		}
	}

	if note_numbers != nil {
		biggest := note_numbers[0]
		for _, v := range note_numbers {
			if v > biggest {
				biggest = v
			}
		}

		highest_number = biggest
	}

	return highest_number, nil
}

// This function generate note names for unnamed notes.
// The note name is as such : (book)_note_(current number of note + 1)
// E.g, foo_note_001, foo_note_002, foo_note_003
// The name consists of 3 parts :
// 1. book is the name of the book.
// 2. _note_ is a constant substring.
// 3. the number of note is the current number of note + 1.
func GenerateNoteName() (string, error) {
	constant_middle := "_note_"
	result := ""

	book, err := GetCurrentBook()
	if err != nil {
		return result, err
	}

	json_data, err := GetDnote()
	if err != nil {
		return result, err
	}

	var note_names []string
	for _, note := range json_data[book] {
		if strings.Contains(note.Name, constant_middle) {
			note_names = append(note_names, note.Name)
		}
	}

	if note_names != nil {
			biggest, err := GetAutoGeneratedNoteNumber(note_names)
			if err != nil {
				return result, err
			}

			if biggest != 0 {
				desired_number := biggest + 1
				final_number := strconv.Itoa(desired_number)
				result = book + constant_middle + final_number
			}else{
				result = book + constant_middle + "0"
			}
	}else{
		result = book + constant_middle + "0"
	}

	return result, nil
}

func GetConfigPath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", usr.HomeDir, configFilename), nil
}

func GetDnotePath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", usr.HomeDir, dnoteFilename), nil
}

func GetYAMLDnoteArchivePath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", usr.HomeDir, ".dnote-yaml-archived"), nil
}

func GenerateConfigFile() error {
	content := []byte("book: general\n")
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(configPath, content, 0644)
	return err
}

func TouchDnoteFile() error {
	dnotePath, err := GetDnotePath()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dnotePath, []byte{}, 0644)
	return err
}

func TouchDnoteUpgradeFile() error {
	dnoteUpdatePath, err := GetDnoteUpdatePath()
	if err != nil {
		return err
	}

	epoch := strconv.FormatInt(time.Now().Unix(), 10)
	content := []byte(fmt.Sprintf("LAST_UPGRADE_EPOCH: %s\n", epoch))

	err = ioutil.WriteFile(dnoteUpdatePath, content, 0644)
	return err
}

func GetDnoteUpdatePath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", usr.HomeDir, DnoteUpdateFilename), nil
}

func AskConfirmation(question string) (bool, error) {
	fmt.Printf("%s [Y/n]: ", question)

	reader := bufio.NewReader(os.Stdin)
	res, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	ok := res == "y\n" || res == "Y\n" || res == "\n"

	return ok, nil
}

// ReadNoteContent reads the content of dnote
func ReadNoteContent() ([]byte, error) {
	notePath, err := GetDnotePath()
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadFile(notePath)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GetDnote reads and parses the dnote
func GetDnote() (Dnote, error) {
	ret := Dnote{}

	b, err := ReadNoteContent()
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(b, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

// WriteDnote persists the state of Dnote into the dnote file
func WriteDnote(dnote Dnote) error {
	d, err := json.MarshalIndent(dnote, "", "  ")
	if err != nil {
		return err
	}

	notePath, err := utils.GetDnotePath()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(notePath, d, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Deprecated. See upgrade/upgrade.go
func GetNote() (YAMLDnote, error) {
	ret := YAMLDnote{}

	b, err := ReadNoteContent()
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(b, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func WriteConfig(config Config) error {
	d, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(configPath, d, 0644)
	if err != nil {
		return err
	}

	return nil
}

func ReadConfig() (Config, error) {
	var ret Config

	configPath, err := GetConfigPath()
	if err != nil {
		return ret, err
	}

	b, err := ioutil.ReadFile(configPath)
	if err != nil {
		return ret, err
	}

	err = yaml.Unmarshal(b, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func GetCurrentBook() (string, error) {
	config, err := ReadConfig()
	if err != nil {
		return "", err
	}

	return config.Book, nil
}

func GetBooks() ([]string, error) {
	dnote, err := GetDnote()
	if err != nil {
		return nil, err
	}

	books := make([]string, 0, len(dnote))
	for k := range dnote {
		books = append(books, k)
	}

	sort.Strings(books)

	return books, nil
}
