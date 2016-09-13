package io

import (
	"github.com/biogo/hts/bam"
	"github.com/biogo/hts/sam"
	"os"
)

func ReadBam(file string, cpus int) <-chan *sam.Record {
	var fi *os.File
	var err error
	var br *bam.Reader

	reads := make(chan *sam.Record, 1000)

	if file == "stdin" || file == "-" {
		fi = os.Stdin
	} else {
		fi, err = os.Open(file)
	}

	if err != nil {
		ExitWithMessage(err)
	}

	br, err = bam.NewReader(fi, cpus)

	if err != nil {
		ExitWithMessage(err)
	}

	go func() {
		for {
			if sr, err3 := br.Read(); err3 != nil {
				break
			} else {
				reads <- sr
			}
		}
		br.Close()
		close(reads)
	}()

	return reads
}
