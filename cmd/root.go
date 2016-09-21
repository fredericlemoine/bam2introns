package cmd

import (
	"errors"
	"fmt"
	"github.com/biogo/hts/sam"
	"github.com/fredericlemoine/bam2introns/io"
	"github.com/spf13/cobra"
	"os"
)

var Version string = "Unknown"

var stranded string
var cpus int
var grouped bool

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
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			io.ExitWithMessage(errors.New("No input file given"))
		}
		buffer := make(map[string]*io.Intron)
		for i, infile := range args {
			fmt.Fprintf(os.Stdout, "File: %s\n", infile)
			reads := io.ReadBam(infile, cpus)
			s := io.Strand(stranded)
			if grouped {
				groupIntrons(buffer, reads, s, len(args), i)
			} else {
				readIntrons(reads, s)
			}
		}
		if grouped {
			for _, intron := range buffer {
				fmt.Fprintf(os.Stdout, "%s\n", io.PrintIntrons(intron))
			}
		}
	},
}

func readIntrons(reads <-chan *sam.Record, s io.Stranded) {
	for r := range reads {
		for _, intron := range io.Introns(r, s) {
			fmt.Fprintf(os.Stdout, "%s\n", io.PrintIntrons(intron))
		}
	}
}

func groupIntrons(buffer map[string]*io.Intron, reads <-chan *sam.Record, s io.Stranded, totalFiles int, currentFile int) {
	for r := range reads {
		for _, intron := range io.Introns(r, s) {
			st := '+'
			if !intron.Strand {
				st = '-'
			}
			key := fmt.Sprintf("%s:%d-%d[%c]", intron.Chr, intron.Start, intron.End, st)
			if i, ok := buffer[key]; !ok {
				intron.Count = make([]int, totalFiles)
				intron.Count[currentFile] = 1
				buffer[key] = intron
			} else {
				i.Count[currentFile]++
			}
		}
	}
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
}
