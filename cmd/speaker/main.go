//
// timesig/cmd/timesig/main.go
//
// Copyright 2022 Naohiro CHIKAMATSU
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	tts "github.com/hegedustibor/htgo-tts"
	"github.com/jessevdk/go-flags"
	"golang.org/x/term"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type options struct {
	Mp3      string `short:"m" long:"mp3" default:"" description:"Convert text to mp3 file. You set output path"`
	Delete   bool   `short:"d" long:"delete" description:"Delete registerd time signal"`
	Lang     string `short:"l" long:"lang" default:"ja" description:"Setting speaking language"`
	Register string `short:"r" long:"register" default:"" description:"Register the time signal (e.g. --register=01:20)"`
	Version  bool   `short:"v" long:"version" description:"Show speaker command version"`
}

var osExit = os.Exit
var errTimeFormat = errors.New("time format is not correct")
var errCmdNotFound = errors.New(cmdName + " is not found $PATH and $GOPATH")

const cmdName string = "speaker"
const version string = "0.1.3"

// Time has hour and minute
type Time struct {
	hour string
	min  string
}

const (
	exitSuccess int = iota // 0
	exitFailure
)

func main() {
	var opts options
	args := parseArgs(&opts)

	if opts.Register != "" {
		register(args[0], opts)
	} else if opts.Delete {
		delete()
	} else if opts.Mp3 != "" {
		convMp3(args[0], opts)
	} else {
		speak(args[0], opts)
	}
	osExit(exitSuccess)
}

func register(text string, opts options) {
	if !isRoot() {
		fmt.Fprintln(os.Stderr, "you need root privileges.")
		showHelpForSudo(text, opts.Register)
		osExit(exitFailure)
	}

	if !existsCmd("cron") {
		die("If you use --register option, please install cron command")
	}

	time, err := cnvToTime(opts.Register)
	if err != nil {
		die(err.Error())
	}

	if err := registerCron(text, time, opts.Lang); err != nil {
		showHelpForSudo(text, opts.Register)
		fmt.Fprintln(os.Stderr, "")
		die("can not register time signal: " + err.Error())
	}
}

func delete() {
	if !isRoot() {
		die("you need root privileges.")
	}

	deleteTargets, err := getDeleteTargets()
	if err != nil {
		die("can not get delete target from crontab")
	}

	if len(deleteTargets) == 0 {
		die("you did not register time signal")
	}

	target, err := decideDeleteTargets(deleteTargets)
	if err != nil {
		die("can not get your input: " + err.Error())
	}

	if err := updateCronFile(target); err != nil {
		die("fail to update cron file: " + err.Error())
	}
}

func isSupportedLang(target string) bool {
	langs := []string{
		"en", "en-UK", "en-AU", "ja", "de", "es", "ru", "ar",
		"bn", "cs", "da", "nl", "fi", "el", "hi", "hu", "id",
		"km", "la", "it", "no", "pl", "sk", "sv", "th", "tr", "uk",
		"vi", "af", "bg", "ca", "cy", "et", "fr", "gu", "is", "jv",
		"kn", "ko", "lv", "ml", "mr", "ms", "ne", "pt", "ro", "si", "sr",
		"su", "ta", "te", "tl", "ur", "zh", "sw", "sq", "my", "mk",
		"hy", "hr", "eo", "bs",
	}
	return contains(langs, target)
}

func contains(list interface{}, elem interface{}) bool {
	rvList := reflect.ValueOf(list)

	if rvList.Kind() == reflect.Slice {
		for i := 0; i < rvList.Len(); i++ {
			item := rvList.Index(i).Interface()
			if !reflect.TypeOf(elem).ConvertibleTo(reflect.TypeOf(item)) {
				continue
			}
			target := reflect.ValueOf(elem).Convert(reflect.TypeOf(item)).Interface()
			if ok := reflect.DeepEqual(item, target); ok {
				return true
			}
		}
	}
	return false
}

func decideDeleteTargets(targets []string) (string, error) {
	if len(targets) == 1 {
		return targets[0], nil
	}

	for i, v := range targets {
		fmt.Printf("[%d] %v", i+1, v)
	}
	fmt.Println("")

	var response string
	fmt.Fprintf(os.Stdout, "Which time signal do you delete [1-%d]: ", len(targets))
	_, err := fmt.Scanln(&response)
	if err != nil {
		// If user input only enter.
		if strings.Contains(err.Error(), "expected newline") {
			return decideDeleteTargets(targets)
		}
		return "", err
	}

	no, err := strconv.Atoi(response)
	if err != nil || no < 1 || no > len(targets) {
		return decideDeleteTargets(targets)
	}
	return targets[no-1], nil
}

func getCronFilePath() string {
	if runtime.GOOS == "darwin" {
		return filepath.Join("/var/at/tabs", os.Getenv("SUDO_USER"))
	} else if runtime.GOOS == "linux" {
		return filepath.Join("/var/spool/cron/crontabs", os.Getenv("SUDO_USER"))
	}
	// BSD: /usr/lib/cron/tabs/USER
	// GNU: /var/spool/cron/crontab/USER
	return ""
}

func getDeleteTargets() ([]string, error) {
	strList, err := readFileToStrList(getCronFilePath())
	if err != nil {
		return []string{}, err
	}

	var deleteTargets []string
	for _, s := range strList {
		if strings.Contains(s, cmdName) {
			deleteTargets = append(deleteTargets, s)
		}
	}
	return deleteTargets, nil
}

func readFileToStrList(path string) ([]string, error) {
	var strList []string
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}
		if err == io.EOF && len(line) == 0 {
			break
		}
		strList = append(strList, line)
	}
	return strList, nil
}

func listToFile(filepath string, lines []string) error {
	fp, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer fp.Close()

	writer := bufio.NewWriter(fp)
	for _, line := range lines {
		writer.WriteString(line)
	}
	return writer.Flush()
}

func updateCronFile(target string) error {
	strList, err := readFileToStrList(getCronFilePath())
	if err != nil {
		return err
	}

	var newList []string
	for _, v := range strList {
		if v != target {
			newList = append(newList, v)
		}
	}
	return listToFile(getCronFilePath(), newList)
}

func convMp3(text string, opts options) {
	speech := tts.Speech{Folder: filepath.Dir(opts.Mp3), Language: opts.Lang}
	_, err := speech.CreateSpeechFile(text, strings.TrimRight(filepath.Base(opts.Mp3), ".mp3"))
	if err != nil {
		die("can not make audio file at " + opts.Mp3)
	}
	fmt.Println("Crate mp3 file at " + opts.Mp3)
}

func speak(text string, opts options) {
	mp3, err := textToMp3(text, opts.Lang)
	if err != nil {
		die("can not create audio file: " + err.Error())
	}
	defer os.Remove(mp3)

	if err := playMp3(mp3); err != nil {
		die("can not play audio file: " + err.Error())
	}
}

func showHelpForSudo(text, time string) {
	fmt.Fprintln(os.Stderr, "If you download "+cmdName+" command by go install, please execute as follows.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "$ sudo -E "+cmdName+" -r "+time+" \""+text+"\"")
}

func registerCron(text string, time Time, lang string) error {
	file, err := os.OpenFile(getCronFilePath(), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	cmdPath, err := speakerCmdPath()
	if err != nil {
		return err
	}

	schedule := time.min + " " + time.hour + " * * * " + cmdPath + " --lang=" + lang + " \"" + text + "\""
	fmt.Fprintln(file, schedule)

	if err := changeCronFileOwnership(); err != nil {
		return err
	}

	return nil
}

func changeCronFileOwnership() error {
	uid, err := lookupUID(os.Getenv("SUDO_USER"))
	if err != nil {
		return err
	}
	gid, err := lookupGID("crontab")

	if err := os.Chown(getCronFilePath(), uid, gid); err != nil {
		return err
	}
	return nil
}

func lookupGID(groupID string) (int, error) {
	group, err := user.LookupGroupId(groupID)
	if err != nil {
		group, err = user.LookupGroup(groupID)
		if err != nil {
			return 0, err
		}
	}

	gid, err := strconv.Atoi(group.Gid)
	if err != nil {
		return 0, err
	}
	return gid, nil
}

func lookupUID(userID string) (int, error) {
	u, err := user.LookupId(userID)
	if err != nil {
		u, err = user.Lookup(userID)
		if err != nil {
			return 0, err
		}
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return 0, err
	}
	return uid, nil
}

func speakerCmdPath() (string, error) {
	cmdPath, err := exec.LookPath(cmdName)
	if err != nil {
		gopath := os.Getenv("GOPATH")
		cmdPath = filepath.Join(gopath, "bin", cmdName)

		if !isFile(cmdPath) {
			return "", errCmdNotFound
		}
	}
	return cmdPath, nil
}

// isFile reports whether the path exists and is a file.
func isFile(path string) bool {
	stat, err := os.Stat(path)
	return (err == nil) && (!stat.IsDir())
}

func cnvToTime(time string) (Time, error) {
	timeList := strings.Split(time, ":")
	if len(timeList) != 2 {
		return Time{}, errTimeFormat
	}

	hour, err := strconv.Atoi(timeList[0])
	if err != nil {
		return Time{}, errTimeFormat
	}

	min, err := strconv.Atoi(timeList[1])
	if err != nil {
		return Time{}, errTimeFormat
	}

	if hour < 0 || hour > 24 {
		return Time{}, errTimeFormat
	}

	if (min < 0 || min > 60) || (hour == 24 && min != 0) {
		return Time{}, errTimeFormat
	}
	return Time{hour: timeList[0], min: timeList[1]}, nil
}

func textToMp3(text, lang string) (string, error) {
	dir := filepath.Join("/tmp")
	speech := tts.Speech{Folder: dir, Language: lang}
	checksum, err := md5sum(text)
	if err != nil {
		die("can not generate audio file name: " + err.Error())
	}

	mp3, err := speech.CreateSpeechFile(text, checksum)
	if err != nil {
		die("can not make audio file: " + err.Error())
	}
	return mp3, nil
}

func existsCmd(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func isRoot() bool {
	u, err := user.Current()
	return err == nil && u.Uid == "0"
}

// playmp3 play mp3 file.
func playMp3(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	st, format, err := mp3.Decode(f)
	if err != nil {
		return err
	}
	defer st.Close()

	if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
		return err
	}

	done := make(chan bool)
	speaker.Play(beep.Seq(st, beep.Callback(func() {
		done <- true
	})))
	<-done

	return nil
}

// md5sum generate md5sum (checksum) from text
func md5sum(text string) (string, error) {
	hash := md5.New()
	defer hash.Reset()
	if _, err := hash.Write([]byte(text)); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// die exit program with message.
func die(msg string) {
	fmt.Fprintln(os.Stderr, cmdName+": "+msg)
	osExit(exitFailure)
}

// parseArgs parse command line arguments.
// In this method, process exists if user specify version option, help option, and lack of arguments.
func parseArgs(opts *options) []string {
	p := newParser(opts)

	args, err := p.Parse()
	if err != nil {
		// If user specify --help option, help message already showed in p.Parse().
		// Moreover, help messages are stored inside err.
		if !strings.Contains(err.Error(), cmdName) {
			showHelp(p)
			showHelpFooter()
		} else {
			showHelpFooter()
		}
		osExit(exitFailure)
	}

	if hasPipeData() && len(args) == 0 {
		stdin, err := fromPIPE()
		if err != nil {
			die("can not get data from pipe: " + err.Error())
		}
		return []string{strings.ReplaceAll(stdin, "\n", "")}
	}

	if opts.Version {
		showVersion(cmdName, version)
		osExit(exitSuccess)
	}

	if opts.Register != "" && opts.Delete {
		die("can't be used --register option and --delete option at same time")
	}

	if !isSupportedLang(opts.Lang) {
		die(opts.Lang + " is not supported language")
	}

	if len(args) < 1 && !opts.Delete {
		showHelp(p)
		showHelpFooter()
		osExit(exitFailure)
	}

	return args
}

func hasPipeData() bool {
	return !term.IsTerminal(syscall.Stdin)
}

func fromPIPE() (string, error) {
	if hasPipeData() {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return "", nil
}

// showHelp print help messages.
func showHelp(p *flags.Parser) {
	p.WriteHelp(os.Stdout)
}

// showHelpFooter print author contact information and supported language.
func showHelpFooter() {
	fmt.Println("")
	fmt.Println("Supported language:")
	fmt.Println("  Please you see: https://github.com/nao1215/speaker/blob/main/doc/SupportedLanguage.md")
	fmt.Println("")
	fmt.Println("Contact:")
	fmt.Println("  If you find the bugs, please report the content of the error.")
	fmt.Println("  [GitHub Issue] https://github.com/nao1215/speaker/issues")
}

// newParser return initialized flags.Parser.
func newParser(opts *options) *flags.Parser {
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = cmdName
	parser.Usage = "[OPTIONS] MESSAGE"
	return parser
}

// showVersion show ubume command version information.
func showVersion(cmdName string, version string) {
	description := cmdName + " version " + version + " (under Apache License version 2.0)"
	fmt.Fprintln(os.Stdout, description)
}
