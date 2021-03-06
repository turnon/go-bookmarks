package gobookmarks

import (
	"bufio"
	"encoding/json"
	"os"
	"regexp"
	"strconv"
	"time"
)

type Bookmark struct {
	dirStack []string
	Items    []Item
}

type Item struct {
	Href    string   `json:"href"`
	AddDate string   `json:"add_date"`
	Title   string   `json:"title"`
	Dirs    []string `json:"dirs"`
}

var (
	dirRe   = regexp.MustCompile(`^\s*<DT><H3 .*>(.*)</H3>$`)
	itemRe  = regexp.MustCompile(`^\s*<DT><A HREF="(.*)" ADD_DATE="([0-9]+)".*>(.*)</A>$`)
	listEnd = regexp.MustCompile(`^\s*</DL><p>$`)
)

func Read(path string) Bookmark {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var bm Bookmark
	bm.parse(file)
	return bm
}

func (bm *Bookmark) JSON() ([]byte, error) {
	return json.Marshal(&bm)
}

func (bm *Bookmark) parse(file *os.File) {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()

		if bm.matchItem(line) {
			continue
		}

		matchDir := dirRe.FindStringSubmatch(line)
		if len(matchDir) > 0 {
			bm.dirStack = append(bm.dirStack, matchDir[1])
			continue
		}

		if listEnd.MatchString(line) && len(bm.dirStack) > 1 {
			bm.dirStack = bm.dirStack[:len(bm.dirStack)-1]
			continue
		}
	}
}

func (bm *Bookmark) matchItem(line string) bool {
	m := itemRe.FindStringSubmatch(line)
	if len(m) == 0 {
		return false
	}

	seconds, err := strconv.ParseInt(m[2], 0, 64)
	if err != nil {
		panic(err)
	}
	date := time.Unix(seconds, 0)

	dirs := make([]string, len(bm.dirStack))
	copy(dirs, bm.dirStack)

	i := Item{m[1], date.Format(time.RFC3339), m[3], dirs}
	bm.Items = append(bm.Items, i)

	return true
}
