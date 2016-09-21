package cmd

import (
	"errors"
	"fmt"
	"github.com/fredericlemoine/bam2introns/io"
	"github.com/spf13/cobra"
	"os"
	"sync"
)

var Version string = "Unknown"

var stranded string
var cpus int
var grouped bool
var qualityFilter int

type indata struct {
	infile string
	index  int
}
type outdata struct {
	intron    *io.Intron
	fileindex int
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var RootCmd = &cobra.Command{
	Use:   "bam2introns <in1.bam> [in2.bam...]",
	Short: "Write the list of intron coordinates from a bam file",
	Long: `Write the list of intron coordinates from a bam file in bed format

If the bam file is not oriented (-s none):
- The strand is the read mapping strand

If the bam file is oriented (-s stranded):
- The strand is the strand of the read mapping if the read is the first read
- The strand is the opposite of the strand of the read mapping if the read is the mate

If the bam file is oriented (-s reverse):
- The strand is the opposite of the strand of the read mapping if the read is the first read
- The strand is the strand of the read mapping if the read is the mate read

Does not take into account secondary alignments (flag 0x100)
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			io.ExitWithMessage(errors.New("No input file given"))
		}

		intronChan := make(chan outdata)
		fileChannel := make(chan indata)
		buffer := make(map[string]*io.Intron)

		// We put files and their index in a channel
		go func() {
			for i, infile := range args {
				fileChannel <- indata{infile, i}
			}
			close(fileChannel)
		}()

		totalFiles := len(args)
		// If > 2 threads per file then we can
		// use them to read bam more quickly
		readThread := max(cpus/totalFiles, 1)

		// We init the thread pool
		var wg sync.WaitGroup
		for cpu := 0; cpu < min(cpus, totalFiles); cpu++ {
			wg.Add(1)
			go func() {
				// We take a file in the channel
				for f := range fileChannel {
					// We init bam reader
					reads := io.ReadBam(f.infile, readThread, qualityFilter, true)
					s := io.Strand(stranded)
					// We take reads from the reader
					for r := range reads {
						// We detect introns in this read
						for _, intron := range io.Introns(r, s) {
							// We give this intron to the channel
							intronChan <- outdata{intron, f.index}
						}
					}
				}
				wg.Done()
			}()
		}

		// We wait for all the threads to end and then close the intron channel
		go func() {
			wg.Wait()
			close(intronChan)
		}()

		// We take detected introns from the channel
		for intron := range intronChan {
			if !grouped {
				// If not grouped, we print it directly
				fmt.Fprintf(os.Stdout, "%s\n", io.PrintIntrons(intron.intron))
			} else {
				// Otherwise, we buffer it and count the occurences
				st := '+'
				if !intron.intron.Strand {
					st = '-'
				}
				key := fmt.Sprintf("%s:%d-%d[%c]", intron.intron.Chr, intron.intron.Start, intron.intron.End, st)
				if i, ok := buffer[key]; !ok {
					intron.intron.Count = make([]int, totalFiles)
					intron.intron.Count[intron.fileindex] = 1
					buffer[key] = intron.intron
				} else {
					i.Count[intron.fileindex]++
				}
			}
		}

		// Finally, if grouped, we print the occurences of all introns
		if grouped {
			for _, intron := range buffer {
				fmt.Fprintf(os.Stdout, "%s\n", io.PrintIntrons(intron))
			}
		}
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().IntVarP(&cpus, "threads", "t", 1, "Number of decompressing threads")
	RootCmd.PersistentFlags().StringVarP(&stranded, "stranded", "s", "none", "Stranded : none, stranded or reverse")
	RootCmd.PersistentFlags().BoolVarP(&grouped, "grouped", "g", false, "Grouped : group introns by positions, and count them")
	RootCmd.PersistentFlags().IntVarP(&qualityFilter, "quality-filter", "q", 255, "Filters reads with map qual < cutoff")

}
