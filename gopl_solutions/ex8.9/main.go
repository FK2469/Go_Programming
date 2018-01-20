// ex8.9: Write a version of du that computes and periodically 
// displays separate totals for each of the root directories.

package main

import(
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"text/tabwriter"
	"time"
)

var verbose = flag.Bool("v",false, "Show verbose progress messages.")

type directoryInformation struct{
	name string
	size int64
}

func main(){
	flag.Parse()
	roots := flag.Args()

	if len(roots) == 0{
		roots = []string{"."}
	}

	fileSizes := make(chan *directoryInformation)
	var waitGroup sync.WaitGroup
	directories := make(map[string]*directoryInformation)

	for _, root := range roots{
		directories[root] = &directoryInformation{name : root}
		waitGroup.Add(1)
		go walkDir(root, root, &waitGroup, fileSizes)
	}
	go func(){
		waitGroup.Wait()
		close(fileSizes)
	}()

	var tick <-chan time.Time

	if *verbose{
		tick = time.Tick(500 * time.Millisecond)
	}
	tablewriter := new(tabwriter.Writer).Init(os.Stdout,0,8,2,' ',0)
	printDirectoryNames(roots, tablewriter)

loop:
	for{
		select{
		case intermediateDirectoryInformation, ok := <- fileSizes:
			if !ok {
				break loop
			}
			directories[intermediateDirectoryInformation.name].size += intermediateDirectoryInformation.size
		case <- tick:
			printDiskUsage(roots, directories, tablewriter)
		}
	}
	printDiskUsage(roots, directories, tablewriter)
	fmt.Printf("\n")
}

func printDirectoryNames(roots []string, tablewriter *tabwriter.Writer){
	for _, root := range roots{
		fmt.Fprintf(tablewriter,"%v\t",root)
	}
	fmt.Fprintf(tablewriter,"\n")

	for i := 0; i<len(roots); i++{
		fmt.Fprintf(tablewriter,"%v\t","------")
	}
	fmt.Fprintf(tablewriter,"\n")
	tablewriter.Flush()
}

func printDiskUsage(roots []string, directories map[string]*directoryInformation, tablewriter *tabwriter.Writer){
	fmt.Printf("\r")
	for _, root := range roots{
		fmt.Fprintf(tablewriter, "%v\t", fmt.Sprintf("%.1f GB", float64(directories[root].size)/1e9))
	}
	tablewriter.Flush()
}

func walkDir(dir string, root string, waitGroup *sync.WaitGroup, fileSizes chan<- *directoryInformation){
	defer waitGroup.Done()
	for _, entry := range direntries(dir){
		if entry.IsDir(){
			subdir := filepath.Join(dir, entry.Name())
			waitGroup.Add(1)
			go walkDir(subdir, root, waitGroup, fileSizes)
		}else{
			fileSizes <- &directoryInformation{name:root, size: entry.Size()}
		}
	}
}

var sema = make(chan struct{}, 32)

func direntries(dir string) []os.FileInfo{
	sema <- struct{}{}
	defer func() {<-sema }()
	entires, err := ioutil.ReadDir(dir)
	if err != nil{
		fmt.Fprintf(os.Stderr, "du4: %v\n",err)
		return nil
	}
	return entires
}
