package main

import (
	"fmt"
	"github.com/clbanning/mxj"
	"launchpad.net/xmlpath"
	"os"
	"reflect"
	"flag"
	"path/filepath"
)

var myStack = &stack{}
var root *xmlpath.Node
var report *os.File
var matched = make([]string, 50)
var skipped = make([]string, 50)
var unmatched = make([]string, 50)

func main() {
	sourceFilePath := flag.String("source-file-path", "", "source file path")
	targetDirPath := flag.String("target-dir-path", "", "target dir path")
	verbose := flag.Bool("verbose", false, "verbose")

	flag.Parse()

	if *sourceFilePath == "" || *targetDirPath == "" {
		flag.PrintDefaults()
		os.Exit(2)
	}

	var err error
	report, err = os.Create("report.txt")
	if err != nil {
		panic(err)
	}
	defer report.Close()

	f, err := os.Open(*sourceFilePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	root, err = xmlpath.Parse(f)
	if err != nil {
		panic(err)
	}

	filepath.Walk(*targetDirPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Base(*sourceFilePath) == filepath.Base(path) {
			fmt.Printf("CHECKING... %s\n", path)
			fmt.Fprintf(report, "\n\nSOURCE : %s\n", *sourceFilePath)
			fmt.Fprintf(report, "TARGET : %s\n", path)

			f, err := os.Open(path)
			if err != nil {
				fmt.Println(err)
				fmt.Fprintln(report, err)
				return nil
			}
			defer f.Close()
			m, err := mxj.NewMapXmlReader(f)
			if err != nil {
				fmt.Println(err)
				fmt.Fprintln(report, err)
				return nil
			}

			matched = matched[:0]
			skipped = skipped[:0]
			unmatched = unmatched[:0]
			lookup_map(m)

			if len(unmatched) != 0 {
				fmt.Fprintf(report, "\n...UNMATCHED...\n")
				for _, s := range unmatched {
					fmt.Fprintf(report, "%s\n", s)
				}
			}
			if len(skipped) != 0 {
				fmt.Fprintf(report, "\n...skipped...\n")
				for _, s := range skipped {
					fmt.Fprintf(report, "%s\n", s)
				}
			}
			if *verbose && len(matched) != 0 {
				fmt.Fprintf(report, "\n...matched...\n")
				for _, s := range matched {
					fmt.Fprintf(report, "%s\n", s)
				}
			}
		}
		return nil
	})

	fmt.Println("DONE... report.txt")
}

func lookup_map(m map[string]interface{}) {
	for n, v := range m {
		myStack.push(n)
		switch v2 := v.(type) {
		case string:
			xpath := xmlpath.MustCompile(myStack.String())
			if xpath.Exists(root) {
				itr := xpath.Iter(root)
				isMatched := false
				var srcStr string
				for itr.Next() {
					node := itr.Node()
					srcStr = node.String()
					if v2 == srcStr {
						s := fmt.Sprintf("%s\t[%s]", myStack, v2)
						matched = append(matched, s)
						isMatched = true
						break
					}
				}
				if !isMatched {
					s := fmt.Sprintf("%s, source: [%s], target: [%s]", myStack, srcStr, v2)
					unmatched = append(unmatched, s)
				}
			} else {
				s := fmt.Sprintf("%s\t[%s]", myStack, v2)
				skipped = append(skipped, s)
			}
		case map[string]interface{}:
			lookup_map(v2)
		case []interface{}:
			lookup_slice(v2)
		default:
			panic(reflect.TypeOf(v))
		}
		myStack.pop()
	}
}

func lookup_slice(l []interface{}) {
	for _, v := range l {
		switch v2 := v.(type) {
		case string:
			panic(v2)
		case map[string]interface{}:
			lookup_map(v2)
		default:
			panic(reflect.TypeOf(v))
		}
	}
}

type stack struct {
	nodes []string
	count int
}

func (s *stack) push(n string) {
	s.nodes = append(s.nodes[:s.count], n)
	s.count++
}

func (s *stack) pop() string {
	if s.count == 0 {
		return ""
	}
	s.count--
	return s.nodes[s.count]
}

func (s *stack) String() (ret string) {
	for _, n := range s.nodes {
		ret += "/" + n
	}
	return
}
